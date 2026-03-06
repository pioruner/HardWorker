package setts

func (ui *SettsUI) Build() {
	for _, w := range ui.UI() {
		w.Build()
	}
}

func (ui *SettsUI) Run() {
	return
}

func (ui *SettsUI) Name() string {
	return "Настройки"
}
