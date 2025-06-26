package ui

import (
	"DataBridge/serial"
	"database/sql"
	"fmt"
	"image/color"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas" // add this import for custom background
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	_ "github.com/mattn/go-sqlite3"
	tarmserial "github.com/tarm/serial"
	"go.bug.st/serial/enumerator"
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

// SerialManager manages all serial port operations
type SerialManager struct {
	mu            sync.RWMutex
	configs       []serial.PortConfig
	runningStates []bool
	portMap       map[string]*tarmserial.Port
	stopChMap     map[string]chan struct{}
	db            *sql.DB
	receivedData  *widget.Entry
}

// NewSerialManager creates a new SerialManager instance
func NewSerialManager() *SerialManager {
	return &SerialManager{
		portMap:   make(map[string]*tarmserial.Port),
		stopChMap: make(map[string]chan struct{}),
	}
}

// Initialize initializes the database and loads configurations
func (sm *SerialManager) Initialize() error {
	if err := sm.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	configs, err := sm.loadConfigs()
	if err != nil {
		log.Printf("Warning: failed to load configs: %v", err)
		return nil // Don't fail completely, just log the warning
	}

	sm.mu.Lock()
	sm.configs = configs
	sm.runningStates = make([]bool, len(configs))
	sm.mu.Unlock()

	return nil
}

// Close closes the database connection and stops all running ports
func (sm *SerialManager) Close() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Stop all running ports
	for i, running := range sm.runningStates {
		if running && i < len(sm.configs) {
			sm.stopSerialUnsafe(sm.configs[i])
		}
	}

	// Close database
	if sm.db != nil {
		return sm.db.Close()
	}
	return nil
}

// GetAllSerialPorts returns all available serial port names
func GetAllSerialPorts() []string {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Printf("Error getting serial ports: %v", err)
		return []string{}
	}

	names := make([]string, 0, len(ports))
	for _, port := range ports {
		if port.Name != "" {
			names = append(names, port.Name)
		}
	}
	return names
}

// MakeSettingsTab creates the settings tab UI
func MakeSettingsTab() fyne.CanvasObject {
	manager := NewSerialManager()
	if err := manager.Initialize(); err != nil {
		log.Printf("Failed to initialize serial manager: %v", err)
	}

	// --- Create configList and updateConfigList before configCard ---
	configList := container.NewVBox()
	updateConfigList := func() {
		manager.updateConfigList(configList)
	}

	// --- Pass updateConfigList to configCard and savedConfigsCard ---
	configCard := manager.createConfigCardWithUpdate(updateConfigList)
	savedConfigsCard := widget.NewCard("Saved Configurations", "", container.NewVScroll(configList))
	updateConfigList() // Initial population

	monitorCard := manager.createMonitorCard()

	tabs := container.NewAppTabs(
		container.NewTabItem("Configuration", configCard),
		container.NewTabItem("Saved Configs", savedConfigsCard),
		container.NewTabItem("Serial Monitor", monitorCard),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	return tabs
}

// --- Tambahkan versi baru createSavedConfigsCard yang mengembalikan configList dan menerima updateConfigList ---
func (sm *SerialManager) createSavedConfigsCardWithList(updateConfigList func()) (*widget.Card, *fyne.Container) {
	configList := container.NewVBox()
	updateConfigList() // Initial population
	card := widget.NewCard("Saved Configurations", "", container.NewVScroll(configList))
	return card, configList
}

// --- Modifikasi createConfigCard agar menerima updateConfigList dan memanggilnya setelah save ---
func (sm *SerialManager) createConfigCardWithUpdate(updateConfigList func()) *widget.Card {
	// Form components
	portSelect, refreshBtn := sm.createPortSelector()
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter configuration name")

	baudSelect := widget.NewSelect([]string{"9600", "19200", "38400", "57600", "115200"}, nil)
	baudSelect.SetSelected(DefaultBaudRate)

	dataBits := widget.NewSelect([]string{"5", "6", "7", "8"}, nil)
	dataBits.SetSelected(DefaultDataBits)

	parity := widget.NewSelect([]string{"None", "Even", "Odd"}, nil)
	parity.SetSelected(DefaultParity)

	stopBits := widget.NewSelect([]string{"1", "1.5", "2"}, nil)
	stopBits.SetSelected(DefaultStopBits)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "COM Port", Widget: container.NewBorder(nil, nil, nil, refreshBtn, portSelect)},
			{Text: "Configuration Name", Widget: nameEntry},
			{Text: "Baud Rate", Widget: baudSelect},
			{Text: "Data Bits", Widget: dataBits},
			{Text: "Parity", Widget: parity},
			{Text: "Stop Bits", Widget: stopBits},
		},
	}

	saveBtn := widget.NewButtonWithIcon("Save Configuration", theme.DocumentSaveIcon(), func() {
		sm.saveConfiguration(portSelect, nameEntry, baudSelect, dataBits, parity, stopBits)
		if updateConfigList != nil {
			updateConfigList()
		}
	})

	card := widget.NewCard("Serial Port Configuration", "", container.NewVBox(
		form,
		container.NewCenter(saveBtn),
	))
	return card
}

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

	// White background for serial monitor
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

