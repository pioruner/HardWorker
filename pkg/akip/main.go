package akip

import (
	"context"
	"net"
	"sync"
)

func Init(adr string, name string, ctx context.Context, wg *sync.WaitGroup) *AkipUI {
	return &AkipUI{
		adr:    adr,
		id:     name,
		cmdCh:  make(chan SCPICommand, 8),
		X:      []float64{0, 1, 2, 3},
		Y:      []float64{1, 1, 1, 1},
		xsize:  3,
		ctx:    ctx,
		wg:     wg,
		update: false,
	}
}

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

func (ui *AkipUI) Run() {
	ui.Load()
	ui.wg.Add(1)
	go ui.connectionLoop()
	ui.Save()
	ui.wg.Done()
}

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
	cursorPos  [3]int32 // позиции курсоров (в индексах)

	connected bool
	conn      net.Conn
	cmdCh     chan SCPICommand

	xdt, xhoffs float64
	xsize       int

	ctx    context.Context
	wg     *sync.WaitGroup
	update bool
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
	CursorPos  [3]int32
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
