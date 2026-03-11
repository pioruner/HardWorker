package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	service *ConfigMasterService
}

func NewApp() *App {
	return &App{service: NewConfigMasterService()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.service.Start(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	a.service.Shutdown()
}

func (a *App) GetState() MasterSnapshot {
	return a.service.GetSnapshot()
}

func (a *App) SaveMasterSettings(input SaveMasterSettingsInput) MasterSnapshot {
	return a.service.SaveMasterSettings(input)
}

func (a *App) GenerateConfigs(input SaveMasterSettingsInput) MasterSnapshot {
	return a.service.GenerateConfigs(input)
}

func (a *App) SelectDirectory() string {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Выбор папки"})
	if err != nil {
		return ""
	}
	return path
}
