package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/tarm/serial"
)

func getPortList() []string {
	// Windows COM ports biasanya dari COM1-COM9
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

func makePortMonitorTab() fyne.CanvasObject {
	// Port List with search
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search ports...")

	portListWidget := widget.NewList(
		func() int { return len(getPortList()) },
		func() fyne.CanvasObject { return widget.NewLabel("Template") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(getPortList()[id])
		},
	)

	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		portListWidget.Refresh()
	})

	connectBtn := widget.NewButtonWithIcon("Connect", theme.ComputerIcon(), func() {})

	buttonBox := container.NewHBox(refreshBtn, connectBtn)

	return widget.NewCard(
		"",
		"",
		container.NewVBox(
			searchEntry,
			widget.NewSeparator(),
			portListWidget,
			buttonBox,
		),
	)
}

func makeSettingsTab() fyne.CanvasObject {
	baudRates := []string{"9600", "19200", "38400", "57600", "115200"}
	baudSelect := widget.NewSelect(baudRates, func(value string) {})
	baudSelect.SetSelected("9600")

	dataBits := widget.NewSelect([]string{"5", "6", "7", "8"}, func(value string) {})
	dataBits.SetSelected("8")

	parity := widget.NewSelect([]string{"None", "Even", "Odd"}, func(value string) {})
	parity.SetSelected("None")

	stopBits := widget.NewSelect([]string{"1", "1.5", "2"}, func(value string) {})
	stopBits.SetSelected("1")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Baud Rate", Widget: baudSelect},
			{Text: "Data Bits", Widget: dataBits},
			{Text: "Parity", Widget: parity},
			{Text: "Stop Bits", Widget: stopBits},
		},
	}

	return widget.NewCard("Serial Settings", "", form)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("DataBridge - Serial Monitor")

	// Header
	header := canvas.NewText("DataBridge", theme.PrimaryColor())
	header.TextSize = 24
	header.TextStyle.Bold = true

	var fullscreenBtn *widget.Button
	fullscreenBtn = widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), func() {
		if myWindow.FullScreen() {
			myWindow.SetFullScreen(false)
			fullscreenBtn.SetIcon(theme.ViewFullScreenIcon())
		} else {
			myWindow.SetFullScreen(true)
			fullscreenBtn.SetIcon(theme.ViewRestoreIcon())
		}
	})

	headerContainer := container.NewBorder(nil, nil, nil, fullscreenBtn, header)

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Monitor", theme.MediaPlayIcon(), makePortMonitorTab()),
		container.NewTabItemWithIcon("Settings", theme.SettingsIcon(), makeSettingsTab()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	content := container.NewVBox(
		headerContainer,
		widget.NewSeparator(),
		tabs,
	)

	paddedContent := container.NewPadded(content)
	myWindow.SetContent(paddedContent)
	myWindow.Resize(fyne.NewSize(500, 600))
	myWindow.CenterOnScreen()
	myWindow.ShowAndRun()
}
