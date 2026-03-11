package main

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	shared "github.com/pioruner/HardWorker.git/pkg/updater"
)

type MasterSnapshot struct {
	ConfigPath string         `json:"config_path"`
	Ready      bool           `json:"ready"`
	Error      string         `json:"error"`
	Settings   MasterSettings `json:"settings"`
	Logs       []LogEntry     `json:"logs"`
}

type MasterSettings struct {
	ProjectName        string             `json:"project_name"`
	ManifestURL        string             `json:"manifest_url"`
	S3Endpoint         string             `json:"s3_endpoint"`
	S3Bucket           string             `json:"s3_bucket"`
	S3Region           string             `json:"s3_region"`
	S3TenantID         string             `json:"s3_tenant_id"`
	S3AccessKeyID      string             `json:"s3_access_key_id"`
	S3SecretKey        string             `json:"s3_secret_key"`
	S3Prefix           string             `json:"s3_prefix"`
	S3PublicBaseURL    string             `json:"s3_public_base_url"`
	UploaderConfigPath string             `json:"uploader_config_path"`
	UpdaterConfigPath  string             `json:"updater_config_path"`
	Apps               []ManagedAppConfig `json:"apps"`
}

type ManagedAppConfig struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Platform   string `json:"platform"`
	Arch       string `json:"arch"`
	InstallDir string `json:"install_dir"`
	Executable string `json:"executable"`
}

type SaveMasterSettingsInput = MasterSettings

type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

type ConfigMasterService struct {
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	configPath string
	config     shared.MasterConfig
	snapshot   MasterSnapshot
}

func NewConfigMasterService() *ConfigMasterService {
	return &ConfigMasterService{
		snapshot: MasterSnapshot{
			Settings: defaultMasterSettings(),
			Logs:     []LogEntry{},
		},
	}
}

func (s *ConfigMasterService) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.loadConfig("")
}

func (s *ConfigMasterService) Shutdown() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *ConfigMasterService) GetSnapshot() MasterSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := s.snapshot
	cp.Logs = append([]LogEntry(nil), s.snapshot.Logs...)
	cp.Settings.Apps = append([]ManagedAppConfig(nil), s.snapshot.Settings.Apps...)
	return cp
}

func (s *ConfigMasterService) SaveMasterSettings(input SaveMasterSettingsInput) MasterSnapshot {
	cfg := buildMasterConfig(input)
	path := s.configPath
	if strings.TrimSpace(path) == "" {
		if resolved, err := shared.ResolveNamedConfigPath("configmaster.local.json", "", "HARDWORKER_CONFIGMASTER_CONFIG"); err == nil {
			path = resolved
		}
	}
	if strings.TrimSpace(path) == "" {
		path = filepath.Join("config", "configmaster.local.json")
	}
	if err := shared.SaveMasterConfig(path, cfg); err != nil {
		s.setError(err.Error())
		return s.GetSnapshot()
	}
	s.mu.Lock()
	s.configPath = path
	s.config = cfg
	s.snapshot.ConfigPath = path
	s.snapshot.Error = ""
	s.snapshot.Ready = true
	s.snapshot.Settings = settingsFromMaster(cfg)
	s.mu.Unlock()
	s.log("INFO", "мастер-конфиг сохранен")
	return s.GetSnapshot()
}

func (s *ConfigMasterService) GenerateConfigs(input SaveMasterSettingsInput) MasterSnapshot {
	snap := s.SaveMasterSettings(input)
	s.mu.RLock()
	cfg := s.config
	path := s.configPath
	s.mu.RUnlock()
	if err := shared.GenerateConfigsFromMaster(path, cfg); err != nil {
		s.setError(err.Error())
		return s.GetSnapshot()
	}
	s.log("INFO", fmt.Sprintf("сгенерированы %s и %s", cfg.Output.UploaderConfigPath, cfg.Output.UpdaterConfigPath))
	return snap
}

func (s *ConfigMasterService) loadConfig(override string) {
	path, err := shared.ResolveNamedConfigPath("configmaster.local.json", override, "HARDWORKER_CONFIGMASTER_CONFIG")
	if err != nil {
		s.mu.Lock()
		s.snapshot.ConfigPath = filepath.Join("config", "configmaster.local.json")
		s.snapshot.Settings = defaultMasterSettings()
		s.snapshot.Ready = true
		s.snapshot.Error = ""
		s.mu.Unlock()
		s.log("INFO", "мастер-конфиг не найден, создан пустой шаблон в памяти")
		return
	}
	cfg, err := shared.LoadMasterConfig(path)
	if err != nil {
		s.setError(err.Error())
		return
	}
	s.mu.Lock()
	s.configPath = path
	s.config = cfg
	s.snapshot.ConfigPath = path
	s.snapshot.Settings = settingsFromMaster(cfg)
	s.snapshot.Ready = true
	s.snapshot.Error = ""
	s.mu.Unlock()
	s.log("INFO", fmt.Sprintf("загружен мастер-конфиг %s", path))
}

