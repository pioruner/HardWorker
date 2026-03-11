package updater

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type InstallRecord struct {
	AppID       string    `json:"app_id"`
	Version     string    `json:"version"`
	InstalledAt time.Time `json:"installed_at"`
}

func DetectInstalledVersion(app ManagedApp) string {
	recordPath := filepath.Join(app.InstallDir, ".hardworker-release.json")
	raw, err := os.ReadFile(recordPath)
	if err != nil {
		return app.CurrentVersion
	}
	var record InstallRecord
	if json.Unmarshal(raw, &record) != nil {
		return app.CurrentVersion
	}
	if record.Version == "" {
		return app.CurrentVersion
	}
	return record.Version
}

func InstallRelease(app ManagedApp, release ReleaseDescriptor, artifactPath string) error {
	parentDir := filepath.Dir(app.InstallDir)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return fmt.Errorf("create install parent: %w", err)
	}

	stagingBase, err := os.MkdirTemp(parentDir, "hw-updater-stage-*")
	if err != nil {
		return fmt.Errorf("create staging dir: %w", err)
	}
	defer os.RemoveAll(stagingBase)

	stagingRoot, err := Unzip(artifactPath, stagingBase)
	if err != nil {
		return fmt.Errorf("unzip release: %w", err)
	}

	record := InstallRecord{
		AppID:       app.ID,
		Version:     release.Version,
		InstalledAt: time.Now().UTC(),
	}
	recordRaw, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal install record: %w", err)
	}
	if err := os.WriteFile(filepath.Join(stagingRoot, ".hardworker-release.json"), recordRaw, 0o644); err != nil {
		return fmt.Errorf("write install record: %w", err)
	}

	backupDir := app.InstallDir + ".backup"
	_ = os.RemoveAll(backupDir)
	if _, err := os.Stat(app.InstallDir); err == nil {
		if err := os.Rename(app.InstallDir, backupDir); err != nil {
			return fmt.Errorf("move current install aside: %w", err)
		}
	}

	if err := os.Rename(stagingRoot, app.InstallDir); err != nil {
		if _, statErr := os.Stat(backupDir); statErr == nil {
			_ = os.Rename(backupDir, app.InstallDir)
		}
		return fmt.Errorf("activate new install: %w", err)
	}

	_ = os.RemoveAll(backupDir)
	return nil
}
