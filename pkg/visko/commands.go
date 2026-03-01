package visko

import (
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
	reg16, err := ui.conn.ReadRegister(100, modbus.HOLDING_REGISTER)
	if err != nil {
		return err
	} else {
		log.Printf("value: %v", reg16)        // as unsigned integer
		log.Printf("value: %v", int16(reg16)) // as signed integer
	}

	//reg16s, err := ui.conn.ReadRegisters(100, 4, modbus.INPUT_REGISTER)

	//client.SetEncoding(modbus.LITTLE_ENDIAN, modbus.LOW_WORD_FIRST)

	//fl32s, err  := client.ReadFloat32s(100, 2, modbus.INPUT_REGISTER)

	ui.setUpdate()
	return nil
}
