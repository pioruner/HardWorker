package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
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

//go:embed assets/inter.ttf
var fontI []byte

var (
	macOS        bool
	macMult      float32 = 1.6
	toggleWindow         = make(chan bool, 1)
	quitApp              = make(chan bool, 1) // Канал для сигнала выхода
	commandInput string                       // Текущая команда для ввода
	lastResponse string                       // Последний ответ прибора

	linedata []int8 = nil
)

func runGUIWindow() {
	window := giu.NewMasterWindow("HardWorker", 1000, 800, 0) // Create main window
	img, _, err := image.Decode(bytes.NewReader(iconApp))     //Decode icon
	if err == nil {
		window.SetIcon(img) //Set icon
	}

	font := giu.Context.FontAtlas.AddFontFromBytes("inter", fontI, 14)
	style := giu.Style()
	style.SetFont(font)
	window.SetStyle(style)
	_, h := giu.GetFramePadding()
	window.Run(func() {
		select {
		case <-quitApp: //Close UI
			window.SetShouldClose(true)
			quitApp <- true //Exit main cycle
			return

		case <-toggleWindow: //Hide to tray
			window.SetShouldClose(true) //Drop UI
			return

		default:
			giu.SingleWindow().Layout( //Main UI
				giu.Style().SetFontSize(10).To(
					giu.Plot("AKIP Graph").Size(-10, 600).AxisLimits(0, 100, -150, 150, giu.ConditionOnce).Plots(
						giu.Line("", UtoF(linedata)),
					)),

				giu.Align(giu.AlignCenter).To(
					//giu.Style().SetFontSize(24).To(giu.Label("АКИП")), //Main Lable
					giu.Label("АКИП"),
				),
				giu.Separator(),
				giu.Child().Size(giu.Auto, (14+(h*2)+2)*macMult).Border(false).Layout(
					giu.Row(
						//giu.Style().SetFontSize(20).To(giu.InputText(&commandInput).Size(-200).Hint("Введите SCPI команду...")), //CMD for send
						giu.InputText(&commandInput).Size(-200).Hint("Введите SCPI команду..."),
						giu.Button("Отправить").Size(190, giu.Auto).OnClick(sendCMD), //Send CMD
					)),

				giu.Dummy(0, 10),

				giu.Label("Последний ответ прибора:"),
				giu.InputTextMultiline(&lastResponse).Size(-1, -1).Flags(giu.InputTextFlagsReadOnly), //Response for CMD

				giu.Dummy(0, 10),
			)
		}
	})
}

func sendCMD() {
	resp := akip.CMD("192.168.1.70:44331", commandInput)
	if len(resp) > 0 {
		if commandInput == "STARTBIN" {
			lastResponse = fmt.Sprintf("%X", resp[:len(resp)-2])
			linedata = BtoI(resp[:1000])
		} else {
			cleanResp := bytes.ReplaceAll(resp, []byte{0}, []byte{})
			lastResponse = fmt.Sprintf("%s", cleanResp[:len(cleanResp)-2])
		}
		giu.Update()
	} else {
		lastResponse = "nil"
	}
}

func main() {
	macOS = runtime.GOOS == "darwin" //Check OS

	if !macOS {
		tray.Tray(toggleWindow, quitApp, iconTray) //Create tray icon
		macMult = 1
	}

	toggleWindow <- true
	for { //Main Cycle
		select {
		case <-toggleWindow: //Recall UI
			runGUIWindow()
			if macOS {
				quitApp <- true
			}
		case <-quitApp: //Quit App
			os.Exit(0)
		case <-time.After(100 * time.Millisecond): //Pause
		}
	}
}

func UtoF(data []int8) []float64 {
	result := make([]float64, len(data))
	for i, v := range data {
		result[i] = float64(v)
	}
	return result
}

func BtoI(data []byte) []int8 {
	result := make([]int8, len(data))
	for i, v := range data {
		result[i] = int8(v)
	}
	return result
}
