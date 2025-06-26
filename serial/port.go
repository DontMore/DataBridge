package serial

import (
	"fmt"

	"github.com/tarm/serial" // Import serial package for serial port operations
)

// PortInfo holds information about a serial port (just the name)
type PortInfo struct {
	Name string // Name of the serial port (e.g., COM1)
}

// PortConfig holds configuration details for a serial port
// This struct is used to store and manage serial port settings
// such as baud rate, data bits, parity, and stop bits
type PortConfig struct {
	PortName   string // Name of the serial port (e.g., COM1)
	ConfigName string // User-defined configuration name
	BaudRate   string // Baud rate as string (e.g., "9600")
	DataBits   string // Number of data bits (e.g., "8")
	Parity     string // Parity setting (e.g., "None")
	StopBits   string // Number of stop bits (e.g., "1")
}

// GetPortList scans COM1 to COM9 and returns a list of available serial ports
// It tries to open each port; if successful, the port is considered available
func GetPortList() []PortInfo {
	var portList []PortInfo // Slice to store available ports
	for i := 1; i <= 9; i++ {
		portName := fmt.Sprintf("COM%d", i)                  // Format port name
		config := &serial.Config{Name: portName, Baud: 9600} // Use default baud rate for test
		port, err := serial.OpenPort(config)                 // Try to open the port
		if err == nil {
			port.Close()                                          // Close if opened successfully
			portList = append(portList, PortInfo{Name: portName}) // Add to list
		}
	}
	return portList // Return all found ports
}
