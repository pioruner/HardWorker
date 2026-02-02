package akip

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func (ui *AkipUI) connectionLoop() {
	ui.Load()
	defer ui.wg.Done()
	defer log.Printf("Module Akip with name: %s --STOPED", ui.id)
	defer ui.Save()
	retry := time.Second
	for {
		select {
		case <-ui.ctx.Done():
			return
		default:
			conn, err := net.DialTimeout("tcp", ui.adr, time.Second)
			if err != nil {
				time.Sleep(retry)
				continue
			}
			ui.conn = conn

			_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))
			ui.lastResponse = "Connected"
			err = ui.sessionLoop()
			conn.Close()
			ui.conn = nil
			ui.connected = false
			time.Sleep(retry)
			ui.lastResponse = "Disconnected"
			ui.setUpdate()
		}

	}
}

func (ui *AkipUI) sessionLoop() error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	ui.cmdCh <- SCPICommand{Cmd: ":SDSLSCPI#"}
	ui.SetTime()
	ui.SetOffset()

	for {
		select {
		case <-ui.ctx.Done():
			return nil

		case <-ticker.C:
			if err := ui.ReadWave(); err != nil {
				return err // <-- триггер reconnect
			}

		case cmd := <-ui.cmdCh:
			if err := ui.SendCMD(cmd.Cmd); err != nil {
				return err
			}
		}
	}
}

func (ui *AkipUI) setUpdate() {
	ui.update = true
}

func (ui *AkipUI) ReadWave() error {
	if _, err := ui.conn.Write([]byte("STARTBIN")); err != nil {
		ui.lastResponse = "Ошибка Write: " + err.Error()
		return err
	}
	header := make([]byte, 12)
	if _, err := io.ReadFull(ui.conn, header); err != nil {
		ui.lastResponse = "Ошибка Read Header: " + err.Error()
		ui.setUpdate()
		return err
	}
	size := binary.LittleEndian.Uint16(header[0:2])
	payload := make([]byte, size)
	if _, err := io.ReadFull(ui.conn, payload); err != nil {
		ui.lastResponse = "Ошибка Read Payload: " + err.Error()
		ui.setUpdate()
		return err
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
	ui.connected = true
	ui.Calc()
	ui.setUpdate()
	return nil
}

func (ui *AkipUI) Calc() {
	time := ui.cursorPos[2] - ui.cursorPos[0]
	repSize, _ := strconv.ParseFloat(ui.reper, 64)
	rep := repSize / (float64(ui.cursorPos[1]-ui.cursorPos[0]) / 1000000)
	speed := rep / 100
	sq, _ := strconv.ParseFloat(ui.square, 64)
	volume := float64(time) * rep * sq
	ui.vtime = fmt.Sprintf("%.2f", time)
	ui.vspeed = fmt.Sprintf("%.2f", speed)
	ui.volume = fmt.Sprintf("%.2f", volume)
}

func (ui *AkipUI) SendCMD(cmd string) error {
	if _, err := ui.conn.Write([]byte(cmd)); err != nil { //+ "\r\n"
		return err
	}
	ui.lastResponse = string(cmd)
	ui.setUpdate()
	return nil
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
