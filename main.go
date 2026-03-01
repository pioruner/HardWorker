package main

import (
	_ "embed"
	"os"
	"runtime"
	"time"

	"github.com/AllenDang/giu"
	"github.com/pioruner/HardWorker.git/pkg/akip"
	"github.com/pioruner/HardWorker.git/pkg/app"
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
var uim []giu.Widget

// HardWare
var akiper *akip.AkipUI
var viskos *visko.ViskoUI

func Init() {
	/*
		akiper = akip.Init("192.168.0.100:3000", "akip")
		mod = append(mod, akiper)
		uim = append(uim, akiper)
	*/
	viskos = visko.Init("192.168.0.100:502", "visko")
	mod = append(mod, viskos)
	uim = append(uim, viskos)
}

func main() {
	if app.CheckInstatse() {
		os.Exit(0)
	}
	runtime.LockOSThread()
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
- Отследить ошибки - чтоб не падало
- Добавить иконку к ехе файлу
- Курсоры могут быть вне зоны поиска!
*/
