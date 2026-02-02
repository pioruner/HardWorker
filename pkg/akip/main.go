package akip

import (
	"context"
	"net"
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
	cursorPos  [3]int32 // позиции курсоров (в индексах)

	connected bool
	conn      net.Conn
	cmdCh     chan SCPICommand

	xdt, xhoffs float64
	xsize       int

	ctx     context.Context
	wcancel context.CancelFunc
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
