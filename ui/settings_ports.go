package ui

import (
	"DataBridge/serial"
	"log"
	"sort"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.bug.st/serial/enumerator"
)

// GetAllSerialPorts returns all available serial port names.
// It combines results from system enumerator and our custom ListSerialPorts
// to ensure both hardware and virtual ports (e.g., from socat) are detected.
func GetAllSerialPorts() []string {
	portSet := make(map[string]bool) // Use map to avoid duplicates

	// 1. Try using enumerator from go.bug.st/serial (for hardware ports)
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Printf("Warning: enumerator failed: %v, falling back to ListSerialPorts", err)
	} else {
		for _, port := range ports {
			if port.Name != "" {
				portSet[port.Name] = true
			}
		}
	}

	// 2. Add from ListSerialPorts (for virtual ports & fallback)
	customPorts, err := serial.ListSerialPorts()
	if err != nil {
		log.Printf("Warning: ListSerialPorts failed: %v", err)
	} else {
		for _, portName := range customPorts {
			portSet[portName] = true
		}
	}

	// 3. Convert map to slice
	names := make([]string, 0, len(portSet))
	for name := range portSet {
		names = append(names, name)
	}

	// 4. Sort for consistent ordering (improves UX)
	sort.Strings(names)

	return names
}

// createPortSelector creates port selection UI with refresh button
func (sm *SerialManager) createPortSelector() (*widget.Select, *widget.Button) {
	var selectedPort string
	portOptions := GetAllSerialPorts()

	if len(portOptions) > 0 {
		selectedPort = portOptions[0]
	}

	portSelect := widget.NewSelect(portOptions, func(value string) {
		selectedPort = value
	})

	if selectedPort != "" {
		portSelect.SetSelected(selectedPort)
	}

	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		newPorts := GetAllSerialPorts()
		portSelect.Options = newPorts
		if len(newPorts) > 0 {
			portSelect.SetSelected(newPorts[0])
		}
		portSelect.Refresh()
	})

	return portSelect, refreshBtn
}