//go:build windows

package serial

import (
	"golang.org/x/sys/windows/registry" // Import Windows registry package
)

// ListSerialPorts returns a list of serial port names available on Windows systems.
// It reads the registry key where Windows stores serial port mappings.
func ListSerialPorts() ([]string, error) {
	// Open the registry key that contains serial port information
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DEVICEMAP\SERIALCOMM`, registry.READ)
	if err != nil {
		return nil, err // Return error if the key cannot be opened
	}
	defer k.Close() // Ensure the key is closed after function returns

	// Read all value names under the key
	names, err := k.ReadValueNames(0)
	if err != nil {
		return nil, err // Return error if value names cannot be read
	}

	var ports []string // Slice to store found port names
	for _, name := range names {
		// For each value name, get the corresponding string value (the port name)
		val, _, err := k.GetStringValue(name)
		if err == nil {
			ports = append(ports, val) // Add port name to the list if no error
		}
	}
	return ports, nil // Return the list of port names
}
