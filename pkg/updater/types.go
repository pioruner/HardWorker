package updater

import "time"

type Config struct {
	ManifestURL string            `json:"manifest_url"`
	HTTPHeaders map[string]string `json:"http_headers"`
	Apps        []ManagedApp      `json:"apps"`
	Publish     PublishConfig     `json:"publish"`
	Storage     StorageConfig     `json:"storage"`
}

type MasterConfig struct {
	ProjectName string            `json:"project_name"`
	ManifestURL string            `json:"manifest_url"`
	HTTPHeaders map[string]string `json:"http_headers"`
	Apps        []ManagedApp      `json:"apps"`
	Publish     PublishConfig     `json:"publish"`
	Storage     StorageConfig     `json:"storage"`
	Output      OutputConfig      `json:"output"`
}

type OutputConfig struct {
	UploaderConfigPath string `json:"uploader_config_path"`
	UpdaterConfigPath  string `json:"updater_config_path"`
}

type ManagedApp struct {
	ID             string `json:"id"`
	Label          string `json:"label"`
	Platform       string `json:"platform"`
	Arch           string `json:"arch"`
	InstallDir     string `json:"install_dir"`
	Executable     string `json:"executable"`
	CurrentVersion string `json:"current_version"`
	ManifestURL    string `json:"manifest_url"`
}

type PublishConfig struct {
	ManifestURL               string            `json:"manifest_url"`
	ManifestUploadURL         string            `json:"manifest_upload_url"`
	ArtifactURLTemplate       string            `json:"artifact_url_template"`
	ArtifactUploadURLTemplate string            `json:"artifact_upload_url_template"`
	Headers                   map[string]string `json:"headers"`
}

type StorageConfig struct {
	S3 *S3Config `json:"s3"`
}

type S3Config struct {
	TenantID        string `json:"tenant_id"`
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Prefix          string `json:"prefix"`
	PublicBaseURL   string `json:"public_base_url"`
	UsePathStyle    bool   `json:"use_path_style"`
	ManifestKey     string `json:"manifest_key"`
}

type Manifest struct {
	GeneratedAt time.Time                               `json:"generated_at"`
	Releases    map[string]map[string]ReleaseDescriptor `json:"releases"`
}

type ReleaseDescriptor struct {
	AppID       string    `json:"app_id"`
	Version     string    `json:"version"`
	Platform    string    `json:"platform"`
	Arch        string    `json:"arch"`
	URL         string    `json:"url"`
	ObjectKey   string    `json:"object_key,omitempty"`
	SHA256      string    `json:"sha256"`
	Size        int64     `json:"size"`
	PublishedAt time.Time `json:"published_at"`
	Notes       string    `json:"notes,omitempty"`
}

type PublishParams struct {
	AppID      string
	Version    string
	Platform   string
	Arch       string
	SourcePath string
	FileName   string
	Notes      string
}

type Progress struct {
	Stage        string    `json:"stage"`
	Message      string    `json:"message"`
	Percent      float64   `json:"percent"`
	BytesDone    int64     `json:"bytes_done"`
	BytesTotal   int64     `json:"bytes_total"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
	DownloadPath string    `json:"download_path,omitempty"`
}

func releaseKey(platform, arch string) string {
	return platform + "-" + arch
}
