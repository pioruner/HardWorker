package updater

import (
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
)

var ErrManifestNotFound = errors.New("manifest not found")

func FetchManifest(ctx context.Context, client *http.Client, manifestURL string, headers map[string]string) (Manifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return Manifest{}, fmt.Errorf("build manifest request: %w", err)
	}
	applyHeaders(req, headers)

	resp, err := client.Do(req)
	if err != nil {
		return Manifest{}, fmt.Errorf("fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Manifest{}, ErrManifestNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Manifest{}, fmt.Errorf("fetch manifest: unexpected status %s", resp.Status)
	}

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	if manifest.Releases == nil {
		manifest.Releases = map[string]map[string]ReleaseDescriptor{}
	}
	return manifest, nil
}

func FetchManifestFromConfig(ctx context.Context, client *http.Client, cfg Config, manifestURL string) (Manifest, error) {
	if cfg.Storage.S3 != nil {
		store, err := newS3Store(ctx, cfg.Storage.S3)
		if err != nil {
			return Manifest{}, err
		}
		return store.fetchManifest(ctx)
	}
	return FetchManifest(ctx, client, manifestURL, cfg.HTTPHeaders)
}

func DownloadRelease(ctx context.Context, client *http.Client, release ReleaseDescriptor, headers map[string]string, destination string, onProgress func(done, total int64)) (string, error) {
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return "", fmt.Errorf("create download dir: %w", err)
	}

	name := filepath.Base(release.URL)
	if strings.TrimSpace(name) == "" || name == "." || name == "/" {
		name = fmt.Sprintf("%s-%s-%s.zip", release.AppID, release.Platform, release.Version)
	}
	target := filepath.Join(destination, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, release.URL, nil)
	if err != nil {
		return "", fmt.Errorf("build download request: %w", err)
	}
	applyHeaders(req, headers)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download release: unexpected status %s", resp.Status)
	}

	out, err := os.Create(target)
	if err != nil {
		return "", fmt.Errorf("create artifact file: %w", err)
	}
	defer out.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(out, hasher)
	done := int64(0)
	total := resp.ContentLength
	buf := make([]byte, 64*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			done += int64(n)
			if _, err := writer.Write(buf[:n]); err != nil {
				return "", fmt.Errorf("write artifact: %w", err)
			}
			if onProgress != nil {
				onProgress(done, total)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return "", fmt.Errorf("download artifact: %w", readErr)
		}
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if release.SHA256 != "" && !strings.EqualFold(actualHash, release.SHA256) {
		return "", fmt.Errorf("sha256 mismatch: got %s want %s", actualHash, release.SHA256)
	}
	return target, nil
}

func DownloadReleaseFromConfig(ctx context.Context, client *http.Client, cfg Config, release ReleaseDescriptor, destination string, onProgress func(done, total int64)) (string, error) {
	if cfg.Storage.S3 != nil && strings.TrimSpace(release.ObjectKey) != "" {
		store, err := newS3Store(ctx, cfg.Storage.S3)
		if err != nil {
			return "", err
		}
		if err := os.MkdirAll(destination, 0o755); err != nil {
			return "", fmt.Errorf("create download dir: %w", err)
		}
		name := filepath.Base(release.ObjectKey)
		if strings.TrimSpace(name) == "" {
			name = fmt.Sprintf("%s-%s-%s.zip", release.AppID, release.Platform, release.Version)
		}
		target := filepath.Join(destination, name)
		out, err := os.Create(target)
		if err != nil {
			return "", fmt.Errorf("create artifact file: %w", err)
		}
		defer out.Close()

		hasher := sha256.New()
		writer := io.MultiWriter(out, hasher)
		if _, err := store.downloadToFile(ctx, release.ObjectKey, writer, onProgress); err != nil {
			return "", err
		}
		actualHash := hex.EncodeToString(hasher.Sum(nil))
		if release.SHA256 != "" && !strings.EqualFold(actualHash, release.SHA256) {
			return "", fmt.Errorf("sha256 mismatch: got %s want %s", actualHash, release.SHA256)
		}
		return target, nil
	}

	return DownloadRelease(ctx, client, release, cfg.HTTPHeaders, destination, onProgress)
}

func applyHeaders(req *http.Request, headers map[string]string) {
	for key, value := range headers {
		if strings.TrimSpace(key) == "" {
			continue
		}
		req.Header.Set(key, value)
	}
}
