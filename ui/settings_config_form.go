package ui

import (
	"DataBridge/serial"
	"strings"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// createConfigCardWithUpdate creates the configuration UI with update callback
func (sm *SerialManager) createConfigCardWithUpdate(updateConfigList func()) *widget.Card {
	// Form components
	portSelect, refreshBtn := sm.createPortSelector()
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter configuration name")

	baudSelect := widget.NewSelect([]string{"9600", "19200", "38400", "57600", "115200"}, nil)
	baudSelect.SetSelected(DefaultBaudRate)

	dataBits := widget.NewSelect([]string{"5", "6", "7", "8"}, nil)
	dataBits.SetSelected(DefaultDataBits)

	parity := widget.NewSelect([]string{"None", "Even", "Odd"}, nil)
	parity.SetSelected(DefaultParity)

	stopBits := widget.NewSelect([]string{"1", "1.5", "2"}, nil)
	stopBits.SetSelected(DefaultStopBits)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "COM Port", Widget: container.NewBorder(nil, nil, nil, refreshBtn, portSelect)},
			{Text: "Configuration Name", Widget: nameEntry},
			{Text: "Baud Rate", Widget: baudSelect},
			{Text: "Data Bits", Widget: dataBits},
			{Text: "Parity", Widget: parity},
			{Text: "Stop Bits", Widget: stopBits},
		},
	}

	saveBtn := widget.NewButtonWithIcon("Save Configuration", theme.DocumentSaveIcon(), func() {
		sm.saveConfiguration(portSelect, nameEntry, baudSelect, dataBits, parity, stopBits)
		if updateConfigList != nil {
			updateConfigList()
		}
	})

	card := widget.NewCard("Serial Port Configuration", "", container.NewVBox(
		form,
		container.NewCenter(saveBtn),
	))

	return card
}

// saveConfiguration saves a new serial configuration
func (sm *SerialManager) saveConfiguration(portSelect *widget.Select, nameEntry *widget.Entry,
	baudSelect, dataBits, parity, stopBits *widget.Select) {

	if portSelect.Selected == "" || strings.TrimSpace(nameEntry.Text) == "" {
		if win := getCurrentWindow(); win != nil {
			dialog.ShowInformation("Warning",
				"Please enter configuration name and select a port", win)
		}
		return
	}

	cfg := serial.PortConfig{
		PortName:   portSelect.Selected,
		ConfigName: strings.TrimSpace(nameEntry.Text),
		BaudRate:   baudSelect.Selected,
		DataBits:   dataBits.Selected,
		Parity:     parity.Selected,
		StopBits:   stopBits.Selected,
	}

	// Check for duplicate names
	sm.mu.RLock()
	for _, existing := range sm.configs {
		if existing.ConfigName == cfg.ConfigName {
			sm.mu.RUnlock()
			if win := getCurrentWindow(); win != nil {
				dialog.ShowInformation("Warning",
					"Configuration name already exists", win)
			}
			return
		}
	}
	sm.mu.RUnlock()

	// Save to database first
	if err := sm.insertConfig(cfg); err != nil {
		if win := getCurrentWindow(); win != nil {
			dialog.ShowError(err, win)
		}
		return
	}

	// Add to memory
	sm.mu.Lock()
	sm.configs = append(sm.configs, cfg)
	sm.runningStates = append(sm.runningStates, false)
	sm.mu.Unlock()

	nameEntry.SetText("")

	if win := getCurrentWindow(); win != nil {
		dialog.ShowInformation("Success", "Configuration saved successfully", win)
	}
}