// createConfigCard creates the configuration UI
func (sm *SerialManager) createConfigCard() *widget.Card {
	// Form components
	portSelect, refreshBtn := sm.createPortSelector()
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter configuration name")

	baudSelect := widget.NewSelect([]string{"9600", "19200", "38400", "57600", "115200"}, nil)
	baudSelect.SetSelected(DefaultBaudRate)

	dataBits := widget.NewSelect([]string{"5", "6", "7", "8"}, nil)
	dataBits.SetSelected(DefaultDataBits)

	parity := widget.NewSelect([]string{"None", "Even", "Odd"}, nil)
	parity.SetSelected(DefaultParity)

	stopBits := widget.NewSelect([]string{"1", "1.5", "2"}, nil)
	stopBits.SetSelected(DefaultStopBits)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "COM Port", Widget: container.NewBorder(nil, nil, nil, refreshBtn, portSelect)},
			{Text: "Configuration Name", Widget: nameEntry},
			{Text: "Baud Rate", Widget: baudSelect},
			{Text: "Data Bits", Widget: dataBits},
			{Text: "Parity", Widget: parity},
			{Text: "Stop Bits", Widget: stopBits},
		},
	}

	saveBtn := widget.NewButtonWithIcon("Save Configuration", theme.DocumentSaveIcon(), func() {
		sm.saveConfiguration(portSelect, nameEntry, baudSelect, dataBits, parity, stopBits)
	})

	card := widget.NewCard("Serial Port Configuration", "", container.NewVBox(
		form,
		container.NewCenter(saveBtn),
	))

	return card
}

// createSavedConfigsCard creates the saved configurations UI
func (sm *SerialManager) createSavedConfigsCard() *widget.Card {
	configList := container.NewVBox()

	updateConfigList := func() {
		sm.updateConfigList(configList)
	}

	updateConfigList() // Initial population

	card := widget.NewCard("Saved Configurations", "", container.NewVScroll(configList))
	return card
}

// createPortSelector creates port selection UI with refresh button
func (sm *SerialManager) createPortSelector() (*widget.Select, *widget.Button) {
	var selectedPort string
	portOptions := GetAllSerialPorts()

	if len(portOptions) > 0 {
		selectedPort = portOptions[0]
	}

	portSelect := widget.NewSelect(portOptions, func(value string) {
		selectedPort = value
	})
	if selectedPort != "" {
		portSelect.SetSelected(selectedPort)
	}

	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		newPorts := GetAllSerialPorts()
		portSelect.Options = newPorts
		if len(newPorts) > 0 {
			portSelect.SetSelected(newPorts[0])
		}
		portSelect.Refresh()
	})

	return portSelect, refreshBtn
}

