package visko

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/AllenDang/giu"
	"github.com/pioruner/HardWorker.git/pkg/app"
	"github.com/pioruner/HardWorker.git/pkg/logger"
	"github.com/simonvetter/modbus"
	"github.com/sqweek/dialog"
)

type TableRow struct {
	T1   float64
	T2   float64
	U1   float64
	U2   float64
	Temp float64
}

type ViskoUI struct {
	rows []TableRow

	// текущие (живые) параметры
	curT1, curT2, curU1, curU2, curTemp string

	// параметры по курсору
	selT1, selT2, selU1, selU2, selTemp string

	cursorIndex int32
	update      bool
	id          string
	adr         string
	conn        *modbus.ModbusClient
	connected   bool
	cmd         uint16
}

func Init(adr string, name string) *ViskoUI {
	return &ViskoUI{
		id:  name,
		adr: adr,
	}
}

/*
func Init(adr string, name string) *ViskoUI {

		ui := &ViskoUI{
			id:  name,
			adr: adr,
		}

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
		ui.curT1 = fmt.Sprintf("%.1f", last.T1)
		ui.curT2 = fmt.Sprintf("%.1f", last.T2)
		ui.curU1 = fmt.Sprintf("%.2f", last.U1)
		ui.curU2 = fmt.Sprintf("%.2f", last.U2)
		ui.curTemp = fmt.Sprintf("%.1f", last.Temp)

		return ui
	}
*/
func (ui *ViskoUI) Run() {
	app.Wg.Add(1)
	go ui.connectionLoop()
	logger.Infof("Module Visko started: %s", ui.id)
}

func (ui *ViskoUI) Name() string {
	return ui.id
}

func (ui *ViskoUI) buildPlots() (
	timePlots []giu.PlotWidget,
	voltagePlots []giu.PlotWidget,
	tempPlots []giu.PlotWidget,
	xMin, xMax float64,
	timeYMin, timeYMax float64,
	voltYMin, voltYMax float64,
	tempYMin, tempYMax float64,
) {

	n := len(ui.rows)
	if n == 0 {
		return
	}

	x := make([]float64, n)
	t1 := make([]float64, n)
	t2 := make([]float64, n)
	u1 := make([]float64, n)
	u2 := make([]float64, n)
	temp := make([]float64, n)

	for i, r := range ui.rows {
		x[i] = float64(i)
		t1[i] = r.T1
		t2[i] = r.T2
		u1[i] = r.U1
		u2[i] = r.U2
		temp[i] = r.Temp
	}

	xMin = 0
	xMax = float64(n - 1)

	timeYMin, timeYMax = getMinMax(append(t1, t2...))
	voltYMin, voltYMax = getMinMax(append(u1, u2...))
	tempYMin, tempYMax = getMinMax(temp)

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

func (ui *ViskoUI) buildTable() []*giu.TableRowWidget {

	var rows []*giu.TableRowWidget

	for _, r := range ui.rows {
		row := giu.TableRow(
			giu.Label(fmt.Sprintf("%.1f", r.T1)),
			giu.Label(fmt.Sprintf("%.1f", r.T2)),
			giu.Label(fmt.Sprintf("%.2f", r.U1)),
			giu.Label(fmt.Sprintf("%.2f", r.U2)),
			giu.Label(fmt.Sprintf("%.1f", r.Temp)),
		)
		rows = append(rows, row)
	}

	return rows
}

func (ui *ViskoUI) updateCursorValues() {
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

	ui.selT1 = fmt.Sprintf("%.1f", r.T1)
	ui.selT2 = fmt.Sprintf("%.1f", r.T2)
	ui.selU1 = fmt.Sprintf("%.2f", r.U1)
	ui.selU2 = fmt.Sprintf("%.2f", r.U2)
	ui.selTemp = fmt.Sprintf("%.1f", r.Temp)
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

func (ui *ViskoUI) AddRow(t1, t2, u1, u2, temp float64) {

	ui.rows = append(ui.rows, TableRow{
		T1:   t1,
		T2:   t2,
		U1:   u1,
		U2:   u2,
		Temp: temp,
	})
}

func (ui *ViskoUI) setUpdate() {
	ui.update = true
}

func (ui *ViskoUI) SaveCSVDialog() {

	path, err := dialog.File().
		Filter("CSV file", "csv").
		Title("Сохранить таблицу").
		Save()

	if err != nil {
		logger.Infof("CSV save dialog cancelled: %s", ui.id)
		return // пользователь отменил
	}
	path = path + ".csv"
	file, err := os.Create(path)
	if err != nil {
		logger.Errorf("CSV file create failed (%s): %v", path, err)
		return
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	if err := w.Write([]string{"T1", "T2", "U1", "U2", "Temp"}); err != nil {
		logger.Errorf("CSV header write failed (%s): %v", path, err)
		return
	}

	for _, r := range ui.rows {
		if err := w.Write([]string{
			fmt.Sprintf("%.0f", r.T1),
			fmt.Sprintf("%.0f", r.T2),
			fmt.Sprintf("%.3f", r.U1),
			fmt.Sprintf("%.3f", r.U2),
			fmt.Sprintf("%.2f", r.Temp),
		}); err != nil {
			logger.Errorf("CSV row write failed (%s): %v", path, err)
			return
		}
	}
	logger.Infof("CSV saved: %s (%d rows)", path, len(ui.rows))
}
