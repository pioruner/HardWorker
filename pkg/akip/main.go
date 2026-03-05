package akip

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pioruner/HardWorker.git/pkg/app"
	"github.com/sqweek/dialog"
)

type AkipUI struct {
	adr          string
	id           string
	commandInput string // Текущая команда для ввода
	lastResponse string // Последний ответ прибора
	linedata     []int8
	plotData     []float64

	X, Y []float64

	FPx, FPy, MacMult                                            float32
	Hoffset, reper, square, vspeed, vtime, volume, minY, minMove string
	auto                                                         bool

	timeB int32

	cursorMode CursorMode
	cursorPos  [3]float32 // позиции курсоров (в индексах)

	connected bool
	conn      net.Conn
	cmdCh     chan SCPICommand

	xdt, xhoffs float64
	xsize       int

	update         bool
	Atime, Aoffset string
	gport          string
	regist         bool
	reg_path       string
}

type CursorMode int32

const (
	CursorStart CursorMode = iota
	CursorReper
	CursorFront
)

type SCPICommand struct {
	Cmd string
}

type AkipState struct {
	Adr   string
	TimeB int32
	Auto  bool

	Hoffset string
	Reper   string
	Square  string
	Vspeed  string
	Vtime   string
	Volume  string
	MinY    string
	MinMove string

	CursorMode CursorMode
	CursorPos  [3]float32
}

var TimeScale []float64 = []float64{
	1e-6 * 1,
	1e-6 * 2,
	1e-6 * 5,
	1e-6 * 10,
	1e-6 * 20,
	1e-6 * 50,
	1e-6 * 100,
}

var TimeScaleS []string = []string{
	"1us",
	"2us",
	"5us",
	"10us",
	"20us",
	"50us",
	"100us",
}

var baseOffest []float64 = []float64{
	7.6 * 1,
	7.6 * 2,
	7.6 * 5,
	7.6 * 10,
	7.6 * 20,
	7.6 * 50,
	7.6 * 100,
}

func Init(adr string, name string, gRPCport string) *AkipUI {
	return &AkipUI{
		adr:    adr,
		id:     name,
		cmdCh:  make(chan SCPICommand, 8),
		X:      []float64{0, 1, 2, 3},
		Y:      []float64{1, 1, 1, 1},
		xsize:  3,
		update: false,
		gport:  gRPCport,
		regist: false,
	}
}

func (ui *AkipUI) Run() {
	app.Wg.Add(1)
	go ui.connectionLoop()
	go ui.gRPC()
	go ui.registrationLoop()
	log.Printf("Module Akip with name: %s --STARTED", ui.id)
}

func (ui *AkipUI) Name() string {
	return ui.id
}

// LOAD && SAVE

func (ui *AkipUI) Save() {
	path, err := AppConfigPath(ui.id)
	if err == nil {
		_ = SaveState(path, ui.ExportState())
	}
}

func (ui *AkipUI) Load() {
	path, err := AppConfigPath(ui.id)
	if err == nil {
		if state, err := LoadState(path); err == nil {
			ui.ImportState(state)
		}
	}
}

func AppConfigPath(name string) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "HardWorker")
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, name+".json"), nil
}

func LoadState(path string) (AkipState, error) {
	var state AkipState

	data, err := os.ReadFile(path)
	if err != nil {
		return state, err
	}
	err = json.Unmarshal(data, &state)
	return state, err
}

func SaveState(path string, state AkipState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (ui *AkipUI) ExportState() AkipState {
	return AkipState{
		Adr:        ui.adr,
		TimeB:      ui.timeB,
		Auto:       ui.auto,
		Hoffset:    ui.Hoffset,
		Reper:      ui.reper,
		Square:     ui.square,
		Vspeed:     ui.vspeed,
		Vtime:      ui.vtime,
		Volume:     ui.volume,
		MinY:       ui.minY,
		MinMove:    ui.minMove,
		CursorMode: ui.cursorMode,
		CursorPos:  ui.cursorPos,
	}
}

func (ui *AkipUI) ImportState(s AkipState) {
	ui.adr = s.Adr
	ui.timeB = s.TimeB
	ui.auto = false //s.Auto
	ui.Hoffset = s.Hoffset
	ui.reper = s.Reper
	ui.square = s.Square
	ui.vspeed = s.Vspeed
	ui.vtime = s.Vtime
	ui.volume = s.Volume
	ui.minY = s.MinY
	ui.minMove = s.MinMove
	ui.cursorMode = s.CursorMode
	ui.cursorPos = s.CursorPos
}

func (ui *AkipUI) registrationLoop() {

	type regState int

	const (
		stateIdle regState = iota
		stateRecording
	)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	state := stateIdle

	var file *os.File
	var writer *csv.Writer

	for {
		select {

		case <-app.Ctx.Done():

			// корректное завершение
			if state == stateRecording {
				writer.Flush()
				file.Close()
			}
			return

		case <-ticker.C:

			switch state {

			// =======================
			// IDLE
			// =======================
			case stateIdle:

				if ui.regist {

					path, err := dialog.File().
						Filter("CSV file", "csv").
						Title("Файл регистрации").
						Save()

					if err != nil {
						ui.regist = false
						break
					}

					if !strings.HasSuffix(path, ".csv") {
						path += ".csv"
					}

					// проверяем существует ли файл
					fileExists := false
					if _, err := os.Stat(path); err == nil {
						fileExists = true
					}

					f, err := os.OpenFile(
						path,
						os.O_CREATE|os.O_WRONLY|os.O_APPEND,
						0644,
					)
					if err != nil {
						ui.regist = false
						break
					}

					file = f
					writer = csv.NewWriter(file)

					// если файл новый — пишем заголовок
					if !fileExists {
						_ = writer.Write([]string{"Date-Time", "Volume ml"})
						writer.Flush()
					}

					state = stateRecording
				}

			// =======================
			// RECORDING
			// =======================
			case stateRecording:

				if !ui.regist {

					writer.Flush()
					file.Close()

					writer = nil
					file = nil

					state = stateIdle
					break
				}

				// запись строки
				_ = writer.Write([]string{
					time.Now().Format(time.DateTime),
					ui.volume,
				})

				writer.Flush()
			}
		}
	}
}
