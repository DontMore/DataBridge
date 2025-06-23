package serial

import (
	"fmt"

	"github.com/tarm/serial"
)

type PortInfo struct {
	Name string
}

type PortConfig struct {
	PortName   string
	ConfigName string
	BaudRate   string
	DataBits   string
	Parity     string
	StopBits   string
}

func GetPortList() []PortInfo {
	var portList []PortInfo
	for i := 1; i <= 9; i++ {
		portName := fmt.Sprintf("COM%d", i)
		config := &serial.Config{Name: portName, Baud: 9600}
		port, err := serial.OpenPort(config)
		if err == nil {
			port.Close()
			portList = append(portList, PortInfo{Name: portName})
		}
	}
	return portList
}
