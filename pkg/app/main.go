package app

import (
	"context"
	"runtime"
	"sync"
)

type Modules interface {
	Run()
}

func Run(mod ...Modules) {
	for _, i := range mod {
		i.Run()
	}
}

// Events
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

// OS
var (
	MacOS         bool
	MacMultiperUI float64
	AppName       string = string("АКИП УЗ")
)

func init() {
	MacOS = runtime.GOOS == "darwin" //Check OS
	MacMultiperUI = 1.6
	//MacOS = true
}

// State

var Ctx, Cancel = context.WithCancel(context.Background())
var Wg sync.WaitGroup

type AppState struct {
	Gui bool
}

var State *AppState

func init() {
	State = &AppState{
		Gui: false,
	}
}
