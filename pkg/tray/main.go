package tray

import (
	"github.com/getlantern/systray"
	"github.com/pioruner/HardWorker.git/pkg/app"
	"github.com/pioruner/HardWorker.git/pkg/logger"
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
		logger.Infof("Tray disabled on macOS mode")
		return
	}
	TD = &TrayData{
		icon: icon,
	}

	go systray.Run(TD.onReady, TD.onExit)
	logger.Infof("Tray event loop started")
}

func (a *TrayData) onReady() {
	//Настройка
	systray.SetTitle(app.AppName)
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
				logger.Infof("Tray toggle clicked")
				if app.State.Gui {
					app.CloseGUI <- struct{}{}
				} else {
					app.Event <- app.EventToggleGUI
				}
			case <-TD.mQuit.ClickedCh:
				logger.Warnf("Tray quit clicked")
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
