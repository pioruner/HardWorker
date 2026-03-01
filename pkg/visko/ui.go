package visko

import (
	"github.com/AllenDang/giu"
)

func (ui *ViskoUI) Build() {
	for _, w := range ui.UI() {
		w.Build()
	}
}
func (ui *ViskoUI) UI() giu.Layout {

	timePlots, voltagePlots, tempPlots,
		xMin, xMax,
		timeYMin, timeYMax,
		voltYMin, voltYMax,
		tempYMin, tempYMax := ui.buildPlots()
	tableRows := ui.buildTable()

	hasData := len(ui.rows) > 0

	if hasData && ui.cursorIndex >= int32(len(ui.rows)) {
		ui.cursorIndex = int32(len(ui.rows))
	}

	if ui.update {
		giu.Update()
		ui.update = false
	}

	return giu.Layout{

		giu.Row(

			// ===================================================
			// ================= ЛЕВАЯ ПАНЕЛЬ ====================
			// ===================================================

			giu.Child().
				Size(440, -1).
				Border(true).
				Layout(

					giu.Row(
						giu.Label("T1"), giu.InputText(&ui.curT1).Flags(giu.InputTextFlagsReadOnly).Size(50),
						giu.Label("T2"), giu.InputText(&ui.curT2).Flags(giu.InputTextFlagsReadOnly).Size(50),
						giu.Label("U1"), giu.InputText(&ui.curU1).Flags(giu.InputTextFlagsReadOnly).Size(50),
						giu.Label("U2"), giu.InputText(&ui.curU2).Flags(giu.InputTextFlagsReadOnly).Size(50),
						giu.Label("Temp"), giu.InputText(&ui.curTemp).Flags(giu.InputTextFlagsReadOnly).Size(50),
					),

					giu.Separator(),
					giu.Row(
						giu.Button("Сохранить CSV").
							Size(200, 35).
							OnClick(ui.SaveCSVDialog),
						giu.Dummy(5, 35),
						giu.Button("Очистить").
							Size(200, 35).
							OnClick(func() {
								ui.rows = nil
								ui.cursorIndex = 0
							}),
					),
					giu.Separator(),

					giu.Table().
						Flags(giu.TableFlagsScrollY).
						Size(-1, -25).
						Columns(
							giu.TableColumn("T1"),
							giu.TableColumn("T2"),
							giu.TableColumn("U1"),
							giu.TableColumn("U2"),
							giu.TableColumn("Temp"),
						).
						Rows(tableRows...),
					giu.Align(giu.AlignRight).To(giu.RadioButton("Связь", ui.connected)),
				),

			// ===================================================
			// ================= ПРАВАЯ ПАНЕЛЬ ===================
			// ===================================================

			giu.Custom(func() {

				availX, availY := giu.GetAvailableRegion()

				// резервируем место под слайдер и индикаторы
				bottomBlock := float32(70)
				plotHeight := int((availY-bottomBlock)/3) - 1

				giu.Child().
					Size(availX, availY).
					Border(false).
					Layout(

						// ===== График времени =====
						giu.Plot("##timePlot").
							Size(-1, plotHeight).
							AxisLimits(xMin, xMax, timeYMin, timeYMax, giu.ConditionAlways).
							Plots(append(timePlots,
								drawCursorLine(int(ui.cursorIndex), -1000, 1000, "cursorT"),
							)...),

						// ===== График напряжения =====
						giu.Plot("##voltagePlot").
							Size(-1, plotHeight).
							AxisLimits(xMin, xMax, voltYMin, voltYMax, giu.ConditionAlways).
							Plots(append(voltagePlots,
								drawCursorLine(int(ui.cursorIndex), -1000, 1000, "cursorU"),
							)...),

						// ===== График температуры =====
						giu.Plot("##tempPlot").
							Size(-1, plotHeight).
							AxisLimits(xMin, xMax, tempYMin, tempYMax, giu.ConditionAlways).
							Plots(append(tempPlots,
								drawCursorLine(int(ui.cursorIndex), -1000, 1000, "cursorTemp"),
							)...),

						giu.Separator(),

						// ===== Общий слайдер =====
						giu.Style().SetDisabled(!hasData).To(
							giu.SliderInt(&ui.cursorIndex, 0, int32(len(ui.rows)-1)).
								Size(-1).
								OnChange(ui.updateCursorValues),
						),

						// ===== Значения выбранной точки =====
						giu.Row(
							giu.Label("T1"), giu.InputText(&ui.selT1).Flags(giu.InputTextFlagsReadOnly).Size(70),
							giu.Label("T2"), giu.InputText(&ui.selT2).Flags(giu.InputTextFlagsReadOnly).Size(70),
							giu.Label("U1"), giu.InputText(&ui.selU1).Flags(giu.InputTextFlagsReadOnly).Size(70),
							giu.Label("U2"), giu.InputText(&ui.selU2).Flags(giu.InputTextFlagsReadOnly).Size(70),
							giu.Label("Temp"), giu.InputText(&ui.selTemp).Flags(giu.InputTextFlagsReadOnly).Size(70),
						),
					).
					Build()
			}),
		),
	}
}
