package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pioruner/HardWorker.git/pkg/updater"
)

func main() {
	var (
		configPath = flag.String("config", "", "path to uploader config json")
		appID      = flag.String("app", "", "app id, for example akip or visco")
		version    = flag.String("version", "", "release version")
		platform   = flag.String("platform", "windows", "target platform")
		arch       = flag.String("arch", "amd64", "target architecture")
		sourcePath = flag.String("source", "", "path to build output directory or zip file")
		fileName   = flag.String("file-name", "", "override zip file name")
		notes      = flag.String("notes", "", "release notes")
	)
	flag.Parse()

	if *appID == "" || *version == "" || *sourcePath == "" {
		fmt.Fprintln(os.Stderr, "required flags: -app -version -source")
		os.Exit(2)
	}

	resolvedConfig, err := updater.ResolveNamedConfigPath("uploader.local.json", *configPath, "HARDWORKER_UPLOADER_CONFIG")
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve config: %v\n", err)
		os.Exit(1)
	}

	cfg, err := updater.LoadConfig(resolvedConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: 30 * time.Minute}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
	defer cancel()

	manifest, release, err := updater.PublishRelease(ctx, client, cfg, updater.PublishParams{
		AppID:      *appID,
		Version:    *version,
		Platform:   *platform,
		Arch:       *arch,
		SourcePath: *sourcePath,
		FileName:   *fileName,
		Notes:      *notes,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "publish release: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("published %s %s for %s/%s\n", release.AppID, release.Version, release.Platform, release.Arch)
	fmt.Printf("artifact: %s\n", release.URL)
	fmt.Printf("sha256: %s\n", release.SHA256)
	fmt.Printf("manifest updated at: %s\n", manifest.GeneratedAt.Format(time.RFC3339))
}