func defaultMasterSettings() MasterSettings {
	return MasterSettings{
		S3Region:           "ru-central-1",
		UploaderConfigPath: filepath.Join("config", "uploader.local.json"),
		UpdaterConfigPath:  filepath.Join("config", "updater.local.json"),
		Apps: []ManagedAppConfig{
			{
				ID:         "app",
				Label:      "Application",
				Platform:   runtime.GOOS,
				Arch:       runtime.GOARCH,
				Executable: "app.exe",
				InstallDir: "./app",
			},
		},
	}
}

func buildMasterConfig(input SaveMasterSettingsInput) shared.MasterConfig {
	cfg := shared.MasterConfig{
		ProjectName: strings.TrimSpace(input.ProjectName),
		ManifestURL: strings.TrimSpace(input.ManifestURL),
		HTTPHeaders: map[string]string{},
		Apps:        make([]shared.ManagedApp, 0, len(input.Apps)),
		Publish: shared.PublishConfig{
			ManifestURL: strings.TrimSpace(input.ManifestURL),
			Headers:     map[string]string{},
		},
		Output: shared.OutputConfig{
			UploaderConfigPath: strings.TrimSpace(input.UploaderConfigPath),
			UpdaterConfigPath:  strings.TrimSpace(input.UpdaterConfigPath),
		},
	}
	if input.S3Endpoint != "" || input.S3AccessKeyID != "" || input.S3SecretKey != "" {
		cfg.Storage.S3 = &shared.S3Config{
			TenantID:        strings.TrimSpace(input.S3TenantID),
			Endpoint:        strings.TrimSpace(input.S3Endpoint),
			Bucket:          strings.TrimSpace(input.S3Bucket),
			Region:          strings.TrimSpace(input.S3Region),
			AccessKeyID:     strings.TrimSpace(input.S3AccessKeyID),
			SecretAccessKey: strings.TrimSpace(input.S3SecretKey),
			Prefix:          strings.TrimSpace(input.S3Prefix),
			PublicBaseURL:   strings.TrimSpace(input.S3PublicBaseURL),
			UsePathStyle:    true,
		}
	}
	for _, app := range input.Apps {
		if strings.TrimSpace(app.ID) == "" {
			continue
		}
		cfg.Apps = append(cfg.Apps, shared.ManagedApp{
			ID:         strings.TrimSpace(app.ID),
			Label:      strings.TrimSpace(app.Label),
			Platform:   firstNonEmpty(strings.TrimSpace(app.Platform), runtime.GOOS),
			Arch:       firstNonEmpty(strings.TrimSpace(app.Arch), runtime.GOARCH),
			InstallDir: strings.TrimSpace(app.InstallDir),
			Executable: strings.TrimSpace(app.Executable),
		})
	}
	return cfg
}

func settingsFromMaster(cfg shared.MasterConfig) MasterSettings {
	settings := defaultMasterSettings()
	settings.ProjectName = cfg.ProjectName
	settings.ManifestURL = cfg.ManifestURL
	settings.UploaderConfigPath = cfg.Output.UploaderConfigPath
	settings.UpdaterConfigPath = cfg.Output.UpdaterConfigPath
	settings.Apps = settings.Apps[:0]
	if cfg.Storage.S3 != nil {
		settings.S3Endpoint = cfg.Storage.S3.Endpoint
		settings.S3Bucket = cfg.Storage.S3.Bucket
		settings.S3Region = cfg.Storage.S3.Region
		settings.S3TenantID = cfg.Storage.S3.TenantID
		settings.S3AccessKeyID = cfg.Storage.S3.AccessKeyID
		settings.S3SecretKey = cfg.Storage.S3.SecretAccessKey
		settings.S3Prefix = cfg.Storage.S3.Prefix
		settings.S3PublicBaseURL = cfg.Storage.S3.PublicBaseURL
	}
	for _, app := range cfg.Apps {
		settings.Apps = append(settings.Apps, ManagedAppConfig{
			ID:         app.ID,
			Label:      app.Label,
			Platform:   app.Platform,
			Arch:       app.Arch,
			InstallDir: app.InstallDir,
			Executable: app.Executable,
		})
	}
	if len(settings.Apps) == 0 {
		settings.Apps = defaultMasterSettings().Apps
	}
	return settings
}

func (s *ConfigMasterService) setError(message string) {
	s.mu.Lock()
	s.snapshot.Error = message
	s.snapshot.Ready = false
	s.mu.Unlock()
	s.log("ERROR", message)
}

func (s *ConfigMasterService) log(level, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := LogEntry{
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Level:   level,
		Message: message,
	}
	s.snapshot.Logs = append([]LogEntry{entry}, s.snapshot.Logs...)
	if len(s.snapshot.Logs) > 150 {
		s.snapshot.Logs = s.snapshot.Logs[:150]
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
