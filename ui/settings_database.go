package ui

import (
	"DataBridge/serial"
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// initDatabase initializes the SQLite database and creates the table if not exists
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

// insertConfig saves a new configuration to the database
func (sm *SerialManager) insertConfig(cfg serial.PortConfig) error {
	_, err := sm.db.Exec(`INSERT INTO serial_configs (port_name, config_name, baud_rate, data_bits, parity, stop_bits) VALUES (?, ?, ?, ?, ?, ?)`,
		cfg.PortName, cfg.ConfigName, cfg.BaudRate, cfg.DataBits, cfg.Parity, cfg.StopBits)
	return err
}

// loadConfigs loads all configurations from the database
func (sm *SerialManager) loadConfigs() ([]serial.PortConfig, error) {
	rows, err := sm.db.Query(`SELECT port_name, config_name, baud_rate, data_bits, parity, stop_bits FROM serial_configs ORDER BY created_at`)
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

// deleteConfig removes a configuration from the database
func (sm *SerialManager) deleteConfig(cfg serial.PortConfig) error {
	_, err := sm.db.Exec(`DELETE FROM serial_configs WHERE config_name = ?`, cfg.ConfigName)
	return err
}