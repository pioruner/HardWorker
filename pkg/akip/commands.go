package akip

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
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
			ak.linedata, _ = ak.binUnpuck(resp, true)
		} else {
			// текстовый SCPI-ответ
			clean := bytes.TrimRight(resp, "\x00\r\n")
			ak.lastResponse = string(clean)
		}

		giu.Update()
	}
}

func (ak *AkipW) binUnpuck(buf []byte, ch1 bool) ([]int8, float64) {
	size := binary.LittleEndian.Uint16(buf[0:2])
	log.Print("Размер осцилограммы: ")
	log.Println(size)

	dataBuf := buf[12:size]
	log.Printf("Осцилограмма: % X", dataBuf)

	ch1_index := bytes.Index(dataBuf, []byte{0x43, 0x48, 0x31})
	log.Printf("CH1 Index: %d", ch1_index)
	ch2_index := bytes.Index(dataBuf, []byte{0x43, 0x48, 0x32})
	log.Printf("CH2 Index: %d", ch2_index)

	if ch1 {
		return ak.chanelUnpuck(dataBuf, ch1_index)
	} else {
		return ak.chanelUnpuck(dataBuf, ch2_index)
	}

}

func (ak *AkipW) chanelUnpuck(buf []byte, index int) ([]int8, float64) {
	nData := int(binary.LittleEndian.Uint16(buf[index+15 : index+17]))
	hMove := int8(binary.LittleEndian.Uint16(buf[index+31 : index+33]))
	dt := TimeScale[buf[index+27]]
	shift := int(buf[index+59])
	var wave = make([]int8, nData)
	for i := 0; i < nData; i++ {
		if shift+2 > len(buf) {
			return nil, dt
		}
		wave[i] = int8(binary.LittleEndian.Uint16(buf[shift:shift+2])) + hMove
		shift += 2
	}
	return wave[len(wave):], dt
}
