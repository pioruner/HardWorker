package loger

func (ui *LogUI) Build() {
	for _, w := range ui.UI() {
		w.Build()
	}
}

func (ui *LogUI) Run() {
	return
}

func (ui *LogUI) Name() string {
	return "Системный ЛОГ"
}
