//go:build linux

package serial

import (
	"path/filepath"
)

// ListSerialPorts returns a list of serial port names available on Linux systems.
// It scans common device paths used for USB-to-Serial and hardware serial ports.
func ListSerialPorts() ([]string, error) {
	var ports []string

	// Pola nama device serial yang umum di Linux
	patterns := []string{
		"/dev/ttyUSB*",  // Umum untuk USB-to-Serial (CH340, CP2102, FTDI)
		"/dev/ttyACM*",  // Umum untuk Arduino / CDC devices
		"/dev/ttyS*",    // Serial port hardware bawaan motherboard
		"/tmp/ttyV*",    
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err == nil {
			ports = append(ports, matches...)
		}
	}

	return ports, nil
}