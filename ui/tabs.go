package ui

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

// MakeTabs creates the main application tab layout.
// It returns an AppTabs container with Monitor and Settings tabs.
func MakeTabs() *container.AppTabs {
	return container.NewAppTabs(
		// Tab for Serial Monitor, uses MediaPlayIcon
		container.NewTabItemWithIcon("Monitor", theme.MediaPlayIcon(), MakeMonitorTab()),
		// Tab for Settings, uses SettingsIcon
		container.NewTabItemWithIcon("Settings", theme.SettingsIcon(), MakeSettingsTab()),
	)
}
