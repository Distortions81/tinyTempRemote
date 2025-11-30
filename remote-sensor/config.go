package main

import "machine"

const (
	mcp9808Addr                   = 0x18
	randomSeedPin                 = machine.P0_31 // Using analog pin P0_31 for random seed
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
	enableOLED                    = false
	testTxModeEnabled             = false
	testTxIntervalMs        int64 = 1000
	testTxStartTempC              = 20.0
	testTxMaxTempC                = 30.0
	testTxStepTempC               = 0.5
)

var (
	displayResetPin machine.Pin = machine.P0_02  // Display reset pin (analog pin P0_02)
	xbeeTxPin       machine.Pin = machine.P0_08  // UART TX (P0_08)
	xbeeRxPin       machine.Pin = machine.P0_06  // UART RX (P0_06)
	xbeeResetPin    machine.Pin = machine.NoPin
	xbeeSleepPin    machine.Pin = machine.NoPin
	xbeeUART                    = machine.UART0  // nice!nano UART0
)
