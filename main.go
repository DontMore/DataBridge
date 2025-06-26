package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"DataBridge/serial"
	"DataBridge/ui"
)

// getPortList retrieves a list of available serial ports using the serial package.
// If no ports are found or an error occurs, it returns a default message.
func getPortList() []string {
	ports, err := serial.ListSerialPorts()
	if err != nil || len(ports) == 0 {
		return []string{"No serial ports found"} // Return message if no ports
	}
	return ports // Return list of port names
}

// main is the entry point of the application.
// It initializes the Fyne app, creates the main window, header, and tabs,
// and sets up the main layout and window properties.
func main() {
	myApp := app.New()                                         // Create a new Fyne application
	myWindow := myApp.NewWindow("DataBridge - Serial Monitor") // Create the main window

	header, fullscreenBtn := ui.MakeHeader(myWindow) // Create header and fullscreen button
	tabs := ui.MakeTabs()                            // Create the main tab layout

	headerContainer := container.NewBorder(nil, nil, nil, fullscreenBtn, header) // Place header and button

	content := container.NewVBox(
		headerContainer,       // Add header
		widget.NewSeparator(), // Add a separator line
		tabs,                  // Add the main tabs
	)

	paddedContent := container.NewPadded(content) // Add padding around the content
	myWindow.SetContent(paddedContent)            // Set the window content
	myWindow.Resize(fyne.NewSize(500, 600))       // Set initial window size
	myWindow.CenterOnScreen()                     // Center the window on the screen
	myWindow.ShowAndRun()                         // Show the window and start the app event loop
}
