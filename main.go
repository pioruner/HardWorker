package main

import (
	_ "embed"
	"os"
	"time"

	"github.com/AllenDang/giu"
	"github.com/pioruner/HardWorker.git/pkg/akip"
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

var mod []app.Modules
var uim []giu.Widget

// HardWare
var akiper *akip.AkipUI

func Init() {
	akiper = akip.Init("192.168.0.100:3000", "akip")
	mod = append(mod, akiper)
	uim = append(uim, akiper)
}

func main() {
	Init()
	app.Run(mod...)
	tray.Tray(iconTray) //Create tray icon
	app.Event <- app.EventToggleGUI
	for { //Main Cycle
		switch <-app.Event {
		case app.EventToggleGUI:
			ui.GUI(iconApp, fontI, uim...)
		case app.EventQuit:
			app.Cancel()
			app.Wg.Wait()
			os.Exit(0)
		}
		time.Sleep(time.Millisecond * 100)
	}
}

//////// TODO //////////
/*
- Определение смещения по данным от осцила
- Информация на UI Hoffset & TimeBase
- Autosearch
*/
