package tray

import (
	"log"

	"github.com/getlantern/systray"
)

// Состояние приложения
type TrayData struct {
	windowVisibility chan bool // Канал для управления окном
	quitApp          chan bool // Канал для сигнала выхода
	mToggle          *systray.MenuItem
	mQuit            *systray.MenuItem
	icon             []byte
}

var TD *TrayData

func Tray(macOS bool, wV chan bool, qA chan bool, icon []byte) {
	if macOS {
		return
	}
	TD = &TrayData{
		windowVisibility: wV,
		quitApp:          qA,
		icon:             icon,
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
				log.Println("Показать/Скрыть UI")
				select {
				case TD.windowVisibility <- true:
				default:
				}

			case <-TD.mQuit.ClickedCh:
				log.Println("Выход из программы")
				systray.Quit()
				return
			}
		}
	}()
}

func (a *TrayData) onExit() {
	log.Println("Завершение работы приложения...")
	select {
	case TD.quitApp <- true:
	default:
	}
	log.Println("Приложение завершено")
}
