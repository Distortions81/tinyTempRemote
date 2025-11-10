package main

import "machine"

const (
	mcp9808Addr             = 0x18
	randomSeedPin           = machine.D23
	charHeight              = 18
	offsetIntervalMs  int64 = 10_000
	displayWidth            = 128
	displayHeight           = 32
	sensorPollDelayMs int64 = 2000
	oledSettleDelayMs int64 = 100
)

var displayResetPin machine.Pin = machine.D03
