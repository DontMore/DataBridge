package ui

import (
	"DataBridge/serial"
	"fmt"
	"log"
	"strconv"
	"strings"

	tarmserial "github.com/tarm/serial"
)

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