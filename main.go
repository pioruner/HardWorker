package main

import (
	"context"
	_ "embed"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/AllenDang/giu"
	"github.com/getlantern/systray"
)

// Импортируем иконку с помощью go:embed
//
//go:embed assets/icon.png
var iconData []byte

var (
	windowVisible = true
	mainWindow    *giu.MasterWindow
	showWindow    = make(chan bool, 1)
)

// Состояние приложения
type AppState struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var appState *AppState

func (a *AppState) onReady() {
	// Устанавливаем иконку
	if len(iconData) > 0 {
		systray.SetIcon(iconData)
	} else {
		log.Println("Иконка не найдена, будет использована стандартная")
	}

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
				// Переключаем видимость окна
				windowVisible = !windowVisible

				if windowVisible {
					log.Println("Окно показано")
					// Отправляем сигнал показать окно
					select {
					case showWindow <- true:
					default:
					}
				} else {
					log.Println("Окно скрыто")
				}

			case <-mQuit.ClickedCh:
				log.Println("Выход из программы")
				systray.Quit()
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

func main() {
	// Инициализируем состояние приложения
	ctx, cancel := context.WithCancel(context.Background())
	appState = &AppState{
		ctx:    ctx,
		cancel: cancel,
	}
	// На Windows запускаем трей
	if runtime.GOOS == "windows" {
		go systray.Run(appState.onReady, appState.onExit)
		log.Println("Запущен системный трей для Windows")
	} else {
		log.Println("macOS/Linux: системный трей отключен")
	}

	// Создаем окно
	window := giu.NewMasterWindow("HardWorker", 400, 300, 0)

	// Устанавливаем обработчик закрытия окна
	window.SetCloseCallback(func() bool {
		// При попытке закрыть окно - просто скрываем его
		windowVisible = false
		log.Println("Окно скрыто (закрытие)")
		return true // Предотвращаем реальное закрытие
	})

	// Запускаем GUI
	window.Run(func() {
		// Проверяем видимость окна
		visible := windowVisible

		// Если окно скрыто - не рисуем его
		if !visible {
			return
		}

		// Главное меню приложения
		giu.MainMenuBar().Layout(
			giu.Menu("Файл").Layout(
				giu.MenuItem("Скрыть в трей").OnClick(func() {
					windowVisible = false
					log.Println("Окно скрыто из меню")
				}),
				giu.MenuItem("Выход").OnClick(func() {
					log.Println("Выход из программы (GUI)")
					if runtime.GOOS == "windows" {
						systray.Quit()
					}
					os.Exit(0)
				}),
			),
			giu.Menu("Вид").Layout(
				giu.MenuItem("Обновить").OnClick(func() {
					log.Println("Обновление интерфейса...")
				}),
			),
		).Build()

		// Основной интерфейс
		giu.SingleWindow().Layout(
			giu.Label("HardWorker"),
			giu.Separator(),
			giu.Label("Управление подключениями к устройствам"),

			giu.Dummy(0, 20),

			giu.Button("Переподключить устройства").Size(200, 40).OnClick(func() {
				log.Println("Переподключение из кнопки...")
			}),

			giu.Dummy(0, 10),

			giu.Row(
				giu.Button("Скрыть в трей").Size(120, 30).OnClick(func() {
					windowVisible = false
					log.Println("Окно скрыто из кнопки")
				}),

				giu.Button("Выход").Size(120, 30).OnClick(func() {
					log.Println("Выход из программы (кнопка)")
					if runtime.GOOS == "windows" {
						systray.Quit()
					}
					os.Exit(0)
				}),
			),

			giu.Separator(),

			giu.Label("Статус:"),
			giu.Label("✓ Приложение активно работает"),
			giu.Label("ОС: "+runtime.GOOS),
		)
	})

}
