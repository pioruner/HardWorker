package ui

import (
	"bytes"
	"image"

	"github.com/AllenDang/giu"
	"github.com/pioruner/HardWorker.git/pkg/app"
)

func close() bool {
	if app.MacOS {
		app.Event <- app.EventQuit
	}
	return true
}

func GUI(iconApp []byte, fontI []byte, w ...giu.Widget) {
	app.State.Gui = true
	window := giu.NewMasterWindow("HardWorker", 1000, 450, 0) // Create main window. giu.MasterWindowFlagsMaximized
	img, _, err := image.Decode(bytes.NewReader(iconApp))     //Decode icon
	if err == nil {
		window.SetIcon(img) //Set icon
	}

	window.SetCloseCallback(close)

	font := giu.Context.FontAtlas.AddFontFromBytes("inter", fontI, 14)
	style := giu.Style()
	style.SetFont(font)
	window.SetStyle(style)
	window.Run(func() {
		select {
		case <-app.CloseGUI: //Hide to tray
			window.SetShouldClose(true) //Drop UI
			return

		default:
			giu.SingleWindow().Layout( //Main UI
				w...,
			)
		}
	})
	app.State.Gui = false
}
