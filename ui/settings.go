package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func MakeSettingsTab() fyne.CanvasObject {
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

	return form
}
