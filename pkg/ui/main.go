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

func GUI(iconApp []byte, fontI []byte, modules ...app.Modules) {

	if len(modules) == 0 {
		return
	}

	app.State.Gui = true

	window := giu.NewMasterWindow(app.AppName, 1000, 450, giu.MasterWindowFlagsMaximized)

	img, _, err := image.Decode(bytes.NewReader(iconApp))
	if err == nil {
		window.SetIcon(img)
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
			)
		}

		// ---------- Основной layout ----------
		giu.SingleWindow().Layout(

			giu.Row(

				// ===== ЛЕВАЯ ПАНЕЛЬ =====
				giu.Child().
					Size(240, -1).
					Border(true).
					Layout(
						giu.Align(giu.AlignCenter).To(giu.Label("МОДУЛИ")),
						giu.Separator(),
						giu.Dummy(220, 5),
						giu.Align(giu.AlignCenter).To(menuLayout...)),

				// ===== ПРАВАЯ ПАНЕЛЬ =====
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
}
