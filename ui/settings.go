package ui

import (
	"DataBridge/serial"
	"database/sql"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	_ "github.com/mattn/go-sqlite3"
	tarmserial "github.com/tarm/serial"
	"go.bug.st/serial/enumerator"
)

// Global variables for serial port configuration and state
var (
	savedConfigs  []serial.PortConfig                 // In-memory list of saved serial port configurations
	runningStates []bool                              // State of each configuration (running or not)
	portMap       = make(map[string]*tarmserial.Port) // Map of open serial ports
	stopChMap     = make(map[string]chan struct{})    // Map of stop channels for goroutines
	mu            sync.Mutex                          // Mutex for thread-safe access
	db            *sql.DB                             // SQLite database connection
)

// Get all available serial port names on the system
func getAllSerialPorts() []string {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return []string{}
	}
	var names []string
	for _, port := range ports {
		names = append(names, port.Name)
	}
	return names
}

// Create the settings tab UI for the application
func MakeSettingsTab() fyne.CanvasObject {
	titleStyle := fyne.TextStyle{Bold: true}     // Style for titles
	sectionStyle := fyne.TextStyle{Italic: true} // Style for section labels

	// --- Serial Monitor Section ---
	monitorCard := widget.NewCard("Serial Monitor", "", nil) // Card for serial monitor
	receivedData := widget.NewMultiLineEntry()               // Multiline entry for received data
	receivedData.SetPlaceHolder("Waiting for data...")
	receivedData.Wrapping = fyne.TextWrapWord
	receivedData.Disable()

	clearBtn := widget.NewButtonWithIcon("Clear", theme.ContentClearIcon(), func() {
		receivedData.SetText("") // Clear the received data
	})

	copyBtn := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
		clipboard.SetContent(receivedData.Text) // Copy data to clipboard
		dialog.ShowInformation("Copied", "Data copied to clipboard",
			fyne.CurrentApp().Driver().AllWindows()[0])
	})

	monitorToolbar := container.NewHBox(
		layout.NewSpacer(),
		clearBtn,
		copyBtn,
	)

	monitorCard.SetContent(container.NewBorder(
		nil,
		monitorToolbar,
		nil,
		nil,
		container.NewScroll(receivedData),
	))

	// --- Port Configuration Section ---
	configCard := widget.NewCard("Serial Port Configuration", "", nil) // Card for port configuration
	baudRates := []string{"9600", "19200", "38400", "57600", "115200"} // Common baud rates
	baudSelect := widget.NewSelect(baudRates, nil)                     // Dropdown for baud rate
	baudSelect.SetSelected("9600")
	baudSelect.PlaceHolder = "Select Baud Rate"

	dataBits := widget.NewSelect([]string{"5", "6", "7", "8"}, nil) // Dropdown for data bits
	dataBits.SetSelected("8")
	dataBits.PlaceHolder = "Select Data Bits"

	parity := widget.NewSelect([]string{"None", "Even", "Odd"}, nil) // Dropdown for parity
	parity.SetSelected("None")
	parity.PlaceHolder = "Select Parity"

	stopBits := widget.NewSelect([]string{"1", "1.5", "2"}, nil) // Dropdown for stop bits
	stopBits.SetSelected("1")
	stopBits.PlaceHolder = "Select Stop Bits"

	portOptions := getAllSerialPorts() // Get available ports
	selectedPort := ""
	if len(portOptions) > 0 {
		selectedPort = portOptions[0]
	}
	portSelect := widget.NewSelect(portOptions, func(value string) {
		selectedPort = value // Update selected port
	})
	portSelect.PlaceHolder = "Select COM Port"
	if len(portOptions) > 0 {
		portSelect.SetSelected(selectedPort)
	}

	refreshPortsBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		portOptions = getAllSerialPorts() // Refresh port list
		portSelect.Options = portOptions
		if len(portOptions) > 0 {
			portSelect.SetSelected(portOptions[0])
		}
		portSelect.Refresh()
	})

	nameEntry := widget.NewEntry() // Entry for configuration name
	nameEntry.SetPlaceHolder("Enter configuration name")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "COM Port", Widget: container.NewBorder(nil, nil, nil, refreshPortsBtn, portSelect)},
			{Text: "Configuration Name", Widget: nameEntry},
			{Text: "Baud Rate", Widget: baudSelect},
			{Text: "Data Bits", Widget: dataBits},
			{Text: "Parity", Widget: parity},
			{Text: "Stop Bits", Widget: stopBits},
		},
	}

	// --- Saved Configurations Section ---
	savedConfigsCard := widget.NewCard("Saved Configurations", "", nil) // Card for saved configs
	configList := container.NewVBox()                                   // List of config cards

	var updateConfigList func() // Function to update config list UI
	updateConfigList = func() {
		configList.Objects = nil
		for i, cfg := range savedConfigs {
			idx := i
			var btn *widget.Button
			if runningStates[idx] {
				btn = widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), func() {
					runningStates[idx] = false
					stopSerial(cfg)
					updateConfigList()
				})
			} else {
				btn = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), func() {
					runningStates[idx] = true
					go startSerial(cfg, receivedData)
					updateConfigList()
				})
			}

			configCard := widget.NewCard("", "", container.NewVBox(
				widget.NewLabelWithStyle(cfg.ConfigName, fyne.TextAlignLeading, titleStyle),
				widget.NewLabel("Port: "+cfg.PortName),
				container.NewHBox(
					widget.NewLabel("Settings: "+cfg.BaudRate+" baud, "+
						cfg.DataBits+" data bits, "+
						cfg.Parity+" parity, "+
						cfg.StopBits+" stop bits"),
				),
				btn,
			))
			configList.Add(configCard)
		}
		if len(savedConfigs) == 0 {
			configList.Add(widget.NewLabelWithStyle("No saved configurations",
				fyne.TextAlignCenter, sectionStyle))
		}
		configList.Refresh()
	}

	saveBtn := widget.NewButtonWithIcon("Save Configuration", theme.DocumentSaveIcon(), func() {
		if selectedPort == "" || nameEntry.Text == "" {
			dialog.ShowInformation("Warning", "Please enter configuration name and select a port",
				fyne.CurrentApp().Driver().AllWindows()[0])
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
		savedConfigs = append(savedConfigs, cfg) // Add to in-memory list
		runningStates = append(runningStates, false)
		_ = insertSerialConfig(cfg) // Save to SQLite database
		nameEntry.SetText("")
		dialog.ShowInformation("Success", "Configuration saved successfully",
			fyne.CurrentApp().Driver().AllWindows()[0])
		updateConfigList()
	})

	configCard.SetContent(container.NewVBox(
		form,
		container.NewCenter(saveBtn),
	))

	savedConfigsCard.SetContent(container.NewVScroll(configList))

	// --- Main Layout ---
	tabs := container.NewAppTabs(
		container.NewTabItem("Configuration", configCard),
		container.NewTabItem("Saved Configs", savedConfigsCard),
		container.NewTabItem("Serial Monitor", monitorCard),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// Initialize database and load configs from SQLite
	if err := initSerialDB(); err == nil {
		configs, err := loadSerialConfigs()
		if err == nil {
			savedConfigs = configs
			runningStates = make([]bool, len(savedConfigs))
		}
	}

	updateConfigList()

	return tabs
}

// Start reading from the serial port and update the UI with received data
func startSerial(cfg serial.PortConfig, receivedData *widget.Entry) {
	baud, err := strconv.Atoi(cfg.BaudRate)
	if err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	c := &tarmserial.Config{
		Name: cfg.PortName,
		Baud: baud,
	}

	s, err := tarmserial.OpenPort(c)
	if err != nil {
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	stopCh := make(chan struct{}) // Channel to signal stop
	mu.Lock()
	portMap[cfg.PortName] = s
	stopChMap[cfg.PortName] = stopCh
	mu.Unlock()

	go func() {
		defer func() {
			mu.Lock()
			delete(portMap, cfg.PortName)
			delete(stopChMap, cfg.PortName)
			mu.Unlock()
			s.Close()
		}()

		buf := make([]byte, 128) // Buffer for reading data
		for {
			select {
			case <-stopCh:
				return
			default:
				n, err := s.Read(buf)
				if err != nil {
					return
				}
				data := string(buf[:n])
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				formatted := "[" + timestamp + "] " + data

				// Show notification for received data
				fyne.CurrentApp().SendNotification(&fyne.Notification{
					Title:   "Serial Data Received",
					Content: data,
				})

				// Update UI safely
				if drv, ok := fyne.CurrentApp().Driver().(interface{ Schedule(func()) }); ok {
					drv.Schedule(func() {
						currentText := receivedData.Text
						if currentText == "" {
							receivedData.SetText(formatted)
						} else {
							receivedData.SetText(currentText + "\n" + formatted)
						}
						receivedData.CursorRow = len(receivedData.Text) - 1
						receivedData.Refresh()
					})
				} else {
					currentText := receivedData.Text
					if currentText == "" {
						receivedData.SetText(formatted)
					} else {
						receivedData.SetText(currentText + "\n" + formatted)
					}
					receivedData.CursorRow = len(receivedData.Text) - 1
					receivedData.Refresh()
				}
			}
		}
	}()
}

// Stop reading from the serial port and clean up resources
func stopSerial(cfg serial.PortConfig) {
	mu.Lock()
	defer mu.Unlock()

	if stopCh, ok := stopChMap[cfg.PortName]; ok {
		close(stopCh)
		delete(stopChMap, cfg.PortName)
	}

	if port, ok := portMap[cfg.PortName]; ok && port != nil {
		port.Close()
		delete(portMap, cfg.PortName)
	}
}

// Initialize the SQLite database and create the table if it doesn't exist
func initSerialDB() error {
	var err error
	db, err = sql.Open("sqlite3", "serial_configs.db")
	if err != nil {
		return err
	}
	createTable := `CREATE TABLE IF NOT EXISTS serial_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		port_name TEXT,
		config_name TEXT,
		baud_rate TEXT,
		data_bits TEXT,
		parity TEXT,
		stop_bits TEXT
	);`
	_, err = db.Exec(createTable)
	return err
}

// Insert a serial port configuration into the database
func insertSerialConfig(cfg serial.PortConfig) error {
	_, err := db.Exec(`INSERT INTO serial_configs (port_name, config_name, baud_rate, data_bits, parity, stop_bits) VALUES (?, ?, ?, ?, ?, ?)`,
		cfg.PortName, cfg.ConfigName, cfg.BaudRate, cfg.DataBits, cfg.Parity, cfg.StopBits)
	return err
}

// Load all serial port configurations from the database
func loadSerialConfigs() ([]serial.PortConfig, error) {
	rows, err := db.Query(`SELECT port_name, config_name, baud_rate, data_bits, parity, stop_bits FROM serial_configs`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var configs []serial.PortConfig
	for rows.Next() {
		var cfg serial.PortConfig
		if err := rows.Scan(&cfg.PortName, &cfg.ConfigName, &cfg.BaudRate, &cfg.DataBits, &cfg.Parity, &cfg.StopBits); err != nil {
			continue
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

// Delete all serial port configurations from the database
func clearSerialConfigs() error {
	_, err := db.Exec(`DELETE FROM serial_configs`)
	return err
}
