package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/tarm/serial"
)

var (
	devicePath = flag.String("device", "/dev/ttyUSB1", "serial device to read from")
	baudRate   = flag.Int("baud", 9600, "baud rate for the serial device")
)

func main() {
	flag.Parse()

	if *devicePath == "" {
		log.Fatalf("device path cannot be empty")
	}

	portConfig := &serial.Config{
		Name: *devicePath,
		Baud: *baudRate,
	}
	port, err := serial.OpenPort(portConfig)
	if err != nil {
		log.Fatalf("unable to open %s: %v", *devicePath, err)
	}
	defer port.Close()

	log.Printf("waiting for data from %s", *devicePath)
	reader := bufio.NewReader(port)
	var builder strings.Builder
	for {
		b, err := reader.ReadByte()
		if err != nil {
			log.Fatalf("read error: %v", err)
		}
		switch b {
		case ';', '\n':
			payload := strings.TrimSpace(builder.String())
			if payload == "" {
				log.Println("no payload received")
				return
			}
			fmt.Println(payload)
			return
		case '\r':
			continue
		default:
			builder.WriteByte(b)
		}
	}
}
