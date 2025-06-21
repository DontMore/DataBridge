package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"DataBridge/ui"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("DataBridge - Serial Monitor")

	header, fullscreenBtn := ui.MakeHeader(myWindow)
	tabs := ui.MakeTabs()

	headerContainer := container.NewBorder(nil, nil, nil, fullscreenBtn, header)

	content := container.NewVBox(
		headerContainer,
		widget.NewSeparator(),
		tabs,
	)

	paddedContent := container.NewPadded(content)
	myWindow.SetContent(paddedContent)
	myWindow.Resize(fyne.NewSize(500, 600))
	myWindow.CenterOnScreen()
	myWindow.ShowAndRun()
}
