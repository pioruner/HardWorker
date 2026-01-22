package main

import (
	"bytes"
	_ "embed"
	"image"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/AllenDang/giu"
	"github.com/pioruner/HardWorker.git/pkg/tray"
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
	toggleWindow   = make(chan bool, 1)
	quitApp        = make(chan bool, 1) // Канал для сигнала выхода
	commandInput   string               // Текущая команда для ввода
	lastResponse   string               // Последний ответ прибора
	commandHistory []string             // История команд
)

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
		//windowVisible = false
		log.Println("Закрытие на крестик")
		// Закрываем окно
		//window.SetShouldClose(true)
		if macOS {
			quitApp <- true
		}
		return true
	})

	// Запускаем GUI с обработкой внешних событий
	window.Run(func() {
		// Проверяем сигнал выхода из приложения
		select {
		case <-quitApp:
			log.Println("Получен сигнал выхода, завершение приложения...")
			window.SetShouldClose(true)
			//os.Exit(0)
			return
		case <-toggleWindow:
			log.Println("Получен сигнал трея, прячем UI...")
			window.SetShouldClose(true)
			return
		default:
		}

		// Проверяем видимость окна
		//if !windowVisible {
		// Если окно должно быть скрыто, закрываем его
		//	window.SetShouldClose(true)
		//	return
		//}

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
	macOS = runtime.GOOS == "darwin"

	tray.Tray(macOS, toggleWindow, quitApp, iconData)

	runGUIWindow()
	// Главный цикл управления окнами
	for {
		// Проверяем сигнал выхода
		select {
		case <-toggleWindow:
			log.Println("Запуск GUI окна...")
			runGUIWindow()
			log.Println("GUI окно закрыто")
		case <-quitApp:
			log.Println("Завершение программы...")
			os.Exit(0)
		case <-time.After(100 * time.Millisecond):
			//pause
		}
	}
}
