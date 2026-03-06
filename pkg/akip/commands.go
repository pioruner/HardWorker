package akip

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"strconv"
	"time"

	"github.com/pioruner/HardWorker.git/pkg/app"
	"github.com/pioruner/HardWorker.git/pkg/logger"
)

func (ui *AkipUI) connectionLoop() {
	ui.Load()
	defer app.Wg.Done()
	defer logger.Infof("Module Akip stopped: %s", ui.id)
	defer ui.Save()
	retry := time.Second
	for {
		select {
		case <-app.Ctx.Done():
			logger.Infof("Akip connection loop cancelled: %s", ui.id)
			return
		default:
			logger.Infof("Akip connecting to %s (%s)", ui.adr, ui.id)
			conn, err := net.DialTimeout("tcp", ui.adr, time.Second)
			if err != nil {
				ui.lastResponse = err.Error()
				logger.Warnf("Akip connection failed (%s): %v", ui.adr, err)
				time.Sleep(retry)
				continue
			}
			ui.conn = conn

			ui.lastResponse = "Connected"
			ui.connected = true
			logger.Infof("Akip connected: %s", ui.adr)
			err = ui.sessionLoop()
			if err != nil {
				logger.Warnf("Akip session closed with error (%s): %v", ui.adr, err)
			} else {
				logger.Infof("Akip session closed: %s", ui.adr)
			}
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
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	ui.cmdCh <- SCPICommand{Cmd: ":SDSLSCPI#"}
	ui.SetTime()
	ui.SetOffset()

	for {
		select {
		case <-app.Ctx.Done():
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
	_ = ui.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 1000))
	if _, err := ui.conn.Write([]byte("STARTBIN")); err != nil {
		ui.lastResponse = "Ошибка Write: " + err.Error()
		logger.Errorf("Akip write STARTBIN failed (%s): %v", ui.adr, err)
		return err
	}
	header := make([]byte, 12)
	_ = ui.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 1000))
	if _, err := io.ReadFull(ui.conn, header); err != nil {
		ui.lastResponse = "Ошибка Read Header: " + err.Error()
		ui.setUpdate()
		logger.Errorf("Akip read header failed (%s): %v", ui.adr, err)
		return err
	}
	size := binary.LittleEndian.Uint16(header[0:2])
	if size == 0 || size > 65535 {
		logger.Errorf("Akip invalid payload size (%s): %d", ui.adr, size)
		return fmt.Errorf("invalid payload size: %d", size)
	}
	payload := make([]byte, size)
	if _, err := io.ReadFull(ui.conn, payload); err != nil {
		ui.lastResponse = "Ошибка Read Payload: " + err.Error()
		ui.setUpdate()
		logger.Errorf("Akip read payload failed (%s): %v", ui.adr, err)
		return err
	}
	packet := make([]byte, 0, 12+len(payload))
	packet = append(packet, header...)
	packet = append(packet, payload...)
	var dt, hoffs float64
	var ndata int
	var err error
	ui.linedata, dt, hoffs, ndata, err = ui.binUnpuck(packet, true)
	if err != nil {
		logger.Errorf("Akip binary unpack failed (%s): %v", ui.adr, err)
		return err
	}
	ui.xsize = ndata - 1
	//dt = dt * 15.2 / float64(ndata) * 1000000  // для пересчета dt полученного от осцилографа
	ui.Atime = fmt.Sprintf("%.1f", dt)
	ui.Aoffset = fmt.Sprintf("%.1f", hoffs)
	dt = TimeScale[ui.timeB] * 15.2 / float64(ndata) * 1000000
	//hoffs, _ = strconv.ParseFloat(ui.Hoffset, 64)  // для использования смещения с UI, если от осцилографа не работает
	ui.plotData = UtoF(ui.linedata)
	ui.Y = ui.plotData
	ui.X = make([]float64, ndata)
	for i := -ndata / 2; i < ndata/2; i++ {
		ui.X[i+ndata/2] = (float64(i) * dt) + hoffs
	}
	ui.xhoffs = hoffs
	ui.xdt = dt
	ui.connected = true
	inx, offs_new, find := ui.findPeak()
	if find {
		ui.cursorPos[2] = float32(ui.X[inx])
		if offs_new != 0 {
			ui.Hoffset = fmt.Sprintf("%.0f", hoffs-offs_new-baseOffest[ui.timeB])
			ui.SetOffset()
		}
	}
	ui.Calc()
	ui.setUpdate()
	return nil
}

