package ui

import (
	"DataBridge/serial"
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	tarmserial "github.com/tarm/serial"
)

// createMonitorCard creates the serial monitor UI
func (sm *SerialManager) createMonitorCard() *widget.Card {
	sm.receivedData = widget.NewMultiLineEntry()
	sm.receivedData.SetPlaceHolder("Waiting for data...")
	sm.receivedData.Wrapping = fyne.TextWrapWord
	sm.receivedData.Disable()

	clearBtn := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), func() {
		sm.receivedData.SetText("")
	})

	copyBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if win := getCurrentWindow(); win != nil {
			win.Clipboard().SetContent(sm.receivedData.Text)
			dialog.ShowInformation("Copied", "Data copied to clipboard", win)
		}
	})

	toolbar := container.NewHBox(layout.NewSpacer(), clearBtn, copyBtn)

	bg := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	content := container.NewMax(
		bg,
		container.NewScroll(sm.receivedData),
	)

	card := widget.NewCard("Serial Monitor", "", container.NewBorder(
		nil, toolbar, nil, nil,
		content,
	))

	return card
}

// updateReceivedData safely updates the received data UI
func (sm *SerialManager) updateReceivedData(data string) {
	if sm.receivedData == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formatted := fmt.Sprintf("[%s] %s", timestamp, strings.TrimSpace(data))

	go func() {
		currentText := sm.receivedData.Text
		newText := formatted
		if currentText != "" {
			newText = currentText + "\n" + formatted
		}

		// Limit text length to prevent memory issues
		if len(newText) > MaxTextLength {
			lines := strings.Split(newText, "\n")
			if len(lines) > 100 {
				lines = lines[len(lines)-100:]
				newText = strings.Join(lines, "\n")
			}
		}

		// Gunakan closure untuk menghindari masalah scope dan duplikasi kode
		updateUI := func() {
			sm.receivedData.SetText(newText)
			sm.receivedData.CursorRow = len(strings.Split(newText, "\n")) - 1
		}

		driver := fyne.CurrentApp().Driver()
		if runner, ok := driver.(interface{ RunOnMain(func()) }); ok {
			runner.RunOnMain(updateUI)
		} else {
			updateUI()
		}
	}()
}

// sendNotification sends a system notification
func (sm *SerialManager) sendNotification(data string) {
	if app := fyne.CurrentApp(); app != nil {
		app.SendNotification(&fyne.Notification{
			Title:   "Serial Data Received",
			Content: strings.TrimSpace(data),
		})
	}
}

// ShowConfigDetailAndMonitor displays a new window with config details and data monitor
func (sm *SerialManager) ShowConfigDetailAndMonitor(cfg serial.PortConfig) {
	win := fyne.CurrentApp().NewWindow("Config Detail & Monitor")

	detail := fmt.Sprintf(
		"Name: %s\nPort: %s\nBaud Rate: %s\nData Bits: %s\nParity: %s\nStop Bits: %s",
		cfg.ConfigName, cfg.PortName, cfg.BaudRate, cfg.DataBits, cfg.Parity, cfg.StopBits,
	)
	detailLabel := widget.NewLabel(detail)

	monitorEntry := widget.NewMultiLineEntry()
	monitorEntry.SetPlaceHolder("Waiting for data...")
	monitorEntry.Wrapping = fyne.TextWrapWord
	monitorEntry.Disable()

	isRunning := false
	var stopCh chan struct{}
	var port *tarmserial.Port

	startBtn := widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), func() {
		if isRunning {
			return
		}

		baud, _ := strconv.Atoi(cfg.BaudRate)
		config := &tarmserial.Config{
			Name: cfg.PortName,
			Baud: baud,
		}

		p, err := tarmserial.OpenPort(config)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to open port: %w", err), win)
			return
		}
		port = p
		stopCh = make(chan struct{})
		isRunning = true

		go func() {
			buf := make([]byte, ReadBufferSize)
			for {
				select {
				case <-stopCh:
					return
				default:
					n, err := port.Read(buf)
					if err != nil {
						if strings.Contains(err.Error(), "timeout") {
							continue
						}
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "Serial Error",
							Content: err.Error(),
						})
						return
					}
					if n > 0 {
						data := string(buf[:n])
						timestamp := time.Now().Format("2006-01-02 15:04:05")
						formatted := fmt.Sprintf("[%s] %s", timestamp, strings.TrimSpace(data))

						// Gunakan closure untuk menghindari masalah scope variabel newText
						updateUI := func() {
							currentText := monitorEntry.Text
							newText := formatted
							if currentText != "" {
								newText = currentText + "\n" + formatted
							}
							monitorEntry.SetText(newText)
							monitorEntry.CursorRow = len(strings.Split(newText, "\n")) - 1
						}

						driver := fyne.CurrentApp().Driver()
						if runner, ok := driver.(interface{ RunOnMain(func()) }); ok {
							runner.RunOnMain(updateUI)
						} else {
							updateUI()
						}
					}
				}
			}
		}()
	})

	stopBtn := widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), func() {
		if isRunning && stopCh != nil {
			close(stopCh)
			isRunning = false
			if port != nil {
				port.Close()
			}
		}
	})

	btns := container.NewHBox(startBtn, stopBtn)

	win.SetContent(container.NewVBox(
		widget.NewLabelWithStyle("Configuration Detail", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		detailLabel,
		widget.NewLabelWithStyle("Serial Monitor", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		monitorEntry,
		btns,
	))

	win.Resize(fyne.NewSize(500, 400))
	win.Show()
}