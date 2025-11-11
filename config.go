package main

import "machine"

const (
	mcp9808Addr                   = 0x18
	randomSeedPin                 = machine.D23
	i2cBitBangFreqHz              = 400_000
	charHeight                    = 18
	offsetIntervalMs        int64 = 10_000
	displayWidth                  = 128
	displayHeight                 = 32
	displayI2CAddr                = 0x3C
	displayContrastOverride       = -1 // set 0-255 to override contrast; keep negative to use panel default
	sensorPollDelayMs       int64 = 5000
	oledSettleDelayMs       int64 = 100
	idleLowPowerMode              = lowPowerModeVLPS
	idleLowPowerMinMs       int64 = 250
	textJiggleStride              = 6
	xbeeBaudRate                  = 9600
	xbeeResetPulseMs        int64 = 5
	xbeeBootDelayMs         int64 = 50
	xbeeBlinkLEDOnTx              = true
	xbeeBlinkDurationMs     int64 = 100
)

var (
	displayResetPin machine.Pin = machine.D03
	xbeeTxPin       machine.Pin = machine.D10
	xbeeRxPin       machine.Pin = machine.D09
	xbeeResetPin    machine.Pin = machine.NoPin
	xbeeSleepPin    machine.Pin = machine.NoPin
	xbeeUART                    = machine.TeensyUART2
)
