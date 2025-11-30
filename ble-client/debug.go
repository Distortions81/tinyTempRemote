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

func debugPrintInt(val int) {
	initDebugSerial()
	debugSerial.Write([]byte(formatInt(val)))
}

func debugPrintHex(val uint8) {
	initDebugSerial()
	hex := [2]byte{toHexChar(val >> 4), toHexChar(val & 0x0F)}
	debugSerial.Write(hex[:])
}

func debugPrintBool(val bool) {
	initDebugSerial()
	if val {
		debugSerial.Write([]byte("true"))
	} else {
		debugSerial.Write([]byte("false"))
	}
}

func formatInt(val int) string {
	if val == 0 {
		return "0"
	}

	negative := val < 0
	if negative {
		val = -val
	}

	var buf [12]byte
	pos := len(buf)

	for val > 0 {
		pos--
		buf[pos] = byte('0' + val%10)
		val /= 10
	}

	if negative {
		pos--
		buf[pos] = '-'
	}

	return string(buf[pos:])
}

func toHexChar(nibble uint8) byte {
	nibble &= 0x0F
	if nibble < 10 {
		return '0' + nibble
	}
	return 'A' + (nibble - 10)
}
