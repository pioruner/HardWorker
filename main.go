package main

import (
	_ "embed"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/pioruner/HardWorker.git/pkg/akip"
	"github.com/pioruner/HardWorker.git/pkg/app"
	"github.com/pioruner/HardWorker.git/pkg/loger"
	"github.com/pioruner/HardWorker.git/pkg/logger"
	"github.com/pioruner/HardWorker.git/pkg/setts"
	"github.com/pioruner/HardWorker.git/pkg/tray"
	"github.com/pioruner/HardWorker.git/pkg/ui"
	"github.com/pioruner/HardWorker.git/pkg/visko"
)

//go:embed assets/icon.ico
var iconTray []byte

//go:embed assets/icon.png
var iconApp []byte

//go:embed assets/inter.ttf
var fontI []byte

var mod []app.Modules

// Modules
var akiper *akip.AkipUI
var viskos *visko.ViskoUI
var set *setts.SettsUI
var logs *loger.LogUI

func Init() {
	cfg := setts.LoadOrDefault()

	akiper = akip.Init(cfg.AKIPAddress, "Сепаратор Ультразвуковой", cfg.GRPCPort)
	mod = append(mod, akiper)
	//mod = append(mod, visko.Init("192.168.0.200:502", "Вискозиметр Магнитный"))
	set = setts.Init(func(s setts.SettsState) {
		if akiper != nil {
			akiper.SetAddress(s.AKIPAddress)
		}
	})
	mod = append(mod, set)

	logs = loger.New()
	log.SetOutput(io.MultiWriter(os.Stdout, &loger.UiWriter{Ui: logs}))
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	mod = append(mod, logs)
	logger.Infof("Modules initialized: %d", len(mod))
}

func main() {

	if app.CheckInstatse() {
		logger.Warnf("Application is already running, exiting")
		os.Exit(0)
	}
	runtime.LockOSThread()
	logger.Infof("Main thread locked")
	Init()
	app.Run(mod...)
	logger.Infof("All modules started")
	tray.Tray(iconTray) //Create tray icon
	app.Event <- app.EventToggleGUI
	for { //Main Cycle
		switch <-app.Event {
		case app.EventToggleGUI:
			logger.Infof("Event: toggle GUI")
			ui.GUI(iconApp, fontI, mod...)
		case app.EventQuit:
			logger.Warnf("Event: quit requested")
			app.Cancel()
			app.Wg.Wait()
			logger.Infof("Graceful shutdown complete")
			os.Exit(0)
		}
		time.Sleep(time.Millisecond * 100)
	}
}
