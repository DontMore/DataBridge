# DataBridge

DataBridge is a cross-platform desktop application for monitoring and managing serial port communications, built with Go and the Fyne GUI framework. It is designed to help users easily connect to, configure, and monitor serial devices (such as microcontrollers, sensors, and other embedded systems) on Windows and other supported platforms.

## Features

- **Serial Port Scanning:**
  - Automatically detects available serial ports on your system.
  - Supports both Windows registry scanning and cross-platform serial enumeration.

- **Configuration Management:**
  - Save, load, and manage multiple serial port configurations (port, baud rate, data bits, parity, stop bits, and custom names).
  - Configurations are stored persistently using SQLite, so your settings are retained between sessions.

- **Serial Monitor:**
  - Real-time display of incoming serial data with timestamps.
  - Clear and copy received data easily.
  - Notifications for new data received.

- **User Interface:**
  - Modern, clean, and responsive UI built with Fyne.
  - Tabbed interface for easy navigation between monitoring and settings.
  - Fullscreen toggle for focused monitoring.

## How It Works

1. **Port Detection:**
   - On startup, DataBridge scans for available serial ports and lists them for selection.
   - On Windows, it uses the system registry; on other platforms, it uses the `go.bug.st/serial` library.

2. **Configuration:**
   - Users can select a port, set communication parameters, and save configurations for future use.
   - All configurations are stored in a local SQLite database (`serial_configs.db`).

3. **Monitoring:**
   - Select a configuration and start monitoring to view real-time serial data.
   - Data is displayed with timestamps and can be copied or cleared as needed.

## Getting Started

### Prerequisites
- Go 1.23 or newer
- A supported operating system (Windows, Linux, macOS)

### Installation
1. Clone this repository:
   ```sh
   git clone https://github.com/yourusername/DataBridge.git
   cd DataBridge
   ```
2. Download dependencies:
   ```sh
   go mod tidy
   ```
3. Build and run the application:
   ```sh
   go run main.go
   ```
   Or build an executable:
   ```sh
   go build -o DataBridge.exe
   ./DataBridge.exe
   ```

### Usage
- Launch the application.
- Use the **Settings** tab to configure and save serial port settings.
- Use the **Monitor** tab to select a port and view incoming data.
- Use the fullscreen button for distraction-free monitoring.

## Project Structure

- `main.go` — Application entry point and main window setup.
- `ui/` — All UI components (header, tabs, monitor, settings, etc.).
- `serial/` — Serial port handling, configuration structs, and platform-specific code.
- `serial_configs.db` — SQLite database file for saved configurations (created at runtime).

## Dependencies
- [Fyne](https://fyne.io/) — Cross-platform GUI in Go
- [github.com/tarm/serial](https://github.com/tarm/serial) — Serial port library
- [github.com/mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) — SQLite driver
- [go.bug.st/serial](https://github.com/bugst/go-serial) — Cross-platform serial support

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions, issues, and feature requests are welcome! Please open an issue or submit a pull request.

## Author

- [Your Name](https://github.com/yourusername)

---

_DataBridge — Simple, powerful serial port monitoring for everyone._
