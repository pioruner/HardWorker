package ui

import (
	"bytes"
	"image"

	"github.com/AllenDang/giu"
	"github.com/pioruner/HardWorker.git/pkg/app"
	"github.com/pioruner/HardWorker.git/pkg/logger"
)

func close() bool {
	if app.MacOS {
		logger.Infof("GUI close callback triggered on macOS")
		app.Event <- app.EventQuit
	}
	return true
}

func GUI(iconApp []byte, fontI []byte, modules ...app.Modules) {

	if len(modules) == 0 {
		logger.Warnf("GUI start skipped: no modules provided")
		return
	}

	app.State.Gui = true
	logger.Infof("GUI opened")

	window := giu.NewMasterWindow(app.AppName, 1000, 450, giu.MasterWindowFlagsMaximized)

	img, _, err := image.Decode(bytes.NewReader(iconApp))
	if err == nil {
		window.SetIcon(img)
	} else {
		logger.Warnf("GUI icon decode failed: %v", err)
	}

	window.SetCloseCallback(close)

	font := giu.Context.FontAtlas.AddFontFromBytes("inter", fontI, 14)
	style := giu.Style()
	style.SetFont(font)
	window.SetStyle(style)

	menu := 0
	_, FPy := giu.GetFramePadding()

	window.Run(func() {

		select {
		case <-app.CloseGUI:
			window.SetShouldClose(true)
			logger.Infof("GUI close signal received")
			return
		default:
		}

		if menu >= len(modules) {
			menu = 0
		}

		// ---------- Формируем меню ----------
		menuLayout := giu.Layout{}

		for i, m := range modules {

			index := i

			menuLayout = append(menuLayout,
				giu.Selectable(m.Name()).
					Selected(menu == index).
					Size(220, (14+(FPy*2)+2)).
					OnClick(func() {
						menu = index
					}),
				giu.Dummy(10, (14+(FPy*2)+2)),
			)
		}

		// ---------- Основной layout ----------
		giu.SingleWindow().Layout(

			giu.Column(

				// ===== Меню =====
				giu.Child().
					Size(-1, (14+(FPy*2)+20)).
					Border(true).
					Layout(
						giu.Align(giu.AlignLeft).To(giu.Row(menuLayout...))),

				// ===== Основная программа =====
				giu.Child().
					Size(-1, -1).
					Border(true).
					Layout(
						modules[menu],
					),
			),
		)
	})

	app.State.Gui = false
	logger.Infof("GUI closed")
}
