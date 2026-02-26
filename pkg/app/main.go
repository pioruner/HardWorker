package app

import (
	"context"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
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

const (
	tcpHost = "127.0.0.1:54424"
)

func CheckInstatse() bool {
	conn, err := net.DialTimeout("tcp", tcpHost, time.Millisecond*100)
	if err == nil {
		conn.Close()
		return true
	}
	go startTCPListener()
	return false
}

func startTCPListener() {
	l, err := net.Listen("tcp", tcpHost)
	if err != nil {
		log.Println("Cannot start TCP listener:", err)
		return
	}
	defer l.Close()

	for {
		_, err := l.Accept()
		if err != nil {
			continue
		}
		if State.Gui {
			Event <- EventToggleGUI
		}
	}
}
