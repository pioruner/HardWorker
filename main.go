package main

import (
	"context"
	_ "embed"
	"os"
	"sync"
	"time"

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

var ctx, cancel = context.WithCancel(context.Background())
var wg sync.WaitGroup

type x interface {
	Run()
}

func Run(modules ...x) {
	for _, i := range modules {
		i.Run()
	}
}

// HardWare
var akiper *akip.AkipUI

func Init() {
	akiper = akip.Init("192.168.0.100:3000", "akip", ctx, &wg)
}

func main() {
	Init()
	Run(akiper)
	tray.Tray(iconTray) //Create tray icon
	app.Event <- app.EventToggleGUI
	for { //Main Cycle
		switch <-app.Event {
		case app.EventToggleGUI:
			ui.GUI(iconApp, fontI, akiper)
		case app.EventQuit:
			cancel()
			wg.Wait()
			os.Exit(0)
		}
		time.Sleep(time.Millisecond * 100)
	}
}

//////// TODO //////////
/*
- Расчеты!
- Определение смещения по данным от осцила
- Информация на UI Hoffset & TimeBase
- Курсор движение мышкой
*/
