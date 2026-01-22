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
	"github.com/pioruner/HardWorker.git/pkg/akip"
	"github.com/pioruner/HardWorker.git/pkg/tray"
)

//go:embed assets/icon.ico
var iconTray []byte

//go:embed assets/icon.png
var iconApp []byte

var (
	testAkip     = true
	macOS        bool
	toggleWindow = make(chan bool, 1)
	quitApp      = make(chan bool, 1) // Канал для сигнала выхода
	commandInput string               // Текущая команда для ввода
	lastResponse string               // Последний ответ прибора
)

// Функция для создания и запуска окна
func runGUIWindow() {
	window := giu.NewMasterWindow("HardWorker", 600, 250, 0) // Create main window

	img, _, err := image.Decode(bytes.NewReader(iconApp)) //Decode icon
	if err == nil {
		window.SetIcon(img) //Set icon
	}

	window.SetCloseCallback(func() bool { //Close callback
		log.Println("Закрытие UI")
		if macOS {
			quitApp <- true //Mac must close app - no tray control
		}
		return true
	})

	window.Run(func() {
		select {
		case <-quitApp: //Close UI
			log.Println("Получен сигнал выхода, завершение приложения...")
			window.SetShouldClose(true)
			quitApp <- true //Exit main cycle
			return

		case <-toggleWindow: //Hide to tray
			log.Println("Получен сигнал трея, прячем UI...")
			window.SetShouldClose(true) //Drop UI
			return

		default:
			giu.SingleWindow().Layout( //Main UI
				giu.Align(giu.AlignCenter).To(
					giu.Style().SetFontSize(24).To(giu.Label("АКИП")), //Main Lable
				),
				giu.Separator(),

				giu.Row(
					giu.Style().SetFontSize(20).To(
						giu.InputText(&commandInput).Size(-200).Hint("Введите SCPI команду...")), //CMD for send
					giu.Button("Отправить").Size(190, 26), //Send CMD
				),

				giu.Dummy(0, 10),

				giu.Label("Последний ответ прибора:"),
				giu.InputTextMultiline(&lastResponse). //Response for CMD
									Size(-1, 130).
									Flags(giu.InputTextFlagsReadOnly),

				giu.Dummy(0, 10),
			)
		}
	})
}

func main() {
	if testAkip {
		println(string(akip.CMD("192.168.1.70:44331", "STARTBIN")))
	} else {
		macOS = runtime.GOOS == "darwin" //Check OS

		tray.Tray(macOS, toggleWindow, quitApp, iconTray) //Create tray icon

		runGUIWindow() //UI

		for { //Main Cycle
			select {
			case <-toggleWindow: //Recall UI
				log.Println("Запуск GUI окна...")
				runGUIWindow()
				log.Println("GUI окно закрыто")
			case <-quitApp: //Quit App
				log.Println("Завершение программы...")
				os.Exit(0)
			case <-time.After(100 * time.Millisecond): //Pause
			}
		}
	}
}
