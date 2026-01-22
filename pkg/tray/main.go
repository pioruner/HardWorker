package tray

import (
	"context"
	"log"
	"time"

	"github.com/getlantern/systray"
)

// Состояние приложения
type AppState struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var appState *AppState

func Tray(macOS bool, icon []byte) {
	if !macOS {
		// Инициализируем состояние приложения
		ctx, cancel := context.WithCancel(context.Background())
		appState = &AppState{
			ctx:    ctx,
			cancel: cancel,
		}
		go systray.Run(appState.onReady, appState.onExit)
		log.Println("Запущен системный трей для Windows")
	}
}

func (a *AppState) onReady() {
	// Устанавливаем иконку в системный трей
	systray.SetIcon(iconData)
	log.Println("Иконка установлена в системный трей")

	// Заголовок и тултип
	systray.SetTitle("HardWorker")
	systray.SetTooltip("Управление устройствами")

	mToggle := systray.AddMenuItem("Показать/скрыть окно", "Переключить видимость окна")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Выход", "Завершить программу")

	// Обработка событий в фоне
	go func() {
		for {
			select {
			case <-mToggle.ClickedCh:
				windowVisible = !windowVisible

				if windowVisible {
					log.Println("Окно показано (из трея)")
					// Отправляем сигнал создать новое окно
					select {
					case needNewWindow <- true:
					default:
					}
				} else {
					log.Println("Окно скрыто (из трея)")
				}

			case <-mQuit.ClickedCh:
				log.Println("Выход из программы (из трея)")
				// Отправляем сигнал выхода
				select {
				case quitApp <- true:
				default:
				}
				return

			case <-a.ctx.Done():
				return
			}
		}
	}()
}

func (a *AppState) onExit() {
	log.Println("Завершение работы приложения...")

	// Отменяем контекст для остановки всех горутин
	a.cancel()

	// Имитация очистки ресурсов
	log.Println("Закрытие соединений с устройствами...")
	time.Sleep(300 * time.Millisecond)

	log.Println("Приложение завершено")
}
