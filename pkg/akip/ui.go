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
	ui.cmdCh <- SCPICommand{Cmd: fmt.Sprintf(":TIMebase:SCALe %s", TimeScaleS[ui.timeB])}
}
func (ui *AkipUI) SetOffset() {
	hoff, err := strconv.ParseFloat(ui.Hoffset, 64)
	if err != nil {
		return
	}
	value := (hoff) / (TimeScale[ui.timeB] / 50.0)
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
		giu.LineXY("Wave", ui.X, ui.Y),
	}
	names := []string{"Start", "Reper", "Front"}
	for i := 0; i < 3; i++ {
		x := float64(ui.cursorPos[i])
		plots = append(plots, drawCursor(names[i], x, -150, 150))
	}
	return giu.Layout{
		giu.Align(giu.AlignCenter).To(
			giu.Style().SetFontSize(16).To(giu.Label("АКИП")), //Main Lable
		),
		giu.Separator(),
		giu.Child().Size(-3, (14+(ui.FPy*2)+2)*ui.MacMult).Border(false).Layout(
			giu.Row(
				giu.RadioButton("Connection", ui.connected),
				giu.Dummy(25, -1),
				giu.Style().SetDisabled(!(ui.connected)).To(
					giu.Combo("TimeBase", TimeScaleS[ui.timeB], TimeScaleS, &ui.timeB).Size(100).OnChange(ui.SetTime),
					giu.Dummy(25, -1),
					giu.InputText(&ui.Hoffset).Label("H Offset").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(ui.SetOffset),
				),
				giu.Dummy(10, -1),
			)),
		giu.Style().SetDisabled(!(ui.connected)).To(
			giu.Child().Size(-3, (14+(ui.FPy*2)+2)*ui.MacMult).Border(false).Layout(
				giu.Row(
					giu.InputText(&ui.reper).Label("dL Reper").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(func() {}),
					giu.Dummy(25, -1),
					giu.InputText(&ui.square).Label("S Square").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(func() {}),
					giu.Dummy(50, -1),
					giu.InputText(&ui.vspeed).Label("Speed").Size(75).Flags(giu.InputTextFlagsReadOnly),
					giu.Dummy(25, -1),
					giu.InputText(&ui.vtime).Label("Time").Size(75).Flags(giu.InputTextFlagsReadOnly),
					giu.Dummy(25, -1),
					giu.InputText(&ui.volume).Label("Volume").Size(75).Flags(giu.InputTextFlagsReadOnly),
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
				giu.RadioButton("Start", ui.cursorMode == CursorStart).
					OnChange(func() { ui.cursorMode = CursorStart }),

				giu.RadioButton("Reper", ui.cursorMode == CursorReper).
					OnChange(func() { ui.cursorMode = CursorReper }),

				giu.RadioButton("Front", ui.cursorMode == CursorFront).
					OnChange(func() { ui.cursorMode = CursorFront }),
				giu.Dummy(10, -1),
				giu.InputText(&ui.Atime).Label("TimeBase").Size(50).Flags(giu.InputTextFlagsReadOnly),
				giu.InputText(&ui.Aoffset).Label("HOffset").Size(50).Flags(giu.InputTextFlagsReadOnly),
				giu.Dummy(10, -1),
				giu.InputText(&ui.minY).Label("Search minY").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(func() {}),
				giu.InputText(&ui.minMove).Label("Search move").Size(50).Hint("").Flags(giu.InputTextFlagsCharsDecimal).OnChange(func() {}),
				giu.RadioButton("Auto Search", ui.auto).
					OnChange(func() { ui.auto = !ui.auto }),
			),
		),
	}
}
