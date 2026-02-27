package visko

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/pioruner/HardWorker.git/pkg/app"
)

func (ui *AkipUI) connectionLoop() {
	ui.Load()
	defer app.Wg.Done()
	defer log.Printf("Module Visko with name: %s --STOPED", ui.id)
	defer ui.Save()
	retry := time.Second
	for {
		select {
		case <-app.Ctx.Done():
			return
		default:
			conn, err := net.DialTimeout("tcp", ui.adr, time.Second)
			if err != nil {
				ui.lastResponse = err.Error()
				time.Sleep(retry)
				continue
			}
			ui.conn = conn

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
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-app.Ctx.Done():
			return nil

		case <-ticker.C:
			if err := ui.ReadWave(); err != nil {
				return err // <-- триггер reconnect
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
	_ = ui.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 1000))
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
	var ndata int

	ui.xsize = ndata - 1
	//dt = dt * 15.2 / float64(ndata) * 1000000  // для пересчета dt полученного от осцилографа
	ui.Atime = fmt.Sprintf("%.1f", dt)
	ui.Aoffset = fmt.Sprintf("%.1f", hoffs)
	dt = TimeScale[ui.timeB] * 15.2 / float64(ndata) * 1000000
	//hoffs, _ = strconv.ParseFloat(ui.Hoffset, 64)  // для использования смещения с UI, если от осцилографа не работает

	ui.Y = ui.plotData
	ui.X = make([]float64, ndata)
	for i := -ndata / 2; i < ndata/2; i++ {
		ui.X[i+ndata/2] = (float64(i) * dt) + hoffs
	}
	ui.xhoffs = hoffs
	ui.xdt = dt
	ui.connected = true

	ui.setUpdate()
	return nil
}
