package app

import (
	"context"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/AllenDang/giu"
	"github.com/pioruner/HardWorker.git/pkg/logger"
)

type Modules interface {
	giu.Widget
	Run()
	Name() string
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
	AppName       string = string("Kortex Worker")
)

func init() {
	MacOS = runtime.GOOS == "darwin" //Check OS
	MacMultiperUI = 1.6
	MacOS = true
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
		logger.Warnf("Detected existing application instance on %s", tcpHost)
		return true
	}
	logger.Infof("No active instance found, starting local listener on %s", tcpHost)
	go startTCPListener()
	return false
}

func startTCPListener() {
	l, err := net.Listen("tcp", tcpHost)
	if err != nil {
		logger.Errorf("Cannot start TCP listener on %s: %v", tcpHost, err)
		return
	}
	logger.Infof("TCP listener started on %s", tcpHost)
	defer l.Close()

	for {
		_, err := l.Accept()
		if err != nil {
			logger.Warnf("TCP listener accept failed: %v", err)
			continue
		}
		logger.Infof("Received activation signal from another instance")
		if State.Gui {
			Event <- EventToggleGUI
		}
	}
}
