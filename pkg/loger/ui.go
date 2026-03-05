package loger

import "github.com/AllenDang/giu"

type LogUI struct {
	adr string
}

func (ui *LogUI) UI() giu.Layout {
	return giu.Layout{}
}
