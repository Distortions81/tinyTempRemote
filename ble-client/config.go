package main

import "machine"

const (
	randomSeedPin                 = machine.P0_31 // Using analog pin P0_31 for random seed
	i2cFreqHz                     = 400_000
	charHeight                    = 18
	offsetIntervalMs        int64 = 10_000
	displayWidth                  = 128
	displayHeight                 = 32
	displayI2CAddr                = 0x3C
	displayContrastOverride       = -1 // set 0-255 to override contrast; keep negative to use panel default
	displayUpdateDelayMs    int64 = 1000
	oledSettleDelayMs       int64 = 100
	idleLowPowerMode              = lowPowerModeVLPS
	idleLowPowerMinMs       int64 = 250
	textJiggleStride              = 6
	bleEnabled                    = true
	remoteSensorName              = "TempSensor" // Name of the remote sensor to connect to
	bleBlinkLEDOnRx               = true
	bleBlinkDurationMs      int64 = 100
	enableOLED                    = true // Enable OLED display for client
)

var (
	displayResetPin machine.Pin = machine.P0_02 // Display reset pin (analog pin P0_02)
)
