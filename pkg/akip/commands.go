package akip

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/AllenDang/giu"
)

//:SDSLSCPI#
//:TIMebase:HOFFset %d
//:TIMebase:SCALe %s
//:CHANnel%2$d:SCALe %1$s
//STARTBIN

func (ak *AkipW) sendSCPI(addr, cmd string, timeout time.Duration) ([]byte, error) {
	dialer := net.Dialer{
		Timeout: timeout,
	}

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// write timeout
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))
	_, err = conn.Write([]byte(cmd + "\r\n"))
	if err != nil {
		return nil, err
	}

	// read timeout
	_ = conn.SetReadDeadline(time.Now().Add(timeout))

	buf := make([]byte, 8192)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
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
		// бинарь → hex
		ak.lastResponse = fmt.Sprintf("% X", resp)

		// рабочие данные отдельно
		if len(resp) > 0 {
			ak.linedata = BtoI(resp[:1000])
		} else {
			// текстовый SCPI-ответ
			clean := bytes.TrimRight(resp, "\x00\r\n")
			ak.lastResponse = string(clean)
		}

		giu.Update()
	}
}
