package app

type Events int

const (
	EventToggleGUI Events = iota
	EventCloseGUI
	EventQuit
)

var (
	Event = make(chan Events, 2)
)

var (
	CloseGUI = make(chan struct{}, 1)
	QuitGUI  = make(chan struct{}, 1) // Канал для сигнала выхода

)
