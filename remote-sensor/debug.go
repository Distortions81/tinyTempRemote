// +build debug

package main

import (
	"machine"
	"sync"
)

var (
	debugSerial = machine.Serial
	debugOnce   sync.Once
)

func initDebugSerial() {
	debugOnce.Do(func() {
		// Configure serial at 115200 baud for debug output
		// Note: This uses USB CDC serial, not hardware UART
		debugSerial.Configure(machine.UARTConfig{BaudRate: 115200})
	})
}

func debugPrint(s string) {
	initDebugSerial()
	debugSerial.Write([]byte(s))
}

func debugPrintln(s string) {
	initDebugSerial()
	debugSerial.Write([]byte(s))
	debugSerial.Write([]byte("\r\n"))
}
