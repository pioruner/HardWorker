package updater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Store struct {
	client        *s3.Client
	bucket        string
	prefix        string
	publicBaseURL string
	manifestKey   string
}

func newS3Store(ctx context.Context, cfg *S3Config) (*s3Store, error) {
	if cfg == nil {
		return nil, fmt.Errorf("s3 config is nil")
	}
	if strings.TrimSpace(cfg.Endpoint) == "" || strings.TrimSpace(cfg.Bucket) == "" {
		return nil, fmt.Errorf("s3 endpoint and bucket are required")
	}

	parsed, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse s3 endpoint: %w", err)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
		o.BaseEndpoint = &cfg.Endpoint
		o.EndpointOptions.DisableHTTPS = parsed.Scheme == "http"
	})

	return &s3Store{
		client:        client,
		bucket:        cfg.Bucket,
		prefix:        strings.Trim(cfg.Prefix, "/"),
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
		manifestKey:   cfg.ManifestKey,
	}, nil
}

func (s *s3Store) objectKey(parts ...string) string {
	withPrefix := make([]string, 0, len(parts)+1)
	if s.prefix != "" {
		withPrefix = append(withPrefix, s.prefix)
	}
	withPrefix = append(withPrefix, parts...)
	return joinObjectKey(withPrefix...)
}

func (s *s3Store) artifactPublicURL(key string) string {
	if s.publicBaseURL == "" {
		return ""
	}
	return s.publicBaseURL + "/" + strings.TrimLeft(key, "/")
}

func (s *s3Store) fetchManifest(ctx context.Context) (Manifest, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &s.manifestKey,
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "notfound") || strings.Contains(strings.ToLower(err.Error()), "nosuchkey") {
			return Manifest{}, ErrManifestNotFound
		}
		return Manifest{}, fmt.Errorf("fetch manifest from s3: %w", err)
	}
	defer resp.Body.Close()

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	if manifest.Releases == nil {
		manifest.Releases = map[string]map[string]ReleaseDescriptor{}
	}
	return manifest, nil
}

func (s *s3Store) uploadFile(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		Body:        body,
		ContentType: &contentType,
	})
	if err != nil {
		return fmt.Errorf("upload %s to s3: %w", key, err)
	}
	return nil
}

func (s *s3Store) uploadBytes(ctx context.Context, key string, body []byte, contentType string) error {
	return s.uploadFile(ctx, key, bytes.NewReader(body), contentType)
}

func (s *s3Store) downloadToFile(ctx context.Context, key string, destination io.Writer, onProgress func(done, total int64)) (int64, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return 0, fmt.Errorf("download %s from s3: %w", key, err)
	}
	defer resp.Body.Close()

	done := int64(0)
	total := int64(0)
	if resp.ContentLength != nil {
		total = *resp.ContentLength
	}
	buf := make([]byte, 64*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			done += int64(n)
			if _, err := destination.Write(buf[:n]); err != nil {
				return done, err
			}
			if onProgress != nil {
				onProgress(done, total)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return done, readErr
		}
	}
	return done, nil
}

func joinObjectKey(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		filtered = append(filtered, strings.Trim(part, "/"))
	}
	if len(filtered) == 0 {
		return ""
	}
	return path.Join(filtered...)
}
