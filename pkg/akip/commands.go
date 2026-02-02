package akip

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"math"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/AllenDang/giu"
)

// Connection

func (ui *AkipUI) toggleConnection() {
	if ui.connected {
		ui.disconnect()
		return
	}
	ui.connect()
}

func (ui *AkipUI) connect() {
	conn, err := net.DialTimeout("tcp", ui.adr, time.Second)
	if err != nil {
		ui.lastResponse = err.Error()
		return
	}
	//_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
	_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 200))
	ui.conn = conn
	ui.connected = true
	ui.lastResponse = "Connected"
	log.Printf("Connected")
	ui.cmdCh = make(chan SCPICommand, 8)
	ui.ctx, ui.wcancel = context.WithCancel(context.Background())

	go ui.worker()
}

func (ui *AkipUI) disconnect() {
	ui.connected = false
	if ui.conn != nil {
		ui.conn.Close()
		ui.conn = nil
	}
	if ui.wcancel != nil {
		ui.wcancel() // 🔥 корректно останавливает worker
	}

	ui.lastResponse = "Disconnected"
}

func (ui *AkipUI) worker() {
	ticker := time.NewTicker(500 * time.Millisecond) // 10 Hz
	defer ticker.Stop()
	ui.cmdCh <- SCPICommand{Cmd: ":SDSLSCPI#"}
	ui.SetTime()
	ui.SetOffset()
	for ui.connected {
		select {
		case <-ui.ctx.Done():
			log.Printf("Worker stopped")
			return
		case cmd := <-ui.cmdCh:
			log.Printf("Send CMD %s", cmd.Cmd)
			ui.SendCMD(cmd.Cmd)

		case <-ticker.C:
			log.Printf("ReadWave")
			ui.ReadWave()
		}
	}
}

func (ui *AkipUI) ReadWave() {
	if _, err := ui.conn.Write([]byte("STARTBIN")); err != nil {
		ui.lastResponse = "Ошибка Write: " + err.Error()
		return
	}
	header := make([]byte, 12)
	if _, err := io.ReadFull(ui.conn, header); err != nil {
		ui.lastResponse = "Ошибка Read Header: " + err.Error()
		giu.Update()
		return
	}
	size := binary.LittleEndian.Uint16(header[0:2])
	payload := make([]byte, size)
	if _, err := io.ReadFull(ui.conn, payload); err != nil {
		ui.lastResponse = "Ошибка Read Payload: " + err.Error()
		giu.Update()
		return
	}
	packet := make([]byte, 0, 12+len(payload))
	packet = append(packet, header...)
	packet = append(packet, payload...)
	var dt, hoffs float64
	ui.linedata, dt, hoffs = ui.binUnpuck(packet, true)
	dt = TimeScale[ui.timeB] * 15.2 / 3040 * 1000000
	hoffs, _ = strconv.ParseFloat(ui.Hoffset, 64)
	ui.plotData = UtoF(ui.linedata)
	ui.Y = ui.plotData
	ui.X = make([]float64, len(ui.plotData))
	for i := 0; i < len(ui.plotData); i++ {
		ui.X[i] = (float64(i) * dt) + hoffs
	}
	ui.xhoffs = hoffs
	ui.xdt = dt
	ui.xsize = len(ui.plotData) - 1
	//log.Printf("X0 %f, XL %f, dT %f, HOffst %f, lenData %d, lastX %f", ui.X[0], ui.X[1519], dt, hoffs, len(ui.plotData)-1, ui.X[ui.xsize])
	giu.Update()
}
func (ui *AkipUI) SendCMD(cmd string) {
	if _, err := ui.conn.Write([]byte(cmd)); err != nil { //+ "\r\n"
		return
	}
	ui.lastResponse = string(cmd)
	giu.Update()
}

func (ui *AkipUI) binUnpuck(buf []byte, ch1 bool) ([]int8, float64, float64) {
	size := binary.LittleEndian.Uint16(buf[0:2])
	log.Print("Размер осцилограммы: ")
	log.Println(size)
	dataBuf := buf[12:size]
	ch1_index := bytes.Index(dataBuf, []byte{0x43, 0x48, 0x31})
	log.Printf("CH1 Index: %d", ch1_index)
	ch2_index := bytes.Index(dataBuf, []byte{0x43, 0x48, 0x32})
	log.Printf("CH2 Index: %d", ch2_index)
	if ch1 {
		return ui.chanelUnpuck(dataBuf, ch1_index)
	} else {
		return ui.chanelUnpuck(dataBuf, ch2_index)
	}

}

func (ui *AkipUI) chanelUnpuck(buf []byte, index int) ([]int8, float64, float64) {
	nData := int(binary.LittleEndian.Uint16(buf[index+15 : index+17]))
	log.Printf("NDATA: %d", nData)
	hMove := binary.LittleEndian.Uint16(buf[index+31 : index+33])
	log.Printf("hMove: %d", hMove)
	dt := TimeScale[buf[index+27]-7]
	log.Printf("dT: %f", dt)
	hoffs := float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[index : index+3])))
	log.Printf("HOffest: %f", hoffs)
	shift := index + 59
	log.Printf("shift: %d", shift)
	var wave = make([]int8, nData)
	for i := 0; i < nData; i++ {
		wave[i] = int8(binary.LittleEndian.Uint16(buf[shift:shift+2]) + hMove)
		shift += 2
	}
	return wave[len(wave)/2:], dt, hoffs
}

func UtoF(data []int8) []float64 {
	result := make([]float64, len(data))
	for i, v := range data {
		result[i] = float64(v)
	}
	return result
}

// LOAD && SAVE
func AppConfigPath(name string) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "HardWorker")
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, name+".json"), nil
}

func LoadState(path string) (AkipState, error) {
	var state AkipState

	data, err := os.ReadFile(path)
	if err != nil {
		return state, err
	}
	err = json.Unmarshal(data, &state)
	return state, err
}

func SaveState(path string, state AkipState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (ui *AkipUI) ExportState() AkipState {
	return AkipState{
		Adr:        ui.adr,
		TimeB:      ui.timeB,
		Auto:       ui.auto,
		Hoffset:    ui.Hoffset,
		Reper:      ui.reper,
		Square:     ui.square,
		Vspeed:     ui.vspeed,
		Vtime:      ui.vtime,
		Volume:     ui.volume,
		MinY:       ui.minY,
		MinMove:    ui.minMove,
		CursorMode: ui.cursorMode,
		CursorPos:  ui.cursorPos,
	}
}

func (ui *AkipUI) ImportState(s AkipState) {
	ui.adr = s.Adr
	ui.timeB = s.TimeB
	ui.auto = s.Auto
	ui.Hoffset = s.Hoffset
	ui.reper = s.Reper
	ui.square = s.Square
	ui.vspeed = s.Vspeed
	ui.vtime = s.Vtime
	ui.volume = s.Volume
	ui.minY = s.MinY
	ui.minMove = s.MinMove
	ui.cursorMode = s.CursorMode
	ui.cursorPos = s.CursorPos
}
