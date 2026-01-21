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
	macOS          bool
	windowVisible  = true
	mainWindow     *giu.MasterWindow
	showWindow     = make(chan bool, 1)
	commandInput   string   // Текущая команда для ввода
	lastResponse   string   // Последний ответ прибора
	commandHistory []string // История команд
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
	}

	// Создаем окно
	window := giu.NewMasterWindow("HardWorker", 600, 250, 0)

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
					if !macOS {
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
			// Шапка - лейбл АКИП
			giu.Align(giu.AlignCenter).To(
				giu.Style().SetFontSize(24).To(giu.Label("АКИП")),
			),
			giu.Separator(),

			// Строка ввода команды - ИСПРАВЛЕНО
			giu.Row(
				giu.Style().SetFontSize(20).To(
					giu.InputText(&commandInput).Size(-200).Hint("Введите SCPI команду...")), // Занимает всё кроме 200px
				giu.Button("Отправить").Size(190, 26), // Фиксированная ширина
			),

			giu.Dummy(0, 10),

			// Подвал: текстовое поле для ответов прибора
			giu.Label("Последний ответ прибора:"),
			giu.InputTextMultiline(&lastResponse).
				Size(-1, 130).
				Flags(giu.InputTextFlagsReadOnly),
			//BGColor(giu.Vec4{0.95, 0.95, 0.95, 1.0}), // Серый фон для read-only

			giu.Dummy(0, 10))
	})

}
