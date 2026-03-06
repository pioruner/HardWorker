package visko

import (
	"fmt"
	"time"

	"github.com/pioruner/HardWorker.git/pkg/app"
	"github.com/pioruner/HardWorker.git/pkg/logger"
	"github.com/simonvetter/modbus"
)

func (ui *ViskoUI) connectionLoop() {
	defer app.Wg.Done()
	defer logger.Infof("Module Visko stopped: %s", ui.id)
	retry := time.Second
	conn, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     "tcp://" + ui.adr,
		Timeout: 1 * time.Second,
	})
	if err != nil {
		logger.Errorf("Visko modbus client init failed (%s): %v", ui.adr, err)
		return
	}
	for {
		select {
		case <-app.Ctx.Done():
			logger.Infof("Visko connection loop cancelled: %s", ui.id)
			return
		default:
			logger.Infof("Visko connecting to %s (%s)", ui.adr, ui.id)
			err = conn.Open()
			if err != nil {
				logger.Warnf("Visko connection failed (%s): %v", ui.adr, err)
				time.Sleep(retry)
				continue
			}
			ui.conn = conn
			ui.connected = true
			logger.Infof("Visko connected: %s", ui.adr)
			err = ui.sessionLoop()
			if err != nil {
				logger.Warnf("Visko session closed with error (%s): %v", ui.adr, err)
			} else {
				logger.Infof("Visko session closed: %s", ui.adr)
			}
			if err = conn.Close(); err != nil {
				logger.Warnf("Visko connection close failed (%s): %v", ui.adr, err)
			}
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
		logger.Errorf("Visko read T1 failed (%s): %v", ui.adr, err)
		return err
	}
	t2, err := ui.conn.ReadRegister(16385, modbus.HOLDING_REGISTER)
	if err != nil {
		logger.Errorf("Visko read T2 failed (%s): %v", ui.adr, err)
		return err
	}
	temp, err := ui.conn.ReadFloat32(16386, modbus.HOLDING_REGISTER)
	if err != nil {
		logger.Errorf("Visko read Temp failed (%s): %v", ui.adr, err)
		return err
	}
	u1, err := ui.conn.ReadFloat32(16388, modbus.HOLDING_REGISTER)
	if err != nil {
		logger.Errorf("Visko read U1 failed (%s): %v", ui.adr, err)
		return err
	}
	u2, err := ui.conn.ReadFloat32(16390, modbus.HOLDING_REGISTER)
	if err != nil {
		logger.Errorf("Visko read U2 failed (%s): %v", ui.adr, err)
		return err
	}
	cmd, err := ui.conn.ReadRegister(16392, modbus.HOLDING_REGISTER)
	if err != nil {
		logger.Errorf("Visko read CMD failed (%s): %v", ui.adr, err)
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
