package ui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MakeSettingsTab creates the settings tab UI
func MakeSettingsTab() fyne.CanvasObject {
	manager := NewSerialManager()
	if err := manager.Initialize(); err != nil {
		log.Printf("Failed to initialize serial manager: %v", err)
	}

	// Create configList and updateConfigList before configCard
	configList := container.NewVBox()
	updateConfigList := func() {
		manager.updateConfigList(configList)
	}

	// Pass updateConfigList to configCard and savedConfigsCard
	configCard := manager.createConfigCardWithUpdate(updateConfigList)
	savedConfigsCard := widget.NewCard("Saved Configurations", "", container.NewVScroll(configList))
	updateConfigList() // Initial population

	monitorCard := manager.createMonitorCard()

	tabs := container.NewAppTabs(
		container.NewTabItem("Configuration", configCard),
		container.NewTabItem("Saved Configs", savedConfigsCard),
		container.NewTabItem("Serial Monitor", monitorCard),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	return tabs
}