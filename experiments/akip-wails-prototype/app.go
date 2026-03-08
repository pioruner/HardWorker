package main

import (
	"context"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails binding entrypoint.
type App struct {
	ctx  context.Context
	akip *AkipService
}

func NewApp() *App {
	return &App{
		akip: NewAkipService("wails-akip"),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.akip.Start(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	a.akip.Shutdown()
}

func (a *App) GetSnapshot() AkipSnapshot {
	return a.akip.GetSnapshot()
}

func (a *App) ApplyControls(in AkipControls) AkipSnapshot {
	a.akip.ApplyControls(in)
	return a.akip.GetSnapshot()
}

func (a *App) SetRegistration(enabled bool) AkipSnapshot {
	if !enabled {
		a.akip.StopRegistration()
		return a.akip.GetSnapshot()
	}

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Файл регистрации",
		DefaultFilename: "akip-registration.csv",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "CSV file",
				Pattern:     "*.csv",
			},
		},
	})
	if err != nil || strings.TrimSpace(path) == "" {
		return a.akip.GetSnapshot()
	}

	_ = a.akip.StartRegistration(path)
	return a.akip.GetSnapshot()
}

func (a *App) ZeroVolumeReference() AkipSnapshot {
	a.akip.ZeroVolumeReference()
	return a.akip.GetSnapshot()
}
