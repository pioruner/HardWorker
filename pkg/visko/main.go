package visko

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"path/filepath"

	"github.com/AllenDang/giu"
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

// / NEW CODE !!!
type TableRow struct {
	T1   float64
	T2   float64
	U1   float64
	U2   float64
	Temp float64
}

type NewModuleUI struct {
	rows []TableRow

	// текущие (живые) параметры
	curT1, curT2, curU1, curU2, curTemp string

	// параметры по курсору
	selT1, selT2, selU1, selU2, selTemp string

	cursorIndex int32
}

/*
	func Init(adr string, name string) *NewModuleUI {
		return &NewModuleUI{}
	}
*/
func Init(adr string, name string) *NewModuleUI {

	ui := &NewModuleUI{}

	baseTemp := 22.0

	for i := 0; i < 20; i++ {

		x := float64(i)

		// Время (две близкие кривые)
		t1 := 100 + 10*math.Sin(x*0.3)
		t2 := 102 + 10*math.Sin(x*0.3+0.2)

		// Напряжение (две близкие кривые)
		u1 := 5 + 0.5*math.Sin(x*0.5)
		u2 := 5.1 + 0.5*math.Sin(x*0.5+0.3)

		// Температура растёт на 1 градус
		temp := baseTemp + float64(i)

		ui.rows = append(ui.rows, TableRow{
			T1:   t1,
			T2:   t2,
			U1:   u1,
			U2:   u2,
			Temp: temp,
		})
	}

	// Инициализируем курсор
	ui.cursorIndex = int32(len(ui.rows) - 1)
	ui.updateCursorValues()

	// Текущие параметры = последняя точка
	last := ui.rows[len(ui.rows)-1]
	ui.curT1 = fmt.Sprintf("%.3f", last.T1)
	ui.curT2 = fmt.Sprintf("%.3f", last.T2)
	ui.curU1 = fmt.Sprintf("%.3f", last.U1)
	ui.curU2 = fmt.Sprintf("%.3f", last.U2)
	ui.curTemp = fmt.Sprintf("%.3f", last.Temp)

	return ui
}

func (ui *NewModuleUI) Run() {
	//app.Wg.Add(1)
	//go ui.connectionLoop()
	//log.Printf("Module Visko with name: %s --STARTED", ui.id)
}

// Save report
func (ui *NewModuleUI) SaveCSV() {
	file, err := os.Create("data.csv")
	if err != nil {
		return
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	w.Write([]string{"T1", "T2", "U1", "U2", "Temp"})

	for _, r := range ui.rows {
		w.Write([]string{
			fmt.Sprintf("%f", r.T1),
			fmt.Sprintf("%f", r.T2),
			fmt.Sprintf("%f", r.U1),
			fmt.Sprintf("%f", r.U2),
			fmt.Sprintf("%f", r.Temp),
		})
	}
}

func (ui *NewModuleUI) buildPlots() (timePlots, voltagePlots, tempPlots []giu.PlotWidget) {

	var x []float64
	var t1, t2, u1, u2, temp []float64

	for i, r := range ui.rows {
		x = append(x, float64(i))
		t1 = append(t1, r.T1)
		t2 = append(t2, r.T2)
		u1 = append(u1, r.U1)
		u2 = append(u2, r.U2)
		temp = append(temp, r.Temp)
	}

	timePlots = []giu.PlotWidget{
		giu.LineXY("T1", x, t1),
		giu.LineXY("T2", x, t2),
	}

	voltagePlots = []giu.PlotWidget{
		giu.LineXY("U1", x, u1),
		giu.LineXY("U2", x, u2),
	}

	tempPlots = []giu.PlotWidget{
		giu.LineXY("Temp", x, temp),
	}

	return
}

func (ui *NewModuleUI) buildTable() []*giu.TableRowWidget {

	var rows []*giu.TableRowWidget

	for _, r := range ui.rows {
		row := giu.TableRow(
			giu.Label(fmt.Sprintf("%.3f", r.T1)),
			giu.Label(fmt.Sprintf("%.3f", r.T2)),
			giu.Label(fmt.Sprintf("%.3f", r.U1)),
			giu.Label(fmt.Sprintf("%.3f", r.U2)),
			giu.Label(fmt.Sprintf("%.3f", r.Temp)),
		)
		rows = append(rows, row)
	}

	return rows
}

func (ui *NewModuleUI) updateCursorValues() {
	if len(ui.rows) == 0 {
		return
	}

	if ui.cursorIndex >= int32(len(ui.rows)) {
		ui.cursorIndex = int32(len(ui.rows))
	}
	if ui.cursorIndex < 0 {
		ui.cursorIndex = 0
	}

	r := ui.rows[ui.cursorIndex]

	ui.selT1 = fmt.Sprintf("%.3f", r.T1)
	ui.selT2 = fmt.Sprintf("%.3f", r.T2)
	ui.selU1 = fmt.Sprintf("%.3f", r.U1)
	ui.selU2 = fmt.Sprintf("%.3f", r.U2)
	ui.selTemp = fmt.Sprintf("%.3f", r.Temp)
}

func drawCursorLine(index int, yMin, yMax float64, name string) giu.PlotWidget {
	x := float64(index)
	return giu.LineXY(
		name,
		[]float64{x, x},
		[]float64{yMin, yMax},
	)
}

func getMinMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 1
	}

	min := values[0]
	max := values[0]

	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// небольшой отступ сверху/снизу
	padding := (max - min) * 0.1
	if padding == 0 {
		padding = 1
	}

	return min - padding, max + padding
}
