package loger

import (
	"image/color"

	"github.com/AllenDang/giu"
)

var (
	infoColor  = color.RGBA{R: 90, G: 200, B: 255, A: 255}
	warnColor  = color.RGBA{R: 240, G: 190, B: 70, A: 255}
	errorColor = color.RGBA{R: 255, G: 100, B: 100, A: 255}
	debugColor = color.RGBA{R: 170, G: 170, B: 170, A: 255}
)

func colorByLevel(level logLevel) color.Color {
	switch level {
	case levelError:
		return errorColor
	case levelWarn:
		return warnColor
	case levelDebug:
		return debugColor
	default:
		return infoColor
	}
}

func (ui *LogUI) UI() giu.Layout {
	ui.mu.Lock()
	lines := make([]logLine, len(ui.lines))
	copy(lines, ui.lines)
	ui.mu.Unlock()

	var widgets []giu.Widget

	// Итерируемся в обратном порядке: новые сообщения сверху
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		widgets = append(widgets,
			giu.Style().
				SetColor(giu.StyleColorText, colorByLevel(line.level)).
				To(giu.Label(line.text)),
		)

		// Добавляем разделитель между сообщениями
		if i > 0 {
			widgets = append(widgets, giu.Separator())
		}
	}

	childWidget := giu.Child().
		Size(-1, -1).
		Layout(widgets...)

	return giu.Layout{
		childWidget,
	}
}
