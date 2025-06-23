package ui

import (
	"DataBridge/serial"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	tarmserial "github.com/tarm/serial"
)

var savedConfigs []serial.PortConfig
var runningStates []bool

func MakeSettingsTab() fyne.CanvasObject {
	baudRates := []string{"9600", "19200", "38400", "57600", "115200"}
	baudSelect := widget.NewSelect(baudRates, nil)
	baudSelect.SetSelected("9600")

	dataBits := widget.NewSelect([]string{"5", "6", "7", "8"}, nil)
	dataBits.SetSelected("8")

	parity := widget.NewSelect([]string{"None", "Even", "Odd"}, nil)
	parity.SetSelected("None")

	stopBits := widget.NewSelect([]string{"1", "1.5", "2"}, nil)
	stopBits.SetSelected("1")

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Nama konfigurasi")

	portLabel := widget.NewLabel("Port: " + selectedPort)

	saveBtn := widget.NewButton("Simpan Konfigurasi", func() {
		if selectedPort == "" || nameEntry.Text == "" {
			return
		}
		cfg := serial.PortConfig{
			PortName:   selectedPort,
			ConfigName: nameEntry.Text,
			BaudRate:   baudSelect.Selected,
			DataBits:   dataBits.Selected,
			Parity:     parity.Selected,
			StopBits:   stopBits.Selected,
		}
		savedConfigs = append(savedConfigs, cfg)
		runningStates = append(runningStates, false)
		nameEntry.SetText("")
	})

	configList := container.NewVBox()
	var updateConfigList func()
	updateConfigList = func() {
		configList.Objects = nil
		for i, cfg := range savedConfigs {
			idx := i
			var btn *widget.Button
			if runningStates[idx] {
				btn = widget.NewButton("Stop", func() {
					runningStates[idx] = false
					stopSerial(cfg)
					updateConfigList()
				})
			} else {
				btn = widget.NewButton("Start", func() {
					runningStates[idx] = true
					go startSerial(cfg)
					updateConfigList()
				})
			}
			row := container.NewHBox(
				widget.NewLabel(cfg.ConfigName+" ("+cfg.PortName+") "),
				btn,
			)
			configList.Add(row)
		}
		configList.Refresh()
	}

	saveBtn.OnTapped = func() {
		if selectedPort == "" || nameEntry.Text == "" {
			return
		}
		cfg := serial.PortConfig{
			PortName:   selectedPort,
			ConfigName: nameEntry.Text,
			BaudRate:   baudSelect.Selected,
			DataBits:   dataBits.Selected,
			Parity:     parity.Selected,
			StopBits:   stopBits.Selected,
		}
		savedConfigs = append(savedConfigs, cfg)
		runningStates = append(runningStates, false)
		nameEntry.SetText("")
		updateConfigList()
	}

	updateConfigList()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Port", Widget: portLabel},
			{Text: "Nama", Widget: nameEntry},
			{Text: "Baud Rate", Widget: baudSelect},
			{Text: "Data Bits", Widget: dataBits},
			{Text: "Parity", Widget: parity},
			{Text: "Stop Bits", Widget: stopBits},
		},
	}

	return container.NewVBox(
		form,
		saveBtn,
		widget.NewLabel("Konfigurasi Tersimpan:"),
		configList,
	)
}

func startSerial(cfg serial.PortConfig) {
	baud, _ := strconv.Atoi(cfg.BaudRate)
	c := &tarmserial.Config{
		Name: cfg.PortName,
		Baud: baud,
	}
	s, err := tarmserial.OpenPort(c)
	if err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	go func() {
		buf := make([]byte, 128)
		for {
			n, err := s.Read(buf)
			if err != nil {
				break
			}
			data := string(buf[:n])
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Serial Data Received",
				Content: data,
			})
		}
	}()
}

func stopSerial(cfg serial.PortConfig) {
	// TODO: Implementasi untuk menghentikan pembacaan data serial
}
