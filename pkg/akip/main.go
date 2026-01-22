package akip

import (
	"net"
	"time"
)

var connection net.Conn

func StartClient(adr string) error {
	server, err := net.ResolveTCPAddr("tcp", adr)
	connection, err = net.DialTCP("tcp", nil, server)
	return err
}

func Write(cmd string) {
	_, err := connection.Write([]byte(cmd + "\r\n"))
	if err != nil {
		println("Write data failed:", err.Error())
	}
}

func Read() []byte {
	err := connection.SetReadDeadline(time.Now().Add(1000 * time.Millisecond))
	if err != nil {
		return nil
	}
	received := make([]byte, 8024)
	_, err = connection.Read(received)
	if err != nil {
		println("Run data failed:", err.Error())
		return nil
	}
	return received
}

func StopClient() {
	err := connection.Close()
	if err != nil {
		println("Close connection failed:", err.Error())
	}
}
