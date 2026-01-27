package app

type AppState struct {
	Gui bool
}

var State *AppState

func init() {
	State = &AppState{
		Gui: false,
	}
}
