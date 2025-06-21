package serial

import (
	"fmt"

	"github.com/tarm/serial"
)

func GetPortList() []string {
	var portList []string
	for i := 1; i <= 9; i++ {
		portName := fmt.Sprintf("COM%d", i)
		config := &serial.Config{Name: portName, Baud: 9600}
		port, err := serial.OpenPort(config)
		if err == nil {
			port.Close()
			portList = append(portList, portName)
		}
	}
	if len(portList) == 0 {
		return []string{"No serial ports found"}
	}
	return portList
}
