package visko

import (
	"fmt"
	"log"
	"time"

	"github.com/pioruner/HardWorker.git/pkg/app"
	"github.com/simonvetter/modbus"
)

func (ui *ViskoUI) connectionLoop() {
	defer app.Wg.Done()
	defer log.Printf("Module Visko with name: %s --STOPED", ui.id)
	retry := time.Second
	conn, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     "tcp://" + ui.adr,
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return
	}
	for {
		select {
		case <-app.Ctx.Done():
			return
		default:
			err = conn.Open()
			if err != nil {
				time.Sleep(retry)
				continue
			}
			ui.conn = conn
			err = ui.sessionLoop()
			err = conn.Close()
			ui.conn = nil
			ui.connected = false
			time.Sleep(retry)
			ui.setUpdate()
		}

	}
}

func (ui *ViskoUI) sessionLoop() error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-app.Ctx.Done():
			return nil

		case <-ticker.C:
			if err := ui.Read(); err != nil {
				return err // <-- триггер reconnect
			}
		}
	}
}

func (ui *ViskoUI) Read() error {
	ui.conn.SetEncoding(modbus.BIG_ENDIAN, modbus.LOW_WORD_FIRST)

	t1, err := ui.conn.ReadRegister(16384, modbus.HOLDING_REGISTER)
	if err != nil {
		return err
	}
	t2, err := ui.conn.ReadRegister(16385, modbus.HOLDING_REGISTER)
	if err != nil {
		return err
	}
	temp, err := ui.conn.ReadFloat32(16386, modbus.HOLDING_REGISTER)
	if err != nil {
		return err
	}
	u1, err := ui.conn.ReadFloat32(16388, modbus.HOLDING_REGISTER)
	if err != nil {
		return err
	}
	u2, err := ui.conn.ReadFloat32(16390, modbus.HOLDING_REGISTER)
	if err != nil {
		return err
	}
	cmd, err := ui.conn.ReadRegister(16392, modbus.HOLDING_REGISTER)
	if err != nil {
		return err
	}

	ui.curT1 = fmt.Sprintf("%d", t1)
	ui.curT2 = fmt.Sprintf("%d", t2)
	ui.curU1 = fmt.Sprintf("%.2f", u1)
	ui.curU2 = fmt.Sprintf("%.2f", u2)
	ui.curTemp = fmt.Sprintf("%.1f", temp)

	if ui.cmd == 0 && cmd > 0 {

		ui.AddRow(float64(t1), float64(t2), float64(u1), float64(u2), float64(temp))
	}
	ui.cmd = cmd
	ui.connected = true
	ui.setUpdate()
	//reg16s, err := ui.conn.ReadRegisters(100, 4, modbus.INPUT_REGISTER)

	//client.SetEncoding(modbus.LITTLE_ENDIAN, modbus.LOW_WORD_FIRST)

	//fl32s, err  := client.ReadFloat32s(100, 2, modbus.INPUT_REGISTER)

	return nil
}
