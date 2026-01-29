package akip

import "github.com/AllenDang/giu"

type AkipUI struct {
	adr          string
	id           string
	commandInput string // Текущая команда для ввода
	lastResponse string // Последний ответ прибора
	linedata     []int8
	plotData     []float64

	FPx, FPy, MacMult float32
	timeB, slide      int32
	rep, start, front bool
}

func Init() *AkipUI {
	return &AkipUI{}
}

func (ui *AkipUI) Build() {
	for _, w := range ui.UI() {
		w.Build()
	}
}

func (ui *AkipUI) UI() giu.Layout {
	ui.FPx, ui.FPy = giu.GetFramePadding()
	ui.MacMult = 1
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
				giu.Combo("TimeBase", TimeScaleS[ui.timeB], TimeScaleS, &ui.timeB).Size(150),
			)),
		giu.Separator(),
		giu.InputText(&ui.lastResponse).Size(giu.Auto).Flags(giu.InputTextFlagsReadOnly).Hint("Последний ответ прибора..."), //Response for CMD
		giu.Separator(),
		giu.Style().SetFontSize(14).To(
			giu.Plot("Осцилограмма").Size(-3, -35-int(14+(ui.FPy*2)+2)*1).AxisLimits(0, 1550, -150, 150, giu.ConditionAlways).Plots(
				giu.Line("", UtoF(ui.linedata)), //ak.plotData),
			)),
		giu.Separator(),
		giu.SliderInt(&ui.slide, 0, 1520).Size(-3),
		giu.Separator(),
		giu.Child().Size(-3, (14+(ui.FPy*2)+2)*1).Border(false).Layout(
			giu.Row(
				giu.RadioButton("Start", ui.start).OnChange(ui.cursorS),
				giu.RadioButton("Reper", ui.rep).OnChange(ui.cursorR),
				giu.RadioButton("Front", ui.front).OnChange(ui.cursorF),
			),
		),
		/*
			giu.Child().Size(giu.Auto, (14+(ui.FPy*2)+2)*ui.MacMult).Border(false).Layout(
				giu.Row(
					//giu.Style().SetFontSize(20).To(giu.InputText(&commandInput).Size(-200).Hint("Введите SCPI команду...")), //CMD for send
					giu.InputText(&ui.commandInput).Size(-200).Hint("Введите SCPI команду..."),
					giu.Button("Отправить").Size(190, giu.Auto).OnClick(func() {}), //Send CMD
				)),

			giu.Dummy(0, 5),

			//giu.Label("Последний ответ прибора:"),
			giu.InputText(&ui.lastResponse).Size(giu.Auto).Flags(giu.InputTextFlagsReadOnly).Hint("Последний ответ прибора..."), //Response for CMD

			giu.Dummy(0, 5),
		*/
	}
}

func (ui *AkipUI) cursorS() {
	ui.start = true
	ui.rep = false
	ui.front = false
}
func (ui *AkipUI) cursorR() {
	ui.start = false
	ui.rep = true
	ui.front = false
}
func (ui *AkipUI) cursorF() {
	ui.start = false
	ui.rep = false
	ui.front = true
}

var TimeScaleS []string = []string{
	"1e-9 * 5",
	"1e-9 * 10",
	"1e-9 * 20",
	"1e-9 * 50",
	"1e-9 * 100",
	"1e-9 * 200",
	"1e-9 * 500",
	"1e-6 * 1",
	"1e-6 * 2",
	"1e-6 * 5",
	"1e-6 * 10",
	"1e-6 * 20",
	"1e-6 * 50",
	"1e-6 * 100",
	"1e-6 * 200",
	"1e-6 * 500",
	"1e-3 * 1",
	"1e-3 * 2",
	"1e-3 * 5",
	"1e-3 * 10",
	"1e-3 * 20",
	"1e-3 * 50",
	"1e-3 * 100",
	"1e-3 * 200",
	"1e-3 * 500",
}
