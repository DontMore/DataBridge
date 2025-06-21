package ui

import (
	"DataBridge/serial"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func MakeMonitorTab() fyne.CanvasObject {
	portListWidget := widget.NewList(
		func() int { return len(serial.GetPortList()) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(serial.GetPortList()[id])
		},
	)
	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		portListWidget.Refresh()
	})
	return container.NewVBox(
		widget.NewLabel("Available Serial Ports:"),
		portListWidget,
		refreshBtn,
	)
}
