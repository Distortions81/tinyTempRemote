package main

import (
	"machine"
	"time"
)

const (
	mcp9808Addr    = 0x18
	enableBlink    = false
	charWidth      = 6
	charHeight     = 8
	offsetInterval = 10 * time.Second
	displayWidth   = 128
	displayHeight  = 32
)

var displayResetPin machine.Pin = machine.D03
