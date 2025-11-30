// +build debug

package main

import (
	"machine"
)

var debugSerial = machine.Serial

func init() {
	debugSerial.Configure(machine.UARTConfig{BaudRate: 115200})
}

func debugPrint(s string) {
	debugSerial.Write([]byte(s))
}

func debugPrintln(s string) {
	debugSerial.Write([]byte(s))
	debugSerial.Write([]byte("\r\n"))
}
