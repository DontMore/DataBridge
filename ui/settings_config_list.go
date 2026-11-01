package ui

import (
	"DataBridge/serial"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// updateConfigList updates the configuration list UI
func (sm *SerialManager) updateConfigList(configList *fyne.Container) {
	sm.mu.RLock()
	configs := make([]serial.PortConfig, len(sm.configs))
	states := make([]bool, len(sm.runningStates))
	copy(configs, sm.configs)
	copy(states, sm.runningStates)
	sm.mu.RUnlock()

	configList.Objects = nil
	if len(configs) == 0 {
		emptyLabel := widget.NewLabel("No saved configurations")
		emptyLabel.Alignment = fyne.TextAlignCenter
		configList.Add(emptyLabel)
	} else {
		for i, cfg := range configs {
			card := sm.createConfigCardSimple(cfg, i, states[i], configList)
			configList.Add(card)
		}
	}
	configList.Refresh()
}

// createConfigCardSimple creates a single configuration card
func (sm *SerialManager) createConfigCardSimple(cfg serial.PortConfig, idx int, isRunning bool, configList *fyne.Container) *widget.Card {
	titleStyle := fyne.TextStyle{Bold: true}

	var actionBtn *widget.Button
	if isRunning {
		actionBtn = widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), func() {
			sm.stopConfiguration(idx, configList)
		})
	} else {
		actionBtn = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), func() {
			sm.startConfiguration(idx, configList)
		})
	}

	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		sm.showDeleteConfirmDialog(cfg, idx, configList)
	})

	viewBtn := widget.NewButtonWithIcon("View", theme.VisibilityIcon(), func() {
		sm.ShowConfigDetailAndMonitor(cfg)
	})

	settingsText := fmt.Sprintf("Settings: %s baud, %s data bits, %s parity, %s stop bits",
		cfg.BaudRate, cfg.DataBits, cfg.Parity, cfg.StopBits)

	card := widget.NewCard("", "", container.NewVBox(
		widget.NewLabelWithStyle(cfg.ConfigName, fyne.TextAlignLeading, titleStyle),
		widget.NewLabel("Port: "+cfg.PortName),
		widget.NewLabel(settingsText),
		container.NewHBox(actionBtn, viewBtn, deleteBtn),
	))

	return card
}

// startConfiguration starts a serial configuration
func (sm *SerialManager) startConfiguration(idx int, configList *fyne.Container) {
	sm.mu.Lock()
	if idx >= len(sm.configs) || idx >= len(sm.runningStates) {
		sm.mu.Unlock()
		return
	}
	cfg := sm.configs[idx]
	sm.runningStates[idx] = true
	sm.mu.Unlock()

	go func() {
		if err := sm.startSerial(cfg); err != nil {
			log.Printf("Failed to start serial port %s: %v", cfg.PortName, err)
			sm.mu.Lock()
			if idx < len(sm.runningStates) {
				sm.runningStates[idx] = false
			}
			sm.mu.Unlock()
			if win := getCurrentWindow(); win != nil {
				dialog.ShowError(fmt.Errorf("failed to start serial port: %w", err), win)
			}
		}
		sm.updateConfigList(configList)
	}()

	sm.updateConfigList(configList)
}

// stopConfiguration stops a serial configuration
func (sm *SerialManager) stopConfiguration(idx int, configList *fyne.Container) {
	sm.mu.Lock()
	if idx >= len(sm.configs) || idx >= len(sm.runningStates) {
		sm.mu.Unlock()
		return
	}
	cfg := sm.configs[idx]
	sm.runningStates[idx] = false
	sm.mu.Unlock()

	sm.stopSerial(cfg)
	sm.updateConfigList(configList)
}

// showDeleteConfirmDialog shows confirmation dialog before deleting
func (sm *SerialManager) showDeleteConfirmDialog(cfg serial.PortConfig, idx int, configList *fyne.Container) {
	win := getCurrentWindow()
	if win == nil {
		return
	}

	dialog.ShowConfirm(
		"Delete Configuration",
		fmt.Sprintf("Are you sure you want to delete the configuration '%s'?", cfg.ConfigName),
		func(confirm bool) {
			if confirm {
				sm.deleteConfiguration(cfg, idx, configList)
			}
		},
		win,
	)
}

// deleteConfiguration deletes a configuration
func (sm *SerialManager) deleteConfiguration(cfg serial.PortConfig, idx int, configList *fyne.Container) {
	// Stop if running
	sm.mu.RLock()
	isRunning := idx < len(sm.runningStates) && sm.runningStates[idx]
	sm.mu.RUnlock()

	if isRunning {
		sm.stopSerial(cfg)
	}

	// Delete from database
	if err := sm.deleteConfig(cfg); err != nil {
		if win := getCurrentWindow(); win != nil {
			dialog.ShowError(err, win)
		}
		return
	}

	// Remove from memory
	sm.mu.Lock()
	if idx < len(sm.configs) {
		sm.configs = append(sm.configs[:idx], sm.configs[idx+1:]...)
	}
	if idx < len(sm.runningStates) {
		sm.runningStates = append(sm.runningStates[:idx], sm.runningStates[idx+1:]...)
	}
	sm.mu.Unlock()

	sm.updateConfigList(configList)
}