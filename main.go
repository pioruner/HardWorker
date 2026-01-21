package main

import (
	"bytes"
	"context"
	_ "embed"
	"image"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/AllenDang/giu"
	"github.com/getlantern/systray"
)

// Импортируем иконку в PNG формате
//
//go:embed assets/icon.ico
var iconData []byte

//go:embed assets/icon.png
var iconPNG []byte

var (
	macOS          bool
	windowVisible  = true
	mainWindow     *giu.MasterWindow
	showWindow     = make(chan bool, 1)
	commandInput   string               // Текущая команда для ввода
	lastResponse   string               // Последний ответ прибора
	commandHistory []string             // История команд
	quitApp        = make(chan bool, 1) // Канал для сигнала выхода
	needNewWindow  = make(chan bool, 1) // Канал для создания нового окна
)

// Состояние приложения
type AppState struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var appState *AppState

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

// Функция для создания и запуска окна
func runGUIWindow() {
	time.Sleep(100 * time.Millisecond)
	// Создаем окно
	window := giu.NewMasterWindow("HardWorker", 600, 250, 0)
	mainWindow = window // Сохраняем ссылку на главное окно

	// Устанавливаем иконку окна (простой способ)
	img, _, err := image.Decode(bytes.NewReader(iconPNG))
	if err == nil {
		window.SetIcon(img)
	}

	// Устанавливаем обработчик закрытия окна
	window.SetCloseCallback(func() bool {
		// При попытке закрыть окно - скрываем его в трей
		windowVisible = false
		log.Println("Окно скрыто в трей (закрытие на крестик)")
		// Закрываем окно
		window.SetShouldClose(true)
		if macOS {
			quitApp <- true
		}
		return true // Предотвращаем стандартное закрытие
	})

	// Запускаем GUI с обработкой внешних событий
	window.Run(func() {
		// Проверяем сигнал выхода из приложения
		select {
		case <-quitApp:
			log.Println("Получен сигнал выхода, завершение приложения...")
			if !macOS {
				systray.Quit()
			}
			os.Exit(0)
		default:
		}

		// Проверяем видимость окна
		if !windowVisible {
			// Если окно должно быть скрыто, закрываем его
			window.SetShouldClose(true)
			return
		}

		// Основной интерфейс
		giu.SingleWindow().Layout(
			// Шапка - лейбл АКИП
			giu.Align(giu.AlignCenter).To(
				giu.Style().SetFontSize(24).To(giu.Label("АКИП")),
			),
			giu.Separator(),

			// Строка ввода команды
			giu.Row(
				giu.Style().SetFontSize(20).To(
					giu.InputText(&commandInput).Size(-200).Hint("Введите SCPI команду...")),
				giu.Button("Отправить").Size(190, 26),
			),

			giu.Dummy(0, 10),

			// Подвал: текстовое поле для ответов прибора
			giu.Label("Последний ответ прибора:"),
			giu.InputTextMultiline(&lastResponse).
				Size(-1, 130).
				Flags(giu.InputTextFlagsReadOnly),

			giu.Dummy(0, 10),
		)
	})
}

func main() {
	// На Windows запускаем трей
	macOS = runtime.GOOS == "darwin"

	if !macOS {
		// Инициализируем состояние приложения
		ctx, cancel := context.WithCancel(context.Background())
		appState = &AppState{
			ctx:    ctx,
			cancel: cancel,
		}
		go systray.Run(appState.onReady, appState.onExit)
		log.Println("Запущен системный трей для Windows")
		windowVisible = false
	}

	// Главный цикл управления окнами
	for {
		// Проверяем сигнал выхода
		select {
		case <-quitApp:
			log.Println("Завершение программы...")
			if !macOS {
				systray.Quit()
			}
			os.Exit(0)
		default:
		}

		// Если окно должно быть видимым - запускаем его
		if windowVisible {
			log.Println("Запуск GUI окна...")
			runGUIWindow()
			log.Println("GUI окно закрыто")
		} else {
			// Если окно скрыто, ждем сигнала показа или выхода
			select {
			case <-quitApp:
				log.Println("Завершение программы...")
				if !macOS {
					systray.Quit()
				}
				os.Exit(0)
			case <-needNewWindow:
				windowVisible = true
				log.Println("Получен запрос на создание нового окна")
			case <-time.After(100 * time.Millisecond):
				// Короткая пауза, чтобы не грузить CPU
			}
		}
	}
}
