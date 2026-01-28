package akip

import (
	"bytes"
	"encoding/binary"
	"io"
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
		// бинарь → hex
		//ak.lastResponse = fmt.Sprintf("% X", resp)

		// рабочие данные отдельно
		if len(resp) > 0 {
			ak.linedata, _ = ak.binUnpuck(resp, true)
			ak.plotData = UtoF(ak.linedata)
			//log.Printf("SIZE: %d , WAVE: % X", len(resp), resp)
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
	//log.Printf("Осцилограмма: % X", dataBuf)

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
	log.Printf("NDATA: %d", nData)
	hMove := binary.LittleEndian.Uint16(buf[index+31 : index+33])
	log.Printf("hMove: %d", hMove)
	dt := TimeScale[buf[index+27]]
	log.Printf("dT: %f", dt)
	shift := index + 59
	log.Printf("shift: %d", shift)
	var wave = make([]int8, nData)
	for i := 0; i < nData; i++ {
		wave[i] = int8(binary.LittleEndian.Uint16(buf[shift:shift+2]) + hMove)
		shift += 2
	}
	return wave[len(wave)/2:], dt
}
