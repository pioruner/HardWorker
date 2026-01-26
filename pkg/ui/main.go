package ui

import (
	"bytes"
	"image"

	"github.com/AllenDang/giu"
	"github.com/pioruner/HardWorker.git/pkg/akip"
)

var (
	akiper *akip.AkipW
)

func init() {
	akiper = akip.New("192.168.1.70:44331")
}

func GUI(iconApp []byte, fontI []byte, toggleWindow chan bool, quitApp chan bool) {
	window := giu.NewMasterWindow("HardWorker", 1000, 450, 0) // Create main window. giu.MasterWindowFlagsMaximized
	img, _, err := image.Decode(bytes.NewReader(iconApp))     //Decode icon
	if err == nil {
		window.SetIcon(img) //Set icon
	}

	font := giu.Context.FontAtlas.AddFontFromBytes("inter", fontI, 14)
	style := giu.Style()
	style.SetFont(font)
	window.SetStyle(style)
	window.Run(func() {
		select {
		case <-quitApp: //Close UI
			window.SetShouldClose(true)
			quitApp <- true //Exit main cycle
			return

		case <-toggleWindow: //Hide to tray
			window.SetShouldClose(true) //Drop UI
			return

		default:
			giu.SingleWindow().Layout( //Main UI
				akiper,
			)
		}
	})
}
