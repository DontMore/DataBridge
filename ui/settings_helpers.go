package ui

import (
	"fyne.io/fyne/v2"
)

// Constants for better maintainability
const (
	DefaultBaudRate = "9600"
	DefaultDataBits = "8"
	DefaultParity   = "None"
	DefaultStopBits = "1"
	ReadBufferSize  = 128
	MaxTextLength   = 10000 // Limit text to prevent memory issues
)

// getCurrentWindow returns the first available window from the Fyne app.
// Returns nil if no window is found.
func getCurrentWindow() fyne.Window {
	if app := fyne.CurrentApp(); app != nil {
		windows := app.Driver().AllWindows()
		if len(windows) > 0 {
			return windows[0]
		}
	}
	return nil
}