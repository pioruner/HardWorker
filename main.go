package main

import (
	_ "embed"
	"os"
	"runtime"
	"time"

	"github.com/pioruner/HardWorker.git/pkg/tray"
	"github.com/pioruner/HardWorker.git/pkg/ui"
)

//go:embed assets/icon.ico
var iconTray []byte

//go:embed assets/icon.png
var iconApp []byte

//go:embed assets/inter.ttf
var fontI []byte

var (
	macOS        bool
	toggleWindow = make(chan bool, 1)
	quitApp      = make(chan bool, 1) // Канал для сигнала выхода
)

func init() {
	macOS = runtime.GOOS == "darwin" //Check OS
}

func main() {

	if !macOS {
		tray.Tray(toggleWindow, quitApp, iconTray) //Create tray icon
	}

	toggleWindow <- true
	for { //Main Cycle
		select {
		case <-toggleWindow: //Recall UI
			ui.GUI(iconApp, fontI, toggleWindow, quitApp)
			if macOS {
				quitApp <- true
			}
		case <-quitApp: //Quit App
			os.Exit(0)
		case <-time.After(100 * time.Millisecond): //Pause
		}
	}
}
