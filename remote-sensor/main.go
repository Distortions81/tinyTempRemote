package main

import (
	"machine"

	"tinygo.org/x/drivers/ssd1306"
)

func main() {
	debugPrintln("=== Remote Sensor Starting ===")

	// Using nice!nano hardware I2C0: SDA=P0_17, SCL=P0_20
	debugPrintln("Configuring I2C...")
	debugPrint("  I2C Frequency: ")
	debugPrintInt(int(i2cBitBangFreqHz))
	debugPrintln(" Hz")
	i2c := machine.I2C0
	i2c.Configure(machine.I2CConfig{
		Frequency: i2cBitBangFreqHz,
	})
	debugPrintln("  I2C configured successfully")

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	debugPrintln("Initializing sensor...")
	sensor, ok := newSensor(i2c)
	if !ok {
		debugPrintln("ERROR: Sensor initialization failed")
		sensor = nil
		blinkError(led)
	} else {
		debugPrintln("Sensor initialized successfully")
	}

	reinitSensor := func() {
		if newSensorInstance, ok := newSensor(i2c); ok {
			sensor = newSensorInstance
			led.Low()
		} else {
			sensor = nil
			blinkError(led)
		}
	}

	var (
		display       *ssd1306.Device
		rng           *tinyRNG
		textPos       textOffset
		lastOffsetMs  int64
		lastBounds    rect
		lastText      string
		lastDrawPos   textOffset
		noDataPos     textOffset
		constNoData   = "F0"
		constFiller   = "00.0 F"
		jiggleCounter = 0
	)

	if enableOLED {
		debugPrintln("Initializing OLED display...")
		debugPrint("  Display dimensions: ")
		debugPrintInt(displayWidth)
		debugPrint("x")
		debugPrintInt(displayHeight)
		debugPrintln("")
		debugPrint("  Display I2C address: 0x")
		debugPrintHex(displayI2CAddr)
		debugPrintln("")
		debugPrintln("  Resetting display...")
		resetDisplay(displayResetPin)
		display = ssd1306.NewI2C(i2c)

		debugPrint("  Waiting ")
		debugPrintInt(int(oledSettleDelayMs))
		debugPrintln("ms for OLED power to settle...")
		sleepMs(oledSettleDelayMs) // let the OLED power rails settle before init
		debugPrintln("  Configuring OLED controller...")
		display.Configure(ssd1306.Config{
			Width:   displayWidth,
			Height:  displayHeight,
			Address: displayI2CAddr,
		})
		contrastOverride := displayContrastOverride
		if contrastOverride >= 0 {
			debugPrint("  Setting contrast override: ")
			debugPrintInt(contrastOverride)
			debugPrintln("")
			display.Command(ssd1306.SETCONTRAST)
			display.Command(uint8(contrastOverride))
		} else {
			debugPrintln("  Using default contrast")
		}
		debugPrintln("  Clearing display...")
		display.ClearDisplay()

		debugPrintln("  Initializing RNG for screen burn-in prevention...")
		rng = newTinyRNG(seedEntropy())
		textPos = randomOffset(rng, constFiller)
		lastOffsetMs = millis()
		noDataPos = textOffset{x: 16, y: 20}
		debugPrintln("OLED display initialized")
	} else {
		debugPrintln("OLED display disabled")
	}

	debugPrintln("Initializing BLE...")
	ble := newBLERadio()
	if ble != nil {
		debugPrintln("BLE initialized and advertising")
	} else {
		debugPrintln("BLE disabled or failed to initialize")
	}

	debugPrintln("Entering main loop...")
	debugPrint("  Sensor poll interval: ")
	debugPrintInt(int(sensorPollDelayMs))
	debugPrintln(" ms")
	if testTxModeEnabled {
		debugPrintln("  Test TX mode ENABLED")
		debugPrint("    Test interval: ")
		debugPrintInt(int(testTxIntervalMs))
		debugPrintln(" ms")
		debugPrint("    Temp range: ")
		debugPrintFloat(testTxStartTempC)
		debugPrint(" C to ")
		debugPrintFloat(testTxMaxTempC)
		debugPrintln(" C")
	}
	debugPrintln("")

	var (
		testTxTempC  = testTxStartTempC
		lastTestTxMs int64
		loopCount    = 0
	)

	const noDataText = "F0"

	for {
		now := millis()
		loopCount++

		if loopCount%10 == 1 {
			debugPrint("[Loop ")
			debugPrintInt(loopCount)
			debugPrint("] Time: ")
			debugPrintInt(int(now))
			debugPrintln(" ms")
		}
		var (
			tempText string
			drawPos  textOffset
		)

		if testTxModeEnabled {
			if now-lastTestTxMs >= testTxIntervalMs {
				debugPrintln("Test TX mode: generating synthetic temperature")
				tempF := testTxTempC*9/5 + 32
				tempText = formatTemp(tempF)
				debugPrint("  Test temp: ")
				debugPrintFloat(testTxTempC)
				debugPrint(" C (")
				debugPrint(tempText)
				debugPrintln(")")
				if ble != nil {
					ble.SendTelemetry(testTxTempC)
					if bleBlinkLEDOnTx && bleBlinkDurationMs > 0 {
						blinkOnce(led, bleBlinkDurationMs)
					}
				}
				lastTestTxMs = now
				testTxTempC += testTxStepTempC
				if testTxTempC >= testTxMaxTempC {
					debugPrintln("  Wrapping test temperature back to start")
					testTxTempC = testTxStartTempC
				}
			}
		} else if sensor != nil {
			debugPrintln("Reading sensor...")
			tempC, tempErr := readSensorTemperature(sensor)
			if tempErr == nil {
				tempF := tempC*9/5 + 32

				tempText = formatTemp(tempF)
				debugPrint("Temp: ")
				debugPrintFloat(tempC)
				debugPrint(" C -> ")
				debugPrint(tempText)
				debugPrintln("")

				if ble != nil {
					ble.SendTelemetry(tempC)
					debugPrintln("  -> Sent via BLE")
					if bleBlinkLEDOnTx && bleBlinkDurationMs > 0 {
						debugPrintln("  -> Blinking LED")
						blinkOnce(led, bleBlinkDurationMs)
					}
				}
				if display != nil {
					if now-lastOffsetMs >= offsetIntervalMs {
						jiggleCounter++
						debugPrint("  Display jiggle check: counter=")
						debugPrintInt(jiggleCounter)
						debugPrintln("")
						if textJiggleStride > 0 && jiggleCounter >= textJiggleStride {
							debugPrintln("  Randomizing display position")
							textPos = randomOffset(rng, tempText)
							jiggleCounter = 0
						}
						lastOffsetMs = now
					}
					textPos = clampOffsetX(textPos, tempText)
					drawPos = textPos
				}
			} else {
				debugPrint("ERROR: Sensor read failed: ")
				debugPrintln(tempErr.Error())
				debugPrintln("  Blinking error LED and reinitializing sensor...")
				blinkError(led)
				reinitSensor()
			}
		}

		if display != nil {
			if tempText == "" {
				debugPrintln("  Display: No data, showing fallback")
				drawPos = noDataPos
				tempText = constNoData
			}

			currentBounds := textBoundsAt(drawPos, tempText)
			if tempText == lastText && drawPos == lastDrawPos && currentBounds.valid() {
				if loopCount%10 == 1 {
					debugPrintln("  Display: No changes, skipping update")
				}
				sleepIdle(sensorPollDelayMs)
				continue
			}
			debugPrintln("  Display: Updating...")
			if lastBounds.valid() {
				debugPrintln("    Clearing old text region")
				clearRect(display, lastBounds)
			}
			if currentBounds.valid() {
				debugPrint("    Drawing text at (")
				debugPrintInt(int(drawPos.x))
				debugPrint(", ")
				debugPrintInt(int(drawPos.y))
				debugPrint("): ")
				debugPrintln(tempText)
				drawText(display, drawPos.x, drawPos.y, tempText)
				lastBounds = currentBounds
				lastText = tempText
				lastDrawPos = drawPos
			} else {
				debugPrintln("    WARNING: Invalid bounds, clearing state")
				lastBounds = rect{}
				lastText = ""
				lastDrawPos = textOffset{}
			}
			debugPrintln("    Flushing display buffer to screen")
			flushDirtyPages(display, i2c)
		}
		sleepDelay := sensorPollDelayMs
		if testTxModeEnabled {
			sleepDelay = testTxIntervalMs
		}
		if loopCount%10 == 1 {
			debugPrint("  Sleeping for ")
			debugPrintInt(int(sleepDelay))
			debugPrintln(" ms...")
		}
		sleepIdle(sleepDelay)
	}
}
func formatTemp(temp float64) string {
	return formatTempWithUnit(temp, 'F')
}