// saveConfiguration saves a new serial configuration
func (sm *SerialManager) saveConfiguration(portSelect *widget.Select, nameEntry *widget.Entry,
	baudSelect, dataBits, parity, stopBits *widget.Select) {

	if portSelect.Selected == "" || strings.TrimSpace(nameEntry.Text) == "" {
		if win := getCurrentWindow(); win != nil {
			dialog.ShowInformation("Warning",
				"Please enter configuration name and select a port", win)
		}
		return
	}

	cfg := serial.PortConfig{
		PortName:   portSelect.Selected,
		ConfigName: strings.TrimSpace(nameEntry.Text),
		BaudRate:   baudSelect.Selected,
		DataBits:   dataBits.Selected,
		Parity:     parity.Selected,
		StopBits:   stopBits.Selected,
	}

	// Check for duplicate names
	sm.mu.RLock()
	for _, existing := range sm.configs {
		if existing.ConfigName == cfg.ConfigName {
			sm.mu.RUnlock()
			if win := getCurrentWindow(); win != nil {
				dialog.ShowInformation("Warning",
					"Configuration name already exists", win)
			}
			return
		}
	}
	sm.mu.RUnlock()

	// Save to database first
	if err := sm.insertConfig(cfg); err != nil {
		log.Printf("Failed to save config to database: %v", err)
		if win := getCurrentWindow(); win != nil {
			dialog.ShowError(err, win)
		}
		return
	}

	// Add to memory
	sm.mu.Lock()
	sm.configs = append(sm.configs, cfg)
	sm.runningStates = append(sm.runningStates, false)
	sm.mu.Unlock()

	nameEntry.SetText("")
	if win := getCurrentWindow(); win != nil {
		dialog.ShowInformation("Success", "Configuration saved successfully", win)
	}
}

// updateConfigList updates the configuration list UI
func (sm *SerialManager) updateConfigList(configList *fyne.Container) {
	sm.mu.RLock()
	configs := make([]serial.PortConfig, len(sm.configs))
	states := make([]bool, len(sm.runningStates))
	copy(configs, sm.configs)
	copy(states, sm.runningStates)
	sm.mu.RUnlock()

	configList.Objects = nil

	if len(configs) == 0 {
		emptyLabel := widget.NewLabel("No saved configurations")
		emptyLabel.Alignment = fyne.TextAlignCenter
		configList.Add(emptyLabel)
	} else {
		for i, cfg := range configs {
			card := sm.createConfigCardSimple(cfg, i, states[i], configList)
			configList.Add(card)
		}
	}

	configList.Refresh()
}

// createConfigCard creates a single configuration card (no extra arguments)
func (sm *SerialManager) createConfigCardSimple(cfg serial.PortConfig, idx int, isRunning bool, configList *fyne.Container) *widget.Card {
	titleStyle := fyne.TextStyle{Bold: true}

	var actionBtn *widget.Button
	if isRunning {
		actionBtn = widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), func() {
			sm.stopConfiguration(idx, configList)
		})
	} else {
		actionBtn = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), func() {
			sm.startConfiguration(idx, configList)
		})
	}

	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		sm.showDeleteConfirmDialog(cfg, idx, configList)
	})

	settingsText := fmt.Sprintf("Settings: %s baud, %s data bits, %s parity, %s stop bits",
		cfg.BaudRate, cfg.DataBits, cfg.Parity, cfg.StopBits)

	card := widget.NewCard("", "", container.NewVBox(
		widget.NewLabelWithStyle(cfg.ConfigName, fyne.TextAlignLeading, titleStyle),
		widget.NewLabel("Port: "+cfg.PortName),
		widget.NewLabel(settingsText),
		container.NewHBox(actionBtn, deleteBtn),
	))

	return card
}

// startConfiguration starts a serial configuration
func (sm *SerialManager) startConfiguration(idx int, configList *fyne.Container) {
	sm.mu.Lock()
	if idx >= len(sm.configs) || idx >= len(sm.runningStates) {
		sm.mu.Unlock()
		return
	}
	cfg := sm.configs[idx]
	sm.runningStates[idx] = true
	sm.mu.Unlock()

	go func() {
		if err := sm.startSerial(cfg); err != nil {
			log.Printf("Failed to start serial port %s: %v", cfg.PortName, err)
			sm.mu.Lock()
			if idx < len(sm.runningStates) {
				sm.runningStates[idx] = false
			}
			sm.mu.Unlock()

			if win := getCurrentWindow(); win != nil {
				dialog.ShowError(fmt.Errorf("failed to start serial port: %w", err), win)
			}
		}
		sm.updateConfigList(configList)
	}()

	sm.updateConfigList(configList)
}

