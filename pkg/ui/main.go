package ui

import (
	"bytes"
	"image"

	"github.com/AllenDang/giu"
	"github.com/pioruner/HardWorker.git/pkg/akip"
	"github.com/pioruner/HardWorker.git/pkg/app"
)

var (
	akiper   *akip.AkipW
	new_akip *akip.AkipUI
)

func init() {
	akiper = akip.New("192.168.0.100:3000")
	new_akip = akip.Init()
}

func close() bool {
	if app.MacOS {
		app.Event <- app.EventQuit
	}
	return true
}

func GUI(iconApp []byte, fontI []byte) {
	app.State.Gui = true
	window := giu.NewMasterWindow("HardWorker", 1000, 450, giu.MasterWindowFlagsMaximized) // Create main window. giu.MasterWindowFlagsMaximized
	img, _, err := image.Decode(bytes.NewReader(iconApp))                                  //Decode icon
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
				//akiper,
				new_akip,
			)
		}
	})
	app.State.Gui = false
}
