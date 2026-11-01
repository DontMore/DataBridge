package serial

// PortInfo holds information about a serial port (just the name)
type PortInfo struct {
	Name string // Name of the serial port (e.g., COM1 or /dev/ttyUSB0)
}

// PortConfig holds configuration details for a serial port
type PortConfig struct {
	PortName   string // Name of the serial port
	ConfigName string // User-defined configuration name
	BaudRate   string // Baud rate as string (e.g., "9600")
	DataBits   string // Number of data bits (e.g., "8")
	Parity     string // Parity setting (e.g., "None")
	StopBits   string // Number of stop bits (e.g., "1")
}

// GetPortList returns a list of available serial ports.
// It uses the cross-platform ListSerialPorts() function.
func GetPortList() []PortInfo {
	var portList []PortInfo

	// Panggil fungsi ListSerialPorts (akan otomatis menggunakan versi Windows atau Linux)
	ports, err := ListSerialPorts()
	if err != nil {
		return portList // Kembalikan list kosong jika error
	}

	// Konversi hasil string ke dalam struct PortInfo
	for _, portName := range ports {
		portList = append(portList, PortInfo{Name: portName})
	}

	return portList
}