func (ui *AkipUI) Calc() {
	c0 := ui.cursorPos[0] / 2
	c1 := ui.cursorPos[1] / 2
	c2 := ui.cursorPos[2] / 2
	if c1-c0 == 0 {
		return
	}
	time := c2 - c0
	repSize, _ := strconv.ParseFloat(ui.reper, 64)
	rep := repSize / (float64(c1-c0) / 1000000)
	speed := rep / 100
	sq, _ := strconv.ParseFloat(ui.square, 64)
	volume := float64(time) * rep * sq / 1000000
	ui.vtime = fmt.Sprintf("%.2f", time)
	ui.vspeed = fmt.Sprintf("%.2f", speed)
	ui.volume = fmt.Sprintf("%.2f", volume)
}

func (ui *AkipUI) findPeak() (int, float64, bool) {
	if ui.auto {
		minx, maxx := -1, -1
		for i, x := range ui.X {
			if x > float64(ui.cursorPos[2])-0.5 && x < float64(ui.cursorPos[2])+0.5 {
				if minx < 0 {
					minx = i
				}
				maxx = i
			}
		}
		if minx < 0 || maxx < 0 {
			return 0, 0, false
		}
		maxy, maxy_index := -1.0, 1
		for i := minx; i < maxx; i++ {
			if ui.Y[i] > maxy {
				maxy = ui.Y[i]
				maxy_index = i
			}
		}
		offset := ui.X[len(ui.X)/2] - ui.X[maxy_index]
		minR, _ := strconv.ParseFloat(ui.minY, 64)
		minMove, _ := strconv.ParseFloat(ui.minMove, 64)
		reply := false
		if maxy > minR {
			reply = true
		}
		if math.Abs(offset) < minMove {
			offset = 0
		}
		return maxy_index, offset, reply

	}
	return 0, 0, false
}

func (ui *AkipUI) SendCMD(cmd string) error {
	_ = ui.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 1000))
	if _, err := ui.conn.Write([]byte(cmd)); err != nil { //+ "\r\n" - работает и без этого
		logger.Errorf("Akip command send failed (%s, cmd=%q): %v", ui.adr, cmd, err)
		return err
	}
	ui.lastResponse = string(cmd)
	ui.setUpdate()
	time.Sleep(time.Millisecond * 200)
	return nil
}

func (ui *AkipUI) binUnpuck(buf []byte, ch1 bool) ([]int8, float64, float64, int, error) {
	size := binary.LittleEndian.Uint16(buf[0:2])
	dataBuf := buf[12:size]
	ch1_index := bytes.Index(dataBuf, []byte{0x43, 0x48, 0x31})
	ch2_index := bytes.Index(dataBuf, []byte{0x43, 0x48, 0x32})
	if ch1 {
		return ui.chanelUnpuck(dataBuf, ch1_index)
	} else {
		return ui.chanelUnpuck(dataBuf, ch2_index)
	}

}

func (ui *AkipUI) chanelUnpuck(buf []byte, index int) ([]int8, float64, float64, int, error) {
	if index == -1 {
		return nil, 0, 0, 0, fmt.Errorf("No Index")
	}
	nData := int(binary.LittleEndian.Uint16(buf[index+15 : index+17]))
	hMove := binary.LittleEndian.Uint16(buf[index+31 : index+33])
	//dt := TimeScale[buf[index+27]-7]  // опасно - не все режимы описаны и может вылезать за пределы!!!
	hoffs := float64(math.Float32frombits(binary.LittleEndian.Uint32(buf[index-12 : index+4-12])))
	shift := index + 59
	var wave = make([]int8, nData)
	for i := 0; i < nData; i++ {
		if shift+2 > len(buf) {
			return nil, 0, 0, 0, fmt.Errorf("Big buf shift")
		}
		wave[i] = int8(binary.LittleEndian.Uint16(buf[shift:shift+2]) + hMove)
		shift += 2
	}
	return wave, 0, hoffs, nData, nil
}

func UtoF(data []int8) []float64 {
	result := make([]float64, len(data))
	for i, v := range data {
		result[i] = float64(v)
	}
	return result
}
