package akip

import "github.com/AllenDang/giu"

type AkipW struct {
	adr          string
	id           string
	commandInput string // Текущая команда для ввода
	lastResponse string // Последний ответ прибора
	linedata     []int8
	plotData     []float64
}

const (
	testAk = false
)

func New(adrPort string) *AkipW {
	return &AkipW{
		adr: adrPort,
	}
}

func (ak *AkipW) Build() {
	for _, w := range ak.UI() {
		w.Build()
	}
}

func (ak *AkipW) UI() giu.Layout {
	_, h := giu.GetFramePadding()
	return giu.Layout{
		giu.Align(giu.AlignCenter).To(
			giu.Style().SetFontSize(16).To(giu.Label("АКИП")), //Main Lable
		),
		giu.Separator(),
		giu.Child().Size(giu.Auto, (14+(h*2)+2)*1).Border(false).Layout(
			giu.Row(
				//giu.Style().SetFontSize(20).To(giu.InputText(&commandInput).Size(-200).Hint("Введите SCPI команду...")), //CMD for send
				giu.InputText(&ak.adr).Size(-200).Hint("Введите IP:Port..."),
				giu.Button("Проверить").Size(190, giu.Auto).OnClick(ak.test), //Send CMD
			)),
		giu.Separator(),

		giu.Child().Size(giu.Auto, (14+(h*2)+2)*1).Border(false).Layout(
			giu.Row(
				//giu.Style().SetFontSize(20).To(giu.InputText(&commandInput).Size(-200).Hint("Введите SCPI команду...")), //CMD for send
				giu.InputText(&ak.commandInput).Size(-200).Hint("Введите SCPI команду..."),
				giu.Button("Отправить").Size(190, giu.Auto).OnClick(ak.send), //Send CMD
			)),

		giu.Dummy(0, 5),

		//giu.Label("Последний ответ прибора:"),
		giu.InputText(&ak.lastResponse).Size(giu.Auto).Flags(giu.InputTextFlagsReadOnly).Hint("Последний ответ прибора..."), //Response for CMD

		giu.Dummy(0, 5),
		giu.Style().SetFontSize(10).To(
			giu.Plot("Осцилограмма").Size(-1, 300).AxisLimits(0, 1550, -150, 150, giu.ConditionAlways).Plots(
				giu.Line("", UtoF(ak.linedata)), //ak.plotData),
			)),
	}
}
