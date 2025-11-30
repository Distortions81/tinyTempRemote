package main

import (
	"machine"

	"tinygo.org/x/drivers/ssd1306"
)

func main() {
	debugPrintln("=== BLE Display Client Starting ===")

	// Using nice!nano hardware I2C0: SDA=P0_17, SCL=P0_20
	debugPrintln("Configuring I2C...")
	debugPrint("  I2C Frequency: ")
	debugPrintInt(int(i2cFreqHz))
	debugPrintln(" Hz")
	i2c := machine.I2C0
	i2c.Configure(machine.I2CConfig{
		Frequency: i2cFreqHz,
	})
	debugPrintln("  I2C configured successfully")

	debugPrintln("Configuring LED...")
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	debugPrintln("  LED configured")

	var (
		display       *ssd1306.Device
		rng           *tinyRNG
		textPos       textOffset
		lastOffsetMs  int64
		lastBounds    rect
		lastText      string
		lastDrawPos   textOffset
		noDataPos     textOffset
		constNoData   = "-- F"
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

	// Initialize BLE client to receive temperature data
	debugPrintln("Initializing BLE client...")
	debugPrint("  Looking for sensor: ")
	debugPrintln(remoteSensorName)
	bleClient := newBLEClient()
	if bleClient != nil {
		debugPrintln("BLE client initialized")
	} else {
		debugPrintln("BLE disabled or failed to initialize")
	}

	debugPrintln("Entering main loop...")
	debugPrint("  Display update interval: ")
	debugPrintInt(int(displayUpdateDelayMs))
	debugPrintln(" ms")
	debugPrint("  LED blink on receive: ")
	debugPrintBool(bleBlinkLEDOnRx)
	debugPrintln("")
	debugPrintln("")

	var loopCount = 0

	for {
		now := millis()
		loopCount++

		if loopCount%20 == 1 {
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

		// Check if we have received temperature data via BLE
		if bleClient != nil && bleClient.HasUpdate() {
			tempText = bleClient.GetLatestTemp()
			bleClient.ClearUpdate()

			debugPrint("Received temp via BLE: ")
			debugPrintln(tempText)

			if bleBlinkLEDOnRx && bleBlinkDurationMs > 0 {
				debugPrintln("  -> Blinking LED on receive")
				blinkOnce(led, bleBlinkDurationMs)
			}

			if display != nil && tempText != "" {
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
		}

		if display != nil {
			if tempText == "" {
				if loopCount%20 == 1 {
					debugPrintln("  Display: No BLE data, showing fallback")
				}
				drawPos = noDataPos
				tempText = constNoData
			}

			currentBounds := textBoundsAt(drawPos, tempText)
			if tempText == lastText && drawPos == lastDrawPos && currentBounds.valid() {
				if loopCount%20 == 1 {
					debugPrintln("  Display: No changes, skipping update")
				}
				sleepIdle(displayUpdateDelayMs)
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
		if loopCount%20 == 1 {
			debugPrint("  Sleeping for ")
			debugPrintInt(int(displayUpdateDelayMs))
			debugPrintln(" ms...")
		}
		sleepIdle(displayUpdateDelayMs)
	}
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
