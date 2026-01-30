package akip

import (
	"net"

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
	cursorPos  [3]int32 // позиции курсоров (в индексах)

	connected bool
	conn      net.Conn
}

type CursorMode int32

const (
	CursorStart CursorMode = iota
	CursorReper
	CursorFront
)

func Init(a string) *AkipUI {
	return &AkipUI{
		adr: a,
	}
}

func (ui *AkipUI) Build() {
	for _, w := range ui.UI() {
		w.Build()
	}
}

func (ui *AkipUI) Save() {
	path, err := AppConfigPath()
	if err == nil {
		_ = SaveState(path, ui.ExportState())
	}
}

func (ui *AkipUI) Load() {
	path, err := AppConfigPath()
	if err == nil {
		if state, err := LoadState(path); err == nil {
			ui.ImportState(state)
		}
	}
}

func drawCursor(x float64, yMin, yMax float64) giu.PlotWidget {
	return giu.LineXY(
		"",
		[]float64{x, x},
		[]float64{yMin, yMax},
	)
}

func (ui *AkipUI) UI() giu.Layout {
	ui.FPx, ui.FPy = giu.GetFramePadding()
	ui.MacMult = 1
	plots := []giu.PlotWidget{
		giu.LineXY("", ui.X, []float64{-127, 127}), //ui.Y),
	}

	for i := 0; i < 3; i++ {
		//x := ui.X[ui.cursorPos[i]]
		x := float64(ui.cursorPos[i])
		plots = append(plots, drawCursor(x, -150, 150))
	}
	return giu.Layout{
		giu.Align(giu.AlignCenter).To(
			giu.Style().SetFontSize(16).To(giu.Label("АКИП")), //Main Lable
		),
		giu.Separator(),
		giu.Child().Size(-3, (14+(ui.FPy*2)+2)*ui.MacMult).Border(false).Layout(
			giu.Row(
				//giu.Style().SetFontSize(20).To(giu.InputText(&commandInput).Size(-200).Hint("Введите SCPI команду...")), //CMD for send
				giu.InputText(&ui.adr).Size(150).Hint("Введите IP:Port...").Flags(giu.InputTextFlags(giu.AlignCenter)),
				giu.Button("Подключить").Size(100, giu.Auto).OnClick(func() {}), //Send CMD
				//giu.Spacing(),
				giu.Dummy(30, -1),
				giu.Combo("TimeBase", TimeScaleS[ui.timeB], TimeScaleS, &ui.timeB).Size(100).OnChange(func() {}),
				giu.InputText(&ui.Hoffset).Label("H Offset").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(func() {}),
				giu.Dummy(10, -1),
			)),
		giu.Child().Size(-3, (14+(ui.FPy*2)+2)*ui.MacMult).Border(false).Layout(
			giu.Row(
				//giu.Dummy(50, -1),
				giu.InputText(&ui.reper).Label("dL Reper").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(func() {}),
				giu.InputText(&ui.square).Label("S Square").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(func() {}),
				giu.Dummy(10, -1),
				giu.InputText(&ui.vspeed).Label("Speed").Size(50).Flags(giu.InputTextFlagsReadOnly),
				giu.InputText(&ui.vtime).Label("Time").Size(50).Flags(giu.InputTextFlagsReadOnly),
				giu.InputText(&ui.volume).Label("Volume").Size(50).Flags(giu.InputTextFlagsReadOnly),
				giu.Dummy(50, -1),
				giu.InputText(&ui.minY).Label("Search minY").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(func() {}),
				giu.InputText(&ui.minMove).Label("Search move").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(func() {}),
				giu.RadioButton("Auto Search", ui.auto).
					OnChange(func() { ui.auto = !ui.auto }),
			)),
		giu.Separator(),
		giu.InputText(&ui.lastResponse).Size(giu.Auto).Flags(giu.InputTextFlagsReadOnly).Hint("Последний ответ прибора..."), //Response for CMD
		giu.Separator(),
		giu.Style().SetFontSize(14).To(
			giu.Plot("Осцилограмма").Size(-3, -35-int(14+(ui.FPy*2)+2)*1).AxisLimits(0, 1550, -150, 150, giu.ConditionAlways).Plots(
				giu.Line("", UtoF(ui.linedata)),
				plots[1], plots[2], plots[3],
			)),
		giu.Separator(),
		giu.SliderInt(&ui.cursorPos[ui.cursorMode], 0, 1520).Size(-1),
		giu.Separator(),
		giu.Child().Size(-3, (14+(ui.FPy*2)+2)*1).Border(false).Layout(
			giu.Row(
				giu.RadioButton("Start", ui.cursorMode == CursorStart).
					OnChange(func() { ui.cursorMode = CursorStart }),

				giu.RadioButton("Reper", ui.cursorMode == CursorReper).
					OnChange(func() { ui.cursorMode = CursorReper }),

				giu.RadioButton("Front", ui.cursorMode == CursorFront).
					OnChange(func() { ui.cursorMode = CursorFront }),
			),
		),
	}
}
