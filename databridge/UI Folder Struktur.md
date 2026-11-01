ui/
├── settings_helpers.go       # Constants & helper functions
├── settings_manager.go       # SerialManager struct, constructor, Initialize, Close
├── settings_ports.go         # GetAllSerialPorts(), createPortSelector()
├── settings_config_form.go   # createConfigCardWithUpdate(), saveConfiguration()
├── settings_config_list.go   # updateConfigList(), start/stop configuration
├── settings_serial.go        # startSerial(), readSerialData(), stopSerial()
├── settings_monitor.go       # createMonitorCard(), updateReceivedData(), ShowConfigDetailAndMonitor()
├── settings_database.go      # Database operations
├── settings_tab.go           # MakeSettingsTab() - entry point
├── header.go                 # (sudah ada)
├── monitor.go                # (sudah ada)
└── tabs.go                   # (sudah ada)