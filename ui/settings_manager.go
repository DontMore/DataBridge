package ui

import (
	"DataBridge/serial"
	"database/sql"
	"fmt"
	"log"
	"sync"

	tarmserial "github.com/tarm/serial"
	"fyne.io/fyne/v2/widget"
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