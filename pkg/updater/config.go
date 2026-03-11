package updater

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func LoadConfig(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if strings.TrimSpace(cfg.ManifestURL) == "" && strings.TrimSpace(cfg.Publish.ManifestURL) != "" {
		cfg.ManifestURL = cfg.Publish.ManifestURL
	}
	if cfg.HTTPHeaders == nil {
		cfg.HTTPHeaders = map[string]string{}
	}
	if cfg.Publish.Headers == nil {
		cfg.Publish.Headers = map[string]string{}
	}
	normalizeS3Config(&cfg)
	if cfg.Apps == nil {
		cfg.Apps = []ManagedApp{}
	}
	return cfg, nil
}

func LoadMasterConfig(path string) (MasterConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return MasterConfig{}, fmt.Errorf("read master config: %w", err)
	}

	var cfg MasterConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return MasterConfig{}, fmt.Errorf("parse master config: %w", err)
	}
	if cfg.HTTPHeaders == nil {
		cfg.HTTPHeaders = map[string]string{}
	}
	if cfg.Publish.Headers == nil {
		cfg.Publish.Headers = map[string]string{}
	}
	normalizeS3Config(&Config{
		ManifestURL: cfg.ManifestURL,
		Publish:     cfg.Publish,
		Storage:     cfg.Storage,
	})
	normalizeMasterConfig(&cfg)
	return cfg, nil
}

func SaveMasterConfig(path string, cfg MasterConfig) error {
	normalizeMasterConfig(&cfg)
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode master config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create master config dir: %w", err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o600); err != nil {
		return fmt.Errorf("write master config: %w", err)
	}
	return nil
}

func SaveConfig(path string, cfg Config) error {
	normalizeS3Config(&cfg)
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func BuildUploaderConfig(master MasterConfig) Config {
	cfg := Config{
		ManifestURL: master.ManifestURL,
		HTTPHeaders: cloneHeaders(master.HTTPHeaders),
		Apps:        append([]ManagedApp(nil), master.Apps...),
		Publish:     master.Publish,
		Storage:     master.Storage,
	}
	normalizeS3Config(&cfg)
	return cfg
}

func BuildUpdaterConfig(master MasterConfig) Config {
	cfg := Config{
		ManifestURL: master.ManifestURL,
		HTTPHeaders: cloneHeaders(master.HTTPHeaders),
		Apps:        append([]ManagedApp(nil), master.Apps...),
		Storage:     master.Storage,
	}
	if cfg.HTTPHeaders == nil {
		cfg.HTTPHeaders = map[string]string{}
	}
	normalizeS3Config(&cfg)
	return cfg
}

func GenerateConfigsFromMaster(masterPath string, master MasterConfig) error {
	uploaderPath := firstNonEmptyString(master.Output.UploaderConfigPath, filepath.Join("config", "uploader.local.json"))
	updaterPath := firstNonEmptyString(master.Output.UpdaterConfigPath, filepath.Join("config", "updater.local.json"))

	if err := SaveConfig(resolveSiblingPath(masterPath, uploaderPath), BuildUploaderConfig(master)); err != nil {
		return err
	}
	if err := SaveConfig(resolveSiblingPath(masterPath, updaterPath), BuildUpdaterConfig(master)); err != nil {
		return err
	}
	return nil
}

func normalizeS3Config(cfg *Config) {
	if cfg.Storage.S3 == nil {
		return
	}
	s3 := cfg.Storage.S3
	if strings.TrimSpace(s3.Endpoint) != "" {
		if parsed, err := url.Parse(s3.Endpoint); err == nil {
			pathPart := strings.Trim(parsed.Path, "/")
			if pathPart != "" {
				parts := strings.Split(pathPart, "/")
				if strings.TrimSpace(s3.Bucket) == "" {
					s3.Bucket = parts[0]
				}
				if s3.Bucket == parts[0] {
					parsed.Path = ""
					s3.Endpoint = strings.TrimRight(parsed.String(), "/")
				}
			}
		}
	}
	if strings.TrimSpace(s3.Region) == "" {
		s3.Region = "ru-central-1"
	}
	if strings.TrimSpace(s3.TenantID) != "" && strings.TrimSpace(s3.AccessKeyID) != "" && !strings.Contains(s3.AccessKeyID, ":") {
		s3.AccessKeyID = s3.TenantID + ":" + s3.AccessKeyID
	}
	if strings.TrimSpace(s3.ManifestKey) == "" {
		s3.ManifestKey = joinObjectKey(s3.Prefix, "manifest.json")
	}
	if strings.TrimSpace(s3.PublicBaseURL) == "" && strings.TrimSpace(s3.Endpoint) != "" && strings.TrimSpace(s3.Bucket) != "" {
		s3.PublicBaseURL = strings.TrimRight(s3.Endpoint, "/") + "/" + strings.Trim(s3.Bucket, "/")
	}
	if !cfg.Storage.S3.UsePathStyle {
		cfg.Storage.S3.UsePathStyle = true
	}
	if strings.TrimSpace(cfg.ManifestURL) == "" && strings.TrimSpace(s3.PublicBaseURL) != "" {
		cfg.ManifestURL = strings.TrimRight(s3.PublicBaseURL, "/") + "/" + strings.TrimLeft(s3.ManifestKey, "/")
	}
	if strings.TrimSpace(cfg.Publish.ManifestURL) == "" {
		cfg.Publish.ManifestURL = cfg.ManifestURL
	}
}

func normalizeMasterConfig(cfg *MasterConfig) {
	tmp := Config{
		ManifestURL: cfg.ManifestURL,
		Publish:     cfg.Publish,
		Storage:     cfg.Storage,
	}
	normalizeS3Config(&tmp)
	cfg.ManifestURL = tmp.ManifestURL
	cfg.Publish = tmp.Publish
	cfg.Storage = tmp.Storage
	if cfg.HTTPHeaders == nil {
		cfg.HTTPHeaders = map[string]string{}
	}
	if cfg.Publish.Headers == nil {
		cfg.Publish.Headers = map[string]string{}
	}
	if cfg.Output.UploaderConfigPath == "" {
		cfg.Output.UploaderConfigPath = filepath.Join("config", "uploader.local.json")
	}
	if cfg.Output.UpdaterConfigPath == "" {
		cfg.Output.UpdaterConfigPath = filepath.Join("config", "updater.local.json")
	}
}

func ResolveNamedConfigPath(fileName string, override string, envName string) (string, error) {
	candidates := make([]string, 0, 6)
	if strings.TrimSpace(override) != "" {
		candidates = append(candidates, override)
	}
	if envName != "" {
		if env := strings.TrimSpace(os.Getenv(envName)); env != "" {
			candidates = append(candidates, env)
		}
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates, filepath.Join(dir, fileName))
		candidates = append(candidates, filepath.Join(dir, "config", fileName))
	}
	candidates = append(candidates, filepath.Join("config", fileName))
	candidates = append(candidates, fileName)

	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("%s not found", fileName)
}

func ResolveConfigPath(override string) (string, error) {
	path, err := ResolveNamedConfigPath("updater.local.json", override, "HARDWORKER_UPDATER_CONFIG")
	if err != nil {
		return "", errors.New("updater config not found")
	}
	return path, nil
}

func cloneHeaders(value map[string]string) map[string]string {
	if value == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(value))
	for k, v := range value {
		out[k] = v
	}
	return out
}

func resolveSiblingPath(basePath, target string) string {
	if filepath.IsAbs(target) || strings.TrimSpace(basePath) == "" {
		return target
	}
	return filepath.Join(filepath.Dir(basePath), target)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