func formatTempWithUnit(temp float64, unit byte) string {
	return buildTempString(temp, unit)
}

func formatTempValue(temp float64) string {
	return buildTempString(temp, 0)
}

func buildTempString(temp float64, unit byte) string {
	scaledValue := temp * 10
	negative := scaledValue < 0
	if negative {
		scaledValue = -scaledValue
	}
	scaled := int32(scaledValue + 0.5)

	whole := scaled / 10
	frac := scaled % 10

	var buf [18]byte
	pos := len(buf)

	if unit != 0 {
		pos--
		buf[pos] = unit
		pos--
		buf[pos] = ' '
	}
	pos--
	buf[pos] = byte('0' + frac)
	pos--
	buf[pos] = '.'

	if whole == 0 {
		pos--
		buf[pos] = '0'
	} else {
		for whole > 0 {
			pos--
			buf[pos] = byte('0' + whole%10)
			whole /= 10
		}
	}

	if negative {
		pos--
		buf[pos] = '-'
	}

	return string(buf[pos:])
}

func seedEntropy() uint32 {
	seed := uint32(millis())
	if randomSeedPin != machine.NoPin {
		randomSeedPin.Configure(machine.PinConfig{Mode: machine.PinInput})
		for i := 0; i < 64; i++ {
			var bit uint32
			if randomSeedPin.Get() {
				bit = 1
			}
			seed ^= bit << (uint(i) % 24)
			sleepMicros(200)
		}
	}
	if seed == 0 {
		seed = machine.CPUFrequency()
	}
	return seed
}
