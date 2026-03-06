package loger

import "github.com/AllenDang/giu"

func (ui *LogUI) UI() giu.Layout {

	var widgets []giu.Widget

	for _, line := range ui.lines {

		widgets = append(widgets,
			giu.Style().
				To(giu.Label(line.text)),
		)
	}

	child := giu.Child().
		Size(-1, -1).
		Layout(widgets...)

	return giu.Layout{
		child,
	}
}
