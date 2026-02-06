package akip

import (
	"fmt"
	"strconv"

	"github.com/AllenDang/giu"
)

func (ui *AkipUI) Build() {
	for _, w := range ui.UI() {
		w.Build()
	}
}

func drawCursor(name string, x float64, yMin, yMax float64) giu.PlotWidget {
	return giu.LineXY(
		name,
		[]float64{x, x},
		[]float64{yMin, yMax},
	)
}

func (ui *AkipUI) SetTime() {
	if ui.timeB > 2 {
		ui.auto = false
	}
	ui.cmdCh <- SCPICommand{Cmd: fmt.Sprintf(":TIMebase:SCALe %s", TimeScaleS[ui.timeB])}
	ui.SetOffset()
}
func (ui *AkipUI) SetOffset() {
	hoff, err := strconv.ParseFloat(ui.Hoffset, 64)
	if err != nil {
		return
	}
	value := (hoff + baseOffest[ui.timeB]) / (TimeScale[ui.timeB] / 50.0)
	ui.cmdCh <- SCPICommand{Cmd: fmt.Sprintf(":TIMebase:HOFFset %d", int(value*1e-6))}
}

func (ui *AkipUI) UI() giu.Layout {
	if ui.update {
		giu.Update()
		ui.update = false
	}
	ui.FPx, ui.FPy = giu.GetFramePadding()
	ui.MacMult = 1

	plots := []giu.PlotWidget{
		giu.LineXY("УЗ Волна", ui.X, ui.Y),
	}
	names := []string{"Начало", "Репер", "Граница"}
	for i := 0; i < 3; i++ {
		x := float64(ui.cursorPos[i])
		plots = append(plots, drawCursor(names[i], x, -150, 150))
	}

	return giu.Layout{
		giu.Child().Size(-3, (14+(ui.FPy*2)+2)*ui.MacMult).Border(false).Layout(
			giu.Row(
				giu.Style().SetDisabled(!(ui.connected)).To(
					giu.Label("Развертка"),
					giu.Combo("мкс", TimeScaleS[ui.timeB], TimeScaleS, &ui.timeB).Size(100).OnChange(ui.SetTime),
					giu.Dummy(25, -1),
					giu.Label("Смещение"),
					giu.InputText(&ui.Hoffset).Label("мкс").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(ui.SetOffset),
					giu.Dummy(25, -1),
					giu.Label("Репер"),
					giu.InputText(&ui.reper).Label("dL см").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(func() {}),
					giu.Dummy(25, -1),
					giu.Label("Площадь трубки"),
					giu.InputText(&ui.square).Label("см^2").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(func() {}),
					giu.Align(giu.AlignRight).To(giu.RadioButton("Связь", ui.connected)),
				),
			)),
		giu.Style().SetDisabled(!(ui.connected)).To(
			giu.Child().Size(-3, (14+(ui.FPy*2)+2)*ui.MacMult).Border(false).Layout(
				giu.Row(

					//giu.Dummy(50, -1),
					giu.Label("Скорость волны"),
					giu.InputText(&ui.vspeed).Label("м/с").Size(75).Flags(giu.InputTextFlagsReadOnly),
					giu.Dummy(25, -1),
					giu.Label("Время волны"),
					giu.InputText(&ui.vtime).Label("мкс").Size(75).Flags(giu.InputTextFlagsReadOnly),
					giu.Dummy(25, -1),
					giu.Label("Объём фазы"),
					giu.InputText(&ui.volume).Label("см^3").Size(75).Flags(giu.InputTextFlagsReadOnly),
				)),
			giu.Separator(),
			giu.InputText(&ui.lastResponse).Size(giu.Auto).Flags(giu.InputTextFlagsReadOnly).Hint("Последний ответ прибора..."), //Response for CMD
			giu.Separator(),
			giu.Style().SetFontSize(14).To(
				giu.Plot("Осцилограмма").Size(-3, -35-int(14+(ui.FPy*2)+2)*1).AxisLimits(ui.X[0], ui.X[ui.xsize], -150, 150, giu.ConditionAlways).Plots(
					plots...,
				)),
			giu.Separator(),
			giu.SliderFloat(&ui.cursorPos[ui.cursorMode], float32(ui.X[0]), float32(ui.X[ui.xsize])).Size(-1),
			giu.Separator(),
			giu.Row(
				giu.Label("Курсоры: "),
				giu.RadioButton("Старт", ui.cursorMode == CursorStart).
					OnChange(func() { ui.cursorMode = CursorStart }),

				giu.RadioButton("Репер", ui.cursorMode == CursorReper).
					OnChange(func() { ui.cursorMode = CursorReper }),

				giu.RadioButton("Граница", ui.cursorMode == CursorFront).
					OnChange(func() { ui.cursorMode = CursorFront }),
				giu.Dummy(25, -1),
				/*
					giu.Label("Развертка"),
					giu.InputText(&ui.Atime).Label("мкс").Size(50).Flags(giu.InputTextFlagsReadOnly),
					giu.Dummy(25, -1),
					giu.Label("Смещение"),
					giu.InputText(&ui.Aoffset).Label("мкс").Size(50).Flags(giu.InputTextFlagsReadOnly),
					giu.Dummy(25, -1),
				*/
				giu.Label("Мин. Y"),
				giu.InputText(&ui.minY).Label("ед.").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(func() {}),
				giu.Dummy(25, -1),
				giu.Label("Мин. смещение"),
				giu.InputText(&ui.minMove).Label("мкс").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(func() {}),
				giu.Dummy(25, -1),
				giu.Style().SetDisabled(ui.timeB > 2).To(
					giu.Label("Автопоиск"),
					giu.RadioButton("", ui.auto).
						OnChange(func() { ui.auto = !ui.auto })),
			),
		),
	}
}
