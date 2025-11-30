package main

import (
	"machine"

	"tinygo.org/x/drivers/ssd1306"
)

func main() {
	debugPrintln("=== BLE Display Client Starting ===")

	// Using nice!nano hardware I2C0: SDA=P0_17, SCL=P0_20
	debugPrintln("Configuring I2C...")
	i2c := machine.I2C0
	i2c.Configure(machine.I2CConfig{
		Frequency: i2cFreqHz,
	})

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

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
		resetDisplay(displayResetPin)
		display = ssd1306.NewI2C(i2c)

		sleepMs(oledSettleDelayMs) // let the OLED power rails settle before init
		display.Configure(ssd1306.Config{
			Width:   displayWidth,
			Height:  displayHeight,
			Address: displayI2CAddr,
		})
		contrastOverride := displayContrastOverride
		if contrastOverride >= 0 {
			display.Command(ssd1306.SETCONTRAST)
			display.Command(uint8(contrastOverride))
		}
		display.ClearDisplay()

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
	debugPrint("Looking for sensor: ")
	debugPrintln(remoteSensorName)
	bleClient := newBLEClient()
	if bleClient != nil {
		debugPrintln("BLE client initialized")
	} else {
		debugPrintln("BLE disabled or failed to initialize")
	}

	debugPrintln("Entering main loop...")
	debugPrintln("")

	for {
		now := millis()
		var (
			tempText string
			drawPos  textOffset
		)

		// Check if we have received temperature data via BLE
		if bleClient != nil && bleClient.HasUpdate() {
			tempText = bleClient.GetLatestTemp()
			bleClient.ClearUpdate()

			debugPrint("Received temp: ")
			debugPrintln(tempText)

			if bleBlinkLEDOnRx && bleBlinkDurationMs > 0 {
				blinkOnce(led, bleBlinkDurationMs)
			}

			if display != nil && tempText != "" {
				if now-lastOffsetMs >= offsetIntervalMs {
					jiggleCounter++
					if textJiggleStride > 0 && jiggleCounter >= textJiggleStride {
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
				drawPos = noDataPos
				tempText = constNoData
			}

			currentBounds := textBoundsAt(drawPos, tempText)
			if tempText == lastText && drawPos == lastDrawPos && currentBounds.valid() {
				sleepIdle(displayUpdateDelayMs)
				continue
			}
			if lastBounds.valid() {
				clearRect(display, lastBounds)
			}
			if currentBounds.valid() {
				drawText(display, drawPos.x, drawPos.y, tempText)
				lastBounds = currentBounds
				lastText = tempText
				lastDrawPos = drawPos
			} else {
				lastBounds = rect{}
				lastText = ""
				lastDrawPos = textOffset{}
			}
			flushDirtyPages(display, i2c)
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
