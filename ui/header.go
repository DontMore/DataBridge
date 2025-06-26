package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MakeHeader creates the application header and a fullscreen toggle button.
// It returns a styled title text and a button to toggle fullscreen mode.
func MakeHeader(win fyne.Window) (*canvas.Text, *widget.Button) {
	// Create the header text with app name and primary color
	header := canvas.NewText("DataBridge", theme.PrimaryColor())
	header.TextSize = 24         // Set font size
	header.TextStyle.Bold = true // Set text to bold

	var fullscreenBtn *widget.Button // Declare fullscreen button
	fullscreenBtn = widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), func() {
		// Toggle fullscreen mode when button is clicked
		if win.FullScreen() {
			win.SetFullScreen(false)                          // Exit fullscreen
			fullscreenBtn.SetIcon(theme.ViewFullScreenIcon()) // Set icon to fullscreen
		} else {
			win.SetFullScreen(true)                        // Enter fullscreen
			fullscreenBtn.SetIcon(theme.ViewRestoreIcon()) // Set icon to restore
		}
	})

	return header, fullscreenBtn // Return header text and button
}
