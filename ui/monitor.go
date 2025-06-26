package ui

import (
	"DataBridge/serial"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// selectedPort holds the currently selected serial port name
var selectedPort string

// MakeMonitorTab creates the Serial Monitor tab UI
// It displays a dropdown to select available serial ports and a refresh button
func MakeMonitorTab() fyne.CanvasObject {
	portList := serial.GetPortList() // Get list of available serial ports
	portNames := []string{}          // Slice to store port names
	for _, p := range portList {
		portNames = append(portNames, p.Name) // Add each port name to the slice
	}
	// Create a dropdown (select) for port selection
	portSelect := widget.NewSelect(portNames, func(value string) {
		selectedPort = value // Update selected port when user selects
	})
	if len(portNames) > 0 {
		portSelect.SetSelected(portNames[0]) // Set default selected port
		selectedPort = portNames[0]
	}
	// Create a refresh button (not yet implemented)
	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		// Not implemented: refresh port list
	})
	// Return the vertical layout containing label, dropdown, and button
	return container.NewVBox(
		widget.NewLabel("Pilih Serial Port yang terhubung:"), // Label for user instruction
		portSelect, // Dropdown for port selection
		refreshBtn, // Button to refresh port list
	)
}
