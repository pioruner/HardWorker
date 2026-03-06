package main

import "context"

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