// stopConfiguration stops a serial configuration
func (sm *SerialManager) stopConfiguration(idx int, configList *fyne.Container) {
	sm.mu.Lock()
	if idx >= len(sm.configs) || idx >= len(sm.runningStates) {
		sm.mu.Unlock()
		return
	}
	cfg := sm.configs[idx]
	sm.runningStates[idx] = false
	sm.mu.Unlock()

	sm.stopSerial(cfg)
	sm.updateConfigList(configList)
}

// startSerial starts reading from a serial port
func (sm *SerialManager) startSerial(cfg serial.PortConfig) error {
	baud, err := strconv.Atoi(cfg.BaudRate)
	if err != nil {
		return fmt.Errorf("invalid baud rate: %w", err)
	}

	config := &tarmserial.Config{
		Name: cfg.PortName,
		Baud: baud,
	}

	port, err := tarmserial.OpenPort(config)
	if err != nil {
		return fmt.Errorf("failed to open port: %w", err)
	}

	stopCh := make(chan struct{})

	sm.mu.Lock()
	// Check if port is already open
	if existingPort, exists := sm.portMap[cfg.PortName]; exists {
		existingPort.Close()
	}
	sm.portMap[cfg.PortName] = port
	sm.stopChMap[cfg.PortName] = stopCh
	sm.mu.Unlock()

	go sm.readSerialData(cfg, port, stopCh)

	return nil
}

// readSerialData reads data from serial port in a goroutine
func (sm *SerialManager) readSerialData(cfg serial.PortConfig, port *tarmserial.Port, stopCh chan struct{}) {
	defer func() {
		sm.mu.Lock()
		delete(sm.portMap, cfg.PortName)
		delete(sm.stopChMap, cfg.PortName)
		sm.mu.Unlock()

		if err := port.Close(); err != nil {
			log.Printf("Error closing port %s: %v", cfg.PortName, err)
		}
	}()

	buf := make([]byte, ReadBufferSize)

	for {
		select {
		case <-stopCh:
			return
		default:
			n, err := port.Read(buf)
			if err != nil {
				// Check if it's a timeout (not a real error)
				if strings.Contains(err.Error(), "timeout") {
					continue
				}
				log.Printf("Error reading from port %s: %v", cfg.PortName, err)
				return
			}

			if n > 0 {
				data := string(buf[:n])
				sm.updateReceivedData(data)
				sm.sendNotification(data)
			}
		}
	}
}

// updateReceivedData safely updates the received data UI
func (sm *SerialManager) updateReceivedData(data string) {
	if sm.receivedData == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formatted := fmt.Sprintf("[%s] %s", timestamp, strings.TrimSpace(data))

	// Use Fyne's thread-safe UI update
	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   "Serial Data Received",
		Content: strings.TrimSpace(data),
	})

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

		// Use Fyne's QueueUpdate or fallback to RunOnMain for compatibility
		driver := fyne.CurrentApp().Driver()
		if runner, ok := driver.(interface{ RunOnMain(func()) }); ok {
			runner.RunOnMain(func() {
				sm.receivedData.SetText(newText)
				sm.receivedData.CursorRow = len(strings.Split(newText, "\n")) - 1
			})
		} else {
			// fallback: just set directly (unsafe, but avoids build error)
			sm.receivedData.SetText(newText)
			sm.receivedData.CursorRow = len(strings.Split(newText, "\n")) - 1
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

// stopSerial stops reading from a serial port
func (sm *SerialManager) stopSerial(cfg serial.PortConfig) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.stopSerialUnsafe(cfg)
}

