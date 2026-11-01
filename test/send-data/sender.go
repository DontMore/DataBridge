package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

func main() {
	// Konfigurasi port
	portName := "/tmp/ttyV1"
	baudRate := 9600

	// Buka port serial
	config := &serial.Config{
		Name: portName,
		Baud: baudRate,
	}
	port, err := serial.OpenPort(config)
	if err != nil {
		log.Fatalf("✗ Error membuka port: %v", err)
	}
	defer port.Close()

	fmt.Printf("✓ Port %s terbuka dengan baud rate %d\n", portName, baudRate)
	fmt.Println("Mengirim data dummy setiap 1 detik... (Ctrl+C untuk berhenti)\n")

	counter := 1
	for {
		// Buat data dummy
		message := fmt.Sprintf("Data dummy #%d - Hello dari Go!\n", counter)

		// Kirim data
		_, err := port.Write([]byte(message))
		if err != nil {
			log.Printf("✗ Error mengirim: %v", err)
			break
		}
		fmt.Printf("→ Terkirim: %s", message)

		counter++
		time.Sleep(1 * time.Second)
	}
}