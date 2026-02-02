package akip

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/AllenDang/giu"
)

//:SDSLSCPI#
//:TIMebase:HOFFset %d
//:TIMebase:SCALe %s
//:CHANnel%2$d:SCALe %1$s
//STARTBIN
/*
func (ak *AkipW) sendSCPI(addr, cmd string, timeout time.Duration) ([]byte, error) {
	return exData, nil
	dialer := net.Dialer{
		Timeout: timeout,
	}

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// write
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))
	if _, err := conn.Write([]byte(cmd + "\r\n")); err != nil {
		return nil, err
	}

	// read header (12 bytes)
	_ = conn.SetReadDeadline(time.Now().Add(timeout))

	header := make([]byte, 12)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}

	// payload size is in first 2 bytes
	size := binary.LittleEndian.Uint16(header[0:2])

	// read payload
	payload := make([]byte, size)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, err
	}

	// если нужен весь пакет целиком
	packet := make([]byte, 0, 12+len(payload))
	packet = append(packet, header...)
	packet = append(packet, payload...)

	return packet, nil
}

func (ak *AkipW) sendCMD() {
	resp, err := ak.sendSCPI(
		ak.adr,
		ak.commandInput,
		1*time.Second,
	)

	if err != nil {
		ak.lastResponse = "Ошибка: " + err.Error()
		giu.Update()
		return
	}

	if ak.commandInput == "STARTBIN" {
		if len(resp) > 0 {
			ak.linedata, _ = ak.binUnpuck(resp, true)
			ak.plotData = UtoF(ak.linedata)
		} else {
			clean := bytes.TrimRight(resp, "\x00\r\n")
			ak.lastResponse = string(clean)
		}

		giu.Update()
	}
}
*/

// Connection
type SCPICommand struct {
	Cmd string
}

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
	//_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 2000))
	ui.conn = conn
	ui.connected = true
	ui.lastResponse = "Connected"
	log.Printf("Connected")
	ui.cmdCh = make(chan SCPICommand, 8)

	go ui.worker()
}

func (ui *AkipUI) disconnect() {
	ui.connected = false
	if ui.conn != nil {
		ui.conn.Close()
		ui.conn = nil
	}
	close(ui.cmdCh)
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
	/*
		if _, err := ui.conn.Write([]byte("STARTBIN")); err != nil {
			ui.lastResponse = "Ошибка Write: " + err.Error()
			return
		}
	*/
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
	ui.plotData = UtoF(ui.linedata)
	ui.Y = ui.plotData
	ui.X = make([]float64, len(ui.plotData))
	hoffs, _ = strconv.ParseFloat(ui.Hoffset, 64)
	for i := 0; i < len(ui.plotData); i++ {
		ui.X[i] = (float64(i) * dt) + hoffs
		//log.Printf("X=%f", ui.X[i])
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
	/*
		reply := make([]byte, 50)
		if _, err := ui.conn.Read(reply); err != nil {
			ui.lastResponse = "Ошибка: " + err.Error()
			giu.Update()
			return
		}
		clean := bytes.TrimRight(reply, "\x00\r\n")
	*/
	ui.lastResponse = string(cmd)
	giu.Update()
}

func (ui *AkipUI) binUnpuck(buf []byte, ch1 bool) ([]int8, float64, float64) {
	size := binary.LittleEndian.Uint16(buf[0:2])
	log.Print("Размер осцилограммы: ")
	log.Println(size)

	dataBuf := buf[12:size]
	//log.Printf("Осцилограмма: % X", dataBuf)
	//hoffs := float64(math.Float32frombits(binary.BigEndian.Uint32(buf[0:4])))
	ch1_index := bytes.Index(dataBuf, []byte{0x43, 0x48, 0x31})
	log.Printf("CH1 Index: %d", ch1_index)
	ch2_index := bytes.Index(dataBuf, []byte{0x43, 0x48, 0x32})
	log.Printf("CH2 Index: %d", ch2_index)
	var data []int8
	var dtx float64
	if ch1 {
		data, dtx = ui.chanelUnpuck(dataBuf, ch1_index)
	} else {
		data, dtx = ui.chanelUnpuck(dataBuf, ch2_index)
	}
	return data, dtx, 0

}

func (ui *AkipUI) chanelUnpuck(buf []byte, index int) ([]int8, float64) {
	nData := int(binary.LittleEndian.Uint16(buf[index+15 : index+17]))
	log.Printf("NDATA: %d", nData)
	hMove := binary.LittleEndian.Uint16(buf[index+31 : index+33])
	log.Printf("hMove: %d", hMove)
	//dt := TimeScale[buf[index+27]-7]
	//log.Printf("dT: %f", dt)
	shift := index + 59
	log.Printf("shift: %d", shift)
	var wave = make([]int8, nData)
	for i := 0; i < nData; i++ {
		wave[i] = int8(binary.LittleEndian.Uint16(buf[shift:shift+2]) + hMove)
		shift += 2
	}
	return wave[len(wave)/2:], 0
}

// LOAD && SAVE

type AkipState struct {
	Adr   string
	TimeB int32
	Auto  bool

	Hoffset string
	Reper   string
	Square  string
	Vspeed  string
	Vtime   string
	Volume  string
	MinY    string
	MinMove string

	CursorMode CursorMode
	CursorPos  [3]int32
}

func AppConfigPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "HardWorker")
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "akip.json"), nil
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
