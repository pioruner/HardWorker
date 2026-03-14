package main

import (
	"context"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	service *RP40Service
}

func NewApp() *App {
	return &App{
		service: NewRP40Service(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.service.SetContext(ctx)
}

func (a *App) GetSnapshot() RP40Snapshot {
	return a.service.GetSnapshot()
}

func (a *App) SelectPassportFile() RP40Snapshot {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Паспортный файл образцов",
		Filters: []runtime.FileFilter{{
			DisplayName: "Passport file",
			Pattern:     "*.xls;*.csv;*.txt",
		}},
	})
	if err != nil || strings.TrimSpace(path) == "" {
		return a.service.GetSnapshot()
	}
	a.service.LoadPassport(path)
	return a.service.GetSnapshot()
}

func (a *App) SelectMeasurementFile() RP40Snapshot {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Файл замера RP40",
		Filters: []runtime.FileFilter{{
			DisplayName: "Measurement file",
			Pattern:     "*.xls;*.csv;*.txt",
		}},
	})
	if err != nil || strings.TrimSpace(path) == "" {
		return a.service.GetSnapshot()
	}
	a.service.SetMeasurementFile(path)
	return a.service.GetSnapshot()
}

func (a *App) SetSelectedSample(sampleID string) RP40Snapshot {
	a.service.SetSelectedSample(sampleID)
	return a.service.GetSnapshot()
}

func (a *App) UpdateInputs(in RP40Inputs) RP40Snapshot {
	a.service.UpdateInputs(in)
	return a.service.GetSnapshot()
}

func (a *App) Calculate() RP40Snapshot {
	a.service.Calculate()
	return a.service.GetSnapshot()
}
