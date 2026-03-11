package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	shared "github.com/pioruner/HardWorker.git/pkg/updater"
)

type UpdaterSnapshot struct {
	ConfigPath string            `json:"config_path"`
	Ready      bool              `json:"ready"`
	Busy       bool              `json:"busy"`
	Error      string            `json:"error"`
	Apps       []ManagedAppState `json:"apps"`
	Progress   shared.Progress   `json:"progress"`
	Logs       []LogEntry        `json:"logs"`
	Settings   SettingsState     `json:"settings"`
}

type ManagedAppState struct {
	ID             string `json:"id"`
	Label          string `json:"label"`
	Platform       string `json:"platform"`
	Arch           string `json:"arch"`
	InstallDir     string `json:"install_dir"`
	Executable     string `json:"executable"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	HasUpdate      bool   `json:"has_update"`
}

type SettingsState struct {
	ManifestURL     string             `json:"manifest_url"`
	S3Endpoint      string             `json:"s3_endpoint"`
	S3Bucket        string             `json:"s3_bucket"`
	S3Region        string             `json:"s3_region"`
	S3TenantID      string             `json:"s3_tenant_id"`
	S3AccessKeyID   string             `json:"s3_access_key_id"`
	S3SecretKey     string             `json:"s3_secret_key"`
	S3Prefix        string             `json:"s3_prefix"`
	S3PublicBaseURL string             `json:"s3_public_base_url"`
	SelectedAppID   string             `json:"selected_app_id"`
	Apps            []ManagedAppConfig `json:"apps"`
}

type ManagedAppConfig struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Platform   string `json:"platform"`
	Arch       string `json:"arch"`
	InstallDir string `json:"install_dir"`
	Executable string `json:"executable"`
}

type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

type SaveSettingsInput struct {
	ManifestURL     string             `json:"manifest_url"`
	S3Endpoint      string             `json:"s3_endpoint"`
	S3Bucket        string             `json:"s3_bucket"`
	S3Region        string             `json:"s3_region"`
	S3TenantID      string             `json:"s3_tenant_id"`
	S3AccessKeyID   string             `json:"s3_access_key_id"`
	S3SecretKey     string             `json:"s3_secret_key"`
	S3Prefix        string             `json:"s3_prefix"`
	S3PublicBaseURL string             `json:"s3_public_base_url"`
	SelectedAppID   string             `json:"selected_app_id"`
	Apps            []ManagedAppConfig `json:"apps"`
}

type UpdaterService struct {
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	client     *http.Client
	configPath string
	config     shared.Config
	snapshot   UpdaterSnapshot
}

func NewUpdaterService() *UpdaterService {
	return &UpdaterService{
		client: &http.Client{Timeout: 30 * time.Minute},
		snapshot: UpdaterSnapshot{
			Apps: []ManagedAppState{},
			Logs: []LogEntry{},
			Progress: shared.Progress{
				Stage: "idle",
			},
			Settings: SettingsState{
				Apps: []ManagedAppConfig{},
			},
		},
	}
}

func (s *UpdaterService) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.loadConfig("")
	s.Refresh()
}

func (s *UpdaterService) Shutdown() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *UpdaterService) GetSnapshot() UpdaterSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cp := s.snapshot
	cp.Apps = append([]ManagedAppState(nil), s.snapshot.Apps...)
	cp.Logs = append([]LogEntry(nil), s.snapshot.Logs...)
	cp.Settings.Apps = append([]ManagedAppConfig(nil), s.snapshot.Settings.Apps...)
	return cp
}

func (s *UpdaterService) Refresh() {
	s.mu.RLock()
	cfg := s.config
	selectedID := s.snapshot.Settings.SelectedAppID
	s.mu.RUnlock()

	if len(cfg.Apps) == 0 {
		s.setError("добавьте хотя бы одно приложение в настройки")
		s.updateSettingsSnapshot(cfg, selectedID)
		return
	}

	states := make([]ManagedAppState, 0, len(cfg.Apps))
	for _, appCfg := range cfg.Apps {
		resolvedApp := resolveManagedAppPaths(appCfg)
		state := ManagedAppState{
			ID:             resolvedApp.ID,
			Label:          firstNonEmpty(resolvedApp.Label, resolvedApp.ID),
			Platform:       resolvedApp.Platform,
			Arch:           resolvedApp.Arch,
			InstallDir:     resolvedApp.InstallDir,
			Executable:     resolvedApp.Executable,
			CurrentVersion: detectCurrentVersion(resolvedApp),
			Status:         "ready",
			Message:        "проверка обновлений",
		}

		if err := validateManagedApp(resolvedApp); err != nil {
			state.Status = "config"
			state.Message = err.Error()
			states = append(states, state)
			continue
		}

		manifestURL := firstNonEmpty(resolvedApp.ManifestURL, cfg.ManifestURL)
		if strings.TrimSpace(manifestURL) == "" {
			state.Status = "config"
			state.Message = "не задан адрес manifest"
			states = append(states, state)
			continue
		}

		manifest, err := shared.FetchManifestFromConfig(s.ctx, s.client, cfg, manifestURL)
		if err != nil {
			state.Status = "offline"
			state.Message = "связи с сервером нет, обновления недоступны"
			states = append(states, state)
			continue
		}

		release, ok := manifest.Releases[resolvedApp.ID][platformArchKey(resolvedApp.Platform, resolvedApp.Arch)]
		if !ok {
			state.Status = "missing"
			state.Message = "на сервере нет релиза для этой платформы"
			states = append(states, state)
			continue
		}

		state.LatestVersion = release.Version
		state.HasUpdate = compareVersions(release.Version, state.CurrentVersion) > 0
		if state.CurrentVersion == "" {
			state.CurrentVersion = "unknown"
		}
		if state.HasUpdate {
			state.Status = "update-available"
			state.Message = "найдено обновление"
		} else {
			state.Status = "up-to-date"
			state.Message = "установлена актуальная версия"
		}
		states = append(states, state)
	}

	sort.Slice(states, func(i, j int) bool { return states[i].Label < states[j].Label })
	s.updateSettingsSnapshot(cfg, selectedID)

	s.mu.Lock()
	s.snapshot.Ready = true
	s.snapshot.Error = ""
	s.snapshot.Apps = states
	if s.snapshot.Settings.SelectedAppID == "" && len(states) > 0 {
		s.snapshot.Settings.SelectedAppID = states[0].ID
	}
	s.mu.Unlock()
	s.log("INFO", "обновлен статус приложения и сервера")
}

func (s *UpdaterService) SaveSettings(input SaveSettingsInput) UpdaterSnapshot {
	cfg := buildConfigFromSettings(input)
	configPath := s.configPath
	if strings.TrimSpace(configPath) == "" {
		resolved, err := shared.ResolveConfigPath("")
		if err == nil {
			configPath = resolved
		}
	}
	if strings.TrimSpace(configPath) == "" {
		configPath = filepath.Join("config", "updater.local.json")
	}
	if err := shared.SaveConfig(configPath, cfg); err != nil {
		s.setError(err.Error())
		return s.GetSnapshot()
	}

	s.mu.Lock()
	s.config = cfg
	s.configPath = configPath
	s.snapshot.ConfigPath = configPath
	s.snapshot.Error = ""
	s.mu.Unlock()
	s.updateSettingsSnapshot(cfg, input.SelectedAppID)
	s.log("INFO", "настройки сохранены")
	s.Refresh()
	return s.GetSnapshot()
}

func (s *UpdaterService) StartUpdate(appID string) UpdaterSnapshot {
	s.mu.Lock()
	if s.snapshot.Busy {
		s.mu.Unlock()
		return s.snapshot
	}
	s.snapshot.Busy = true
	s.snapshot.Progress = shared.Progress{
		Stage:     "prepare",
		Message:   "подготовка к обновлению",
		StartedAt: time.Now().UTC(),
	}
	s.mu.Unlock()

	go s.runUpdate(appID)
	return s.GetSnapshot()
}

func (s *UpdaterService) LaunchApp(appID string) error {
	s.mu.RLock()
	cfg := s.config
	s.mu.RUnlock()

	appCfg, err := findManagedApp(cfg.Apps, appID)
	if err != nil {
		return err
	}
	resolved := resolveManagedAppPaths(appCfg)
	targetPath, err := resolveExecutablePath(resolved)
	if err != nil {
		return err
	}
	cmd := exec.Command(targetPath)
	cmd.Dir = resolved.InstallDir
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch app: %w", err)
	}
	s.log("INFO", fmt.Sprintf("запуск приложения %s", targetPath))
	return nil
}

func (s *UpdaterService) runUpdate(appID string) {
	defer func() {
		s.mu.Lock()
		s.snapshot.Busy = false
		s.snapshot.Progress.FinishedAt = time.Now().UTC()
		s.mu.Unlock()
	}()

	s.mu.RLock()
	cfg := s.config
	s.mu.RUnlock()

	target, err := findManagedApp(cfg.Apps, appID)
	if err != nil {
		s.failProgress(err.Error())
		return
	}
	target = resolveManagedAppPaths(target)
	if err := validateManagedApp(target); err != nil {
		s.failProgress(err.Error())
		return
	}

	manifestURL := firstNonEmpty(target.ManifestURL, cfg.ManifestURL)
	manifest, err := shared.FetchManifestFromConfig(s.ctx, s.client, cfg, manifestURL)
	if err != nil {
		s.failProgress("нет связи с сервером обновлений")
		return
	}

	release, ok := manifest.Releases[target.ID][platformArchKey(target.Platform, target.Arch)]
	if !ok {
		s.failProgress("релиз не найден на сервере")
		return
	}

	currentVersion := detectCurrentVersion(target)
	if compareVersions(release.Version, currentVersion) <= 0 {
		s.setProgress("done", "обновление не требуется", 100, 0, 0)
		s.log("INFO", fmt.Sprintf("%s уже на версии %s", target.ID, currentVersion))
		s.Refresh()
		return
	}

	downloadDir, err := os.MkdirTemp("", "hardworker-update-download-*")
	if err != nil {
		s.failProgress(err.Error())
		return
	}
	defer os.RemoveAll(downloadDir)

	s.setProgress("download", fmt.Sprintf("скачивание %s %s", target.Label, release.Version), 0, 0, release.Size)
	artifactPath, err := shared.DownloadReleaseFromConfig(s.ctx, s.client, cfg, release, downloadDir, func(done, total int64) {
		percent := percentOf(done, total)
		s.setProgress("download", fmt.Sprintf("скачивание %s", filepath.Base(release.URL)), percent, done, total)
	})
	if err != nil {
		s.failProgress(err.Error())
		return
	}

	s.mu.Lock()
	s.snapshot.Progress.DownloadPath = artifactPath
	s.mu.Unlock()

	s.setProgress("install", "установка новой версии", 92, 0, 0)
	if err := shared.InstallRelease(target, release, artifactPath); err != nil {
		s.failProgress(err.Error())
		return
	}

	s.setProgress("done", fmt.Sprintf("%s обновлен до %s", target.Label, release.Version), 100, release.Size, release.Size)
	s.log("INFO", fmt.Sprintf("%s обновлен до %s", target.ID, release.Version))
	s.Refresh()
}

func (s *UpdaterService) loadConfig(override string) {
	configPath, err := shared.ResolveConfigPath(override)
	if err != nil {
		s.setError(err.Error())
		return
	}

	cfg, err := shared.LoadConfig(configPath)
	if err != nil {
		s.setError(err.Error())
		return
	}

	selectedID := ""
	if len(cfg.Apps) > 0 {
		selectedID = cfg.Apps[0].ID
	}

	s.mu.Lock()
	s.configPath = configPath
	s.config = cfg
	s.snapshot.ConfigPath = configPath
	s.snapshot.Error = ""
	s.mu.Unlock()
	s.updateSettingsSnapshot(cfg, selectedID)
	s.log("INFO", fmt.Sprintf("загружен конфиг %s", configPath))
}

func (s *UpdaterService) updateSettingsSnapshot(cfg shared.Config, selectedID string) {
	settings := SettingsState{
		ManifestURL:   cfg.ManifestURL,
		SelectedAppID: selectedID,
		Apps:          make([]ManagedAppConfig, 0, len(cfg.Apps)),
	}
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
	for _, appCfg := range cfg.Apps {
		settings.Apps = append(settings.Apps, ManagedAppConfig{
			ID:         appCfg.ID,
			Label:      appCfg.Label,
			Platform:   appCfg.Platform,
			Arch:       appCfg.Arch,
			InstallDir: appCfg.InstallDir,
			Executable: appCfg.Executable,
		})
	}
	if settings.SelectedAppID == "" && len(settings.Apps) > 0 {
		settings.SelectedAppID = settings.Apps[0].ID
	}

	s.mu.Lock()
	s.snapshot.Settings = settings
	s.mu.Unlock()
}

func buildConfigFromSettings(input SaveSettingsInput) shared.Config {
	cfg := shared.Config{
		ManifestURL: strings.TrimSpace(input.ManifestURL),
		HTTPHeaders: map[string]string{},
		Apps:        make([]shared.ManagedApp, 0, len(input.Apps)),
		Publish: shared.PublishConfig{
			ManifestURL: strings.TrimSpace(input.ManifestURL),
			Headers:     map[string]string{},
		},
	}

	hasS3 := strings.TrimSpace(input.S3Endpoint) != "" || strings.TrimSpace(input.S3AccessKeyID) != "" || strings.TrimSpace(input.S3SecretKey) != ""
	if hasS3 {
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

func (s *UpdaterService) setError(message string) {
	s.mu.Lock()
	s.snapshot.Error = message
	s.snapshot.Ready = false
	s.mu.Unlock()
	s.log("ERROR", message)
}

func (s *UpdaterService) setProgress(stage, message string, percent float64, done, total int64) {
	s.mu.Lock()
	s.snapshot.Progress.Stage = stage
	s.snapshot.Progress.Message = message
	s.snapshot.Progress.Percent = percent
	s.snapshot.Progress.BytesDone = done
	s.snapshot.Progress.BytesTotal = total
	s.mu.Unlock()
}

func (s *UpdaterService) failProgress(message string) {
	s.mu.Lock()
	s.snapshot.Progress.Stage = "error"
	s.snapshot.Progress.Message = message
	s.snapshot.Progress.Percent = 0
	s.snapshot.Error = message
	s.mu.Unlock()
	s.log("ERROR", message)
}

func (s *UpdaterService) log(level, message string) {
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

func validateManagedApp(app shared.ManagedApp) error {
	if strings.TrimSpace(app.InstallDir) == "" {
		return errors.New("не указан путь к приложению")
	}
	if strings.TrimSpace(app.Executable) == "" {
		return errors.New("не указано имя exe файла")
	}
	if err := validateInstallTargetAgainstUpdater(app.InstallDir); err != nil {
		return err
	}
	return nil
}

func validateInstallTargetAgainstUpdater(installDir string) error {
	updaterExe, err := os.Executable()
	if err != nil {
		return nil
	}
	updaterDir := filepath.Clean(filepath.Dir(updaterExe))
	targetDir := filepath.Clean(installDir)
	rel, err := filepath.Rel(targetDir, updaterDir)
	if err != nil {
		return nil
	}
	if rel == "." || (!strings.HasPrefix(rel, "..") && rel != "..") {
		return errors.New("нельзя обновлять каталог, который содержит сам updater; выберите отдельную папку приложения")
	}
	return nil
}

func detectCurrentVersion(app shared.ManagedApp) string {
	version := shared.DetectInstalledVersion(app)
	if version != "" {
		return version
	}
	targetPath, err := resolveExecutablePath(app)
	if err != nil {
		return app.CurrentVersion
	}
	if _, err := os.Stat(targetPath); err == nil {
		return firstNonEmpty(app.CurrentVersion, "installed")
	}
	return firstNonEmpty(app.CurrentVersion, "not installed")
}

func resolveExecutablePath(app shared.ManagedApp) (string, error) {
	if strings.TrimSpace(app.InstallDir) == "" {
		return "", errors.New("install dir is empty")
	}
	if strings.TrimSpace(app.Executable) == "" {
		return "", errors.New("executable is empty")
	}
	executable := app.Executable
	if filepath.IsAbs(executable) {
		return executable, nil
	}
	return filepath.Join(app.InstallDir, executable), nil
}

func findManagedApp(apps []shared.ManagedApp, appID string) (shared.ManagedApp, error) {
	for _, appCfg := range apps {
		if appCfg.ID == appID {
			return appCfg, nil
		}
	}
	return shared.ManagedApp{}, fmt.Errorf("приложение %s не найдено в настройках", appID)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func percentOf(done, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return (float64(done) / float64(total)) * 100
}

func compareVersions(left, right string) int {
	if left == right {
		return 0
	}
	if strings.TrimSpace(right) == "" {
		return 1
	}
	a := splitVersion(left)
	b := splitVersion(right)
	size := len(a)
	if len(b) > size {
		size = len(b)
	}
	for i := 0; i < size; i++ {
		ai := 0
		bi := 0
		if i < len(a) {
			ai = a[i]
		}
		if i < len(b) {
			bi = b[i]
		}
		if ai > bi {
			return 1
		}
		if ai < bi {
			return -1
		}
	}
	return strings.Compare(left, right)
}

func splitVersion(value string) []int {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '.' || r == '-' || r == '_'
	})
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			result = append(result, 0)
			continue
		}
		result = append(result, n)
	}
	return result
}

func platformArchKey(platform, arch string) string {
	return platform + "-" + arch
}

func resolveManagedAppPaths(app shared.ManagedApp) shared.ManagedApp {
	app.InstallDir = resolveAppPath(app.InstallDir)
	if filepath.IsAbs(app.Executable) {
		return app
	}
	return app
}

func resolveAppPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return trimmed
	}
	if filepath.IsAbs(trimmed) {
		return trimmed
	}
	exe, err := os.Executable()
	if err != nil {
		return trimmed
	}
	return filepath.Clean(filepath.Join(filepath.Dir(exe), trimmed))
}
