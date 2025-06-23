package ui

import (
	"DataBridge/serial"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var selectedPort string

func MakeMonitorTab() fyne.CanvasObject {
	portList := serial.GetPortList()
	portNames := []string{}
	for _, p := range portList {
		portNames = append(portNames, p.Name)
	}
	portSelect := widget.NewSelect(portNames, func(value string) {
		selectedPort = value
	})
	if len(portNames) > 0 {
		portSelect.SetSelected(portNames[0])
		selectedPort = portNames[0]
	}
	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		// Not implemented: refresh port list
	})
	return container.NewVBox(
		widget.NewLabel("Pilih Serial Port yang terhubung:"),
		portSelect,
		refreshBtn,
	)
}
