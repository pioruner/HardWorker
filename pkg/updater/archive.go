package updater

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ZipPath(sourcePath, destination string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("create zip: %w", err)
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	baseDir := filepath.Dir(sourcePath)
	if !info.IsDir() {
		baseDir = filepath.Dir(sourcePath)
	}

	return filepath.Walk(sourcePath, func(path string, fileInfo os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if fileInfo.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}
		header.Name = rel
		header.Method = zip.Deflate

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		if _, err := io.Copy(writer, in); err != nil {
			return err
		}
		return nil
	})
}

func Unzip(zipPath, destination string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("open zip: %w", err)
	}
	defer reader.Close()

	if err := os.MkdirAll(destination, 0o755); err != nil {
		return "", fmt.Errorf("create destination: %w", err)
	}

	for _, file := range reader.File {
		target := filepath.Join(destination, file.Name)
		cleanDest := filepath.Clean(destination) + string(os.PathSeparator)
		cleanTarget := filepath.Clean(target)
		if !strings.HasPrefix(cleanTarget, cleanDest) && cleanTarget != filepath.Clean(destination) {
			return "", fmt.Errorf("invalid zip path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(cleanTarget, file.Mode()); err != nil {
				return "", err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(cleanTarget), 0o755); err != nil {
			return "", err
		}

		in, err := file.Open()
		if err != nil {
			return "", err
		}

		out, err := os.OpenFile(cleanTarget, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, file.Mode())
		if err != nil {
			in.Close()
			return "", err
		}

		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		in.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
	}

	return normalizeExtractRoot(destination)
}

func normalizeExtractRoot(destination string) (string, error) {
	entries, err := os.ReadDir(destination)
	if err != nil {
		return "", err
	}
	if len(entries) == 1 && entries[0].IsDir() {
		return filepath.Join(destination, entries[0].Name()), nil
	}
	return destination, nil
}
