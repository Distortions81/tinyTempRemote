package main

import "machine"

const (
	randomSeedPin                 = machine.D23
	i2cBitBangFreqHz              = 400_000
	charHeight                    = 18
	offsetIntervalMs        int64 = 10_000
	displayWidth                  = 128
	displayHeight                 = 32
	displayI2CAddr                = 0x3C
	displayContrastOverride       = -1 // set 0-255 to override contrast; keep negative to use panel default
	oledSettleDelayMs       int64 = 100
	idleLowPowerMode              = lowPowerModeVLPS
	idleLowPowerMinMs       int64 = 250
	textJiggleStride              = 6
	telemetryIdleDelayMs    int64 = 150
	telemetryStaleTimeoutMs int64 = 20_000
	xbeeBaudRate                  = 9600
	xbeeResetPulseMs        int64 = 5
	xbeeBootDelayMs         int64 = 50
	xbeeBlinkLEDOnRx              = true
	xbeeBlinkDurationMs     int64 = 100
	xbeeLineMaxLen                = 96
)

var (
	displayResetPin machine.Pin = machine.D03
	xbeeTxPin       machine.Pin = machine.D10
	xbeeRxPin       machine.Pin = machine.D09
	xbeeResetPin    machine.Pin = machine.NoPin
	xbeeSleepPin    machine.Pin = machine.NoPin
	xbeeUART                    = machine.TeensyUART2
)
