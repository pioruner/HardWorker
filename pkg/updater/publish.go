package updater

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func PublishRelease(ctx context.Context, client *http.Client, cfg Config, params PublishParams) (Manifest, ReleaseDescriptor, error) {
	artifactFileName := params.FileName
	if strings.TrimSpace(artifactFileName) == "" {
		artifactFileName = fmt.Sprintf("%s-%s-%s-%s.zip", params.AppID, params.Platform, params.Arch, params.Version)
	}

	tempZip, cleanup, err := prepareZip(params.SourcePath, artifactFileName)
	if err != nil {
		return Manifest{}, ReleaseDescriptor{}, err
	}
	defer cleanup()

	size, digest, err := fileDigest(tempZip)
	if err != nil {
		return Manifest{}, ReleaseDescriptor{}, err
	}

	artifactURL := applyTemplate(cfg.Publish.ArtifactURLTemplate, params, artifactFileName)
	uploadURL := applyTemplate(cfg.Publish.ArtifactUploadURLTemplate, params, artifactFileName)
	objectKey := ""
	if cfg.Storage.S3 != nil {
		store, err := newS3Store(ctx, cfg.Storage.S3)
		if err != nil {
			return Manifest{}, ReleaseDescriptor{}, err
		}
		objectKey = store.objectKey(params.AppID, params.Platform, params.Arch, params.Version, artifactFileName)
		if artifactURL == "" {
			artifactURL = store.artifactPublicURL(objectKey)
		}
	}
	release := ReleaseDescriptor{
		AppID:       params.AppID,
		Version:     params.Version,
		Platform:    params.Platform,
		Arch:        params.Arch,
		URL:         artifactURL,
		ObjectKey:   objectKey,
		SHA256:      digest,
		Size:        size,
		PublishedAt: time.Now().UTC(),
		Notes:       params.Notes,
	}

	if cfg.Storage.S3 != nil {
		store, err := newS3Store(ctx, cfg.Storage.S3)
		if err != nil {
			return Manifest{}, ReleaseDescriptor{}, err
		}
		if err := uploadFileToS3(ctx, store, objectKey, tempZip); err != nil {
			return Manifest{}, ReleaseDescriptor{}, err
		}
	} else {
		if err := uploadFile(ctx, client, uploadURL, cfg.Publish.Headers, tempZip, "application/zip"); err != nil {
			return Manifest{}, ReleaseDescriptor{}, err
		}
	}

	manifestURL := cfg.Publish.ManifestURL
	if strings.TrimSpace(manifestURL) == "" {
		manifestURL = cfg.ManifestURL
	}

	manifest, err := FetchManifestFromConfig(ctx, client, cfg, manifestURL)
	if err != nil {
		if !errors.Is(err, ErrManifestNotFound) {
			return Manifest{}, ReleaseDescriptor{}, err
		}
		manifest = Manifest{Releases: map[string]map[string]ReleaseDescriptor{}}
	}
	if manifest.Releases == nil {
		manifest.Releases = map[string]map[string]ReleaseDescriptor{}
	}
	if manifest.Releases[params.AppID] == nil {
		manifest.Releases[params.AppID] = map[string]ReleaseDescriptor{}
	}
	manifest.GeneratedAt = time.Now().UTC()
	manifest.Releases[params.AppID][releaseKey(params.Platform, params.Arch)] = release

	body, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return Manifest{}, ReleaseDescriptor{}, fmt.Errorf("encode manifest: %w", err)
	}
	manifestUploadURL := cfg.Publish.ManifestUploadURL
	if strings.TrimSpace(manifestUploadURL) == "" {
		manifestUploadURL = manifestURL
	}
	if cfg.Storage.S3 != nil {
		store, err := newS3Store(ctx, cfg.Storage.S3)
		if err != nil {
			return Manifest{}, ReleaseDescriptor{}, err
		}
		if err := store.uploadBytes(ctx, store.manifestKey, body, "application/json"); err != nil {
			return Manifest{}, ReleaseDescriptor{}, err
		}
	} else {
		if err := uploadBytes(ctx, client, manifestUploadURL, cfg.Publish.Headers, body, "application/json"); err != nil {
			return Manifest{}, ReleaseDescriptor{}, err
		}
	}

	return manifest, release, nil
}

func uploadFileToS3(ctx context.Context, store *s3Store, key string, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open upload file: %w", err)
	}
	defer file.Close()
	return store.uploadFile(ctx, key, file, "application/zip")
}

func prepareZip(sourcePath, fallbackFileName string) (string, func(), error) {
	if strings.HasSuffix(strings.ToLower(sourcePath), ".zip") {
		return sourcePath, func() {}, nil
	}

	tempDir, err := os.MkdirTemp("", "hw-publish-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp zip dir: %w", err)
	}
	zipPath := filepath.Join(tempDir, fallbackFileName)
	if err := ZipPath(sourcePath, zipPath); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", nil, err
	}
	return zipPath, func() { _ = os.RemoveAll(tempDir) }, nil
}

func fileDigest(path string) (int64, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, "", fmt.Errorf("open artifact: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	size, err := io.Copy(hasher, file)
	if err != nil {
		return 0, "", fmt.Errorf("hash artifact: %w", err)
	}
	return size, hex.EncodeToString(hasher.Sum(nil)), nil
}

func uploadFile(ctx context.Context, client *http.Client, target string, headers map[string]string, path string, contentType string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open upload file: %w", err)
	}
	defer file.Close()
	return uploadReader(ctx, client, target, headers, file, contentType)
}

func uploadBytes(ctx context.Context, client *http.Client, target string, headers map[string]string, body []byte, contentType string) error {
	return uploadReader(ctx, client, target, headers, bytes.NewReader(body), contentType)
}

func uploadReader(ctx context.Context, client *http.Client, target string, headers map[string]string, body io.Reader, contentType string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, target, body)
	if err != nil {
		return fmt.Errorf("build upload request: %w", err)
	}
	applyHeaders(req, headers)
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload %s: %w", target, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upload %s: unexpected status %s", target, resp.Status)
	}
	return nil
}

func applyTemplate(template string, params PublishParams, fileName string) string {
	replacer := strings.NewReplacer(
		"{app}", params.AppID,
		"{version}", params.Version,
		"{platform}", params.Platform,
		"{arch}", params.Arch,
		"{file}", fileName,
	)
	return replacer.Replace(template)
}
