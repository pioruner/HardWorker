package tray

import (
	"log"

	"github.com/getlantern/systray"
	"github.com/pioruner/HardWorker.git/pkg/app"
)

// Состояние приложения
type TrayData struct {
	mToggle *systray.MenuItem
	mQuit   *systray.MenuItem
	icon    []byte
}

var TD *TrayData

func Tray(icon []byte) {
	if app.MacOS {
		return
	}
	TD = &TrayData{
		icon: icon,
	}

	go systray.Run(TD.onReady, TD.onExit)
	log.Println("Запущен системный трей для Windows")
}

func (a *TrayData) onReady() {
	//Настройка
	systray.SetTitle("HardWorker")
	systray.SetTooltip("Управление устройствами")
	//Меню
	TD.mToggle = systray.AddMenuItem("Показать/скрыть окно", "Переключить видимость окна")
	systray.AddSeparator()
	TD.mQuit = systray.AddMenuItem("Выход", "Завершить программу")
	//Icon
	systray.SetIcon(TD.icon)
	// Обработка событий
	go func() {
		for {
			select {
			case <-TD.mToggle.ClickedCh:
				if app.State.Gui {
					app.CloseGUI <- struct{}{}
				} else {
					app.Event <- app.EventToggleGUI
				}
			case <-TD.mQuit.ClickedCh:
				if app.State.Gui {
					app.CloseGUI <- struct{}{}
				}
				systray.Quit()
				app.Event <- app.EventQuit
				return
			}
		}
	}()
}

func (a *TrayData) onExit() {}
