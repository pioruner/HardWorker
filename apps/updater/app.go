package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	service *UpdaterService
}

func NewApp() *App {
	return &App{
		service: NewUpdaterService(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.service.Start(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	a.service.Shutdown()
}

func (a *App) GetState() UpdaterSnapshot {
	return a.service.GetSnapshot()
}

func (a *App) Refresh() UpdaterSnapshot {
	a.service.Refresh()
	return a.service.GetSnapshot()
}

func (a *App) SaveSettings(input SaveSettingsInput) UpdaterSnapshot {
	return a.service.SaveSettings(input)
}

func (a *App) StartUpdate(appID string) UpdaterSnapshot {
	return a.service.StartUpdate(appID)
}

func (a *App) SelectInstallDir() string {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Папка приложения",
	})
	if err != nil {
		return ""
	}
	return path
}

func (a *App) LaunchApp(appID string) string {
	if err := a.service.LaunchApp(appID); err != nil {
		return err.Error()
	}
	runtime.Quit(a.ctx)
	return ""
}
