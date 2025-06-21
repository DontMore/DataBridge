package ui

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

func MakeTabs() *container.AppTabs {
	return container.NewAppTabs(
		container.NewTabItemWithIcon("Monitor", theme.MediaPlayIcon(), MakeMonitorTab()),
		container.NewTabItemWithIcon("Settings", theme.SettingsIcon(), MakeSettingsTab()),
	)
}