// stopSerialUnsafe stops serial without locking (assumes lock is held)
func (sm *SerialManager) stopSerialUnsafe(cfg serial.PortConfig) {
	if stopCh, ok := sm.stopChMap[cfg.PortName]; ok {
		close(stopCh)
		delete(sm.stopChMap, cfg.PortName)
	}

	if port, ok := sm.portMap[cfg.PortName]; ok && port != nil {
		port.Close()
		delete(sm.portMap, cfg.PortName)
	}
}

// showDeleteConfirmDialog shows confirmation dialog before deleting
func (sm *SerialManager) showDeleteConfirmDialog(cfg serial.PortConfig, idx int, configList *fyne.Container) {
	win := getCurrentWindow()
	if win == nil {
		return
	}

	dialog.ShowConfirm(
		"Delete Configuration",
		fmt.Sprintf("Are you sure you want to delete the configuration '%s'?", cfg.ConfigName),
		func(confirm bool) {
			if confirm {
				sm.deleteConfiguration(cfg, idx, configList)
			}
		},
		win,
	)
}

// deleteConfiguration deletes a configuration
func (sm *SerialManager) deleteConfiguration(cfg serial.PortConfig, idx int, configList *fyne.Container) {
	// Stop if running
	sm.mu.RLock()
	isRunning := idx < len(sm.runningStates) && sm.runningStates[idx]
	sm.mu.RUnlock()

	if isRunning {
		sm.stopSerial(cfg)
	}

	// Delete from database
	if err := sm.deleteConfig(cfg); err != nil {
		log.Printf("Failed to delete config from database: %v", err)
		if win := getCurrentWindow(); win != nil {
			dialog.ShowError(err, win)
		}
		return
	}

	// Remove from memory
	sm.mu.Lock()
	if idx < len(sm.configs) {
		sm.configs = append(sm.configs[:idx], sm.configs[idx+1:]...)
	}
	if idx < len(sm.runningStates) {
		sm.runningStates = append(sm.runningStates[:idx], sm.runningStates[idx+1:]...)
	}
	sm.mu.Unlock()

	sm.updateConfigList(configList)
}

// Database operations
func (sm *SerialManager) initDatabase() error {
	var err error
	sm.db, err = sql.Open("sqlite3", "serial_configs.db")
	if err != nil {
		return err
	}

	createTable := `CREATE TABLE IF NOT EXISTS serial_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		port_name TEXT NOT NULL,
		config_name TEXT NOT NULL UNIQUE,
		baud_rate TEXT NOT NULL,
		data_bits TEXT NOT NULL,
		parity TEXT NOT NULL,
		stop_bits TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = sm.db.Exec(createTable)
	return err
}

func (sm *SerialManager) insertConfig(cfg serial.PortConfig) error {
	_, err := sm.db.Exec(`INSERT INTO serial_configs 
		(port_name, config_name, baud_rate, data_bits, parity, stop_bits) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		cfg.PortName, cfg.ConfigName, cfg.BaudRate, cfg.DataBits, cfg.Parity, cfg.StopBits)
	return err
}

func (sm *SerialManager) loadConfigs() ([]serial.PortConfig, error) {
	rows, err := sm.db.Query(`SELECT port_name, config_name, baud_rate, data_bits, parity, stop_bits 
		FROM serial_configs ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []serial.PortConfig
	for rows.Next() {
		var cfg serial.PortConfig
		if err := rows.Scan(&cfg.PortName, &cfg.ConfigName, &cfg.BaudRate,
			&cfg.DataBits, &cfg.Parity, &cfg.StopBits); err != nil {
			log.Printf("Error scanning config row: %v", err)
			continue
		}
		configs = append(configs, cfg)
	}

	return configs, rows.Err()
}

func (sm *SerialManager) deleteConfig(cfg serial.PortConfig) error {
	_, err := sm.db.Exec(`DELETE FROM serial_configs WHERE config_name = ?`, cfg.ConfigName)
	return err
}

// Helper function to get current window
func getCurrentWindow() fyne.Window {
	if app := fyne.CurrentApp(); app != nil {
		windows := app.Driver().AllWindows()
		if len(windows) > 0 {
			return windows[0]
		}
	}
	return nil
}
