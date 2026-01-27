package main

import (
	_ "embed"
	"os"
	"time"

	"github.com/pioruner/HardWorker.git/pkg/app"
	"github.com/pioruner/HardWorker.git/pkg/tray"
	"github.com/pioruner/HardWorker.git/pkg/ui"
)

//go:embed assets/icon.ico
var iconTray []byte

//go:embed assets/icon.png
var iconApp []byte

//go:embed assets/inter.ttf
var fontI []byte

func main() {
	tray.Tray(iconTray) //Create tray icon
	app.Event <- app.EventToggleGUI
	for { //Main Cycle
		switch <-app.Event {
		case app.EventToggleGUI:
			ui.GUI(iconApp, fontI)
		case app.EventQuit:
			os.Exit(0)
		}
		time.Sleep(time.Millisecond * 100)
	}
}
