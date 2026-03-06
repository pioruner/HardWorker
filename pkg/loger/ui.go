package loger

import "github.com/AllenDang/giu"

func (ui *LogUI) UI() giu.Layout {
	ui.mu.Lock()
	lines := make([]logLine, len(ui.lines))
	copy(lines, ui.lines)
	ui.mu.Unlock()

	var widgets []giu.Widget

	// Итерируемся в обратном порядке: новые сообщения сверху
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		widgets = append(widgets, giu.Label(line.text))

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
