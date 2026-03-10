package main

import (
	"context"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails binding entrypoint.
type App struct {
	ctx   context.Context
	visco *ViskoService
}

func NewApp() *App {
	return &App{
		visco: NewViskoService("wails-visco"),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.visco.Start(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	a.visco.Shutdown()
}

func (a *App) GetSnapshot() ViskoSnapshot {
	return a.visco.GetSnapshot()
}

func (a *App) ApplyControls(in ViskoControls) ViskoSnapshot {
	a.visco.ApplyControls(in)
	return a.visco.GetSnapshot()
}

func (a *App) GetLogs() []LogEntry {
	return a.visco.GetLogs()
}

func (a *App) SetCursorIndex(index int) ViskoSnapshot {
	a.visco.SetCursorIndex(index)
	return a.visco.GetSnapshot()
}

func (a *App) ClearRows() ViskoSnapshot {
	a.visco.ClearRows()
	return a.visco.GetSnapshot()
}

func (a *App) ExportRows() ViskoSnapshot {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Сохранить CSV",
		DefaultFilename: "visco-export.csv",
		Filters: []runtime.FileFilter{{
			DisplayName: "CSV file",
			Pattern:     "*.csv",
		}},
	})
	if err != nil || strings.TrimSpace(path) == "" {
		return a.visco.GetSnapshot()
	}

	if err := a.visco.ExportCSV(path); err != nil {
		a.visco.logError("Export CSV failed: " + err.Error())
	} else {
		a.visco.logInfo("CSV exported: " + path)
	}
	return a.visco.GetSnapshot()
}
