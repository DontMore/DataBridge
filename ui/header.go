package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func MakeHeader(win fyne.Window) (*canvas.Text, *widget.Button) {
	header := canvas.NewText("DataBridge", theme.PrimaryColor())
	header.TextSize = 24
	header.TextStyle.Bold = true

	var fullscreenBtn *widget.Button
	fullscreenBtn = widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), func() {
		if win.FullScreen() {
			win.SetFullScreen(false)
			fullscreenBtn.SetIcon(theme.ViewFullScreenIcon())
		} else {
			win.SetFullScreen(true)
			fullscreenBtn.SetIcon(theme.ViewRestoreIcon())
		}
	})

	return header, fullscreenBtn
}
