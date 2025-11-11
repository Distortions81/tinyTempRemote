package main

import (
	"machine"

	"tinygo.org/x/drivers/ssd1306"
)

func main() {
	bitBang := &SoftI2C{SDA: machine.D18, SCL: machine.D19}
	bitBang.Configure(i2cBitBangFreqHz)

	i2c := &softI2CBus{bus: bitBang}

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	resetDisplay(displayResetPin)

	display := ssd1306.NewI2C(i2c)

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

	xbee := newXBeeRadio()

	rng := newTinyRNG(seedEntropy())
	const noDataText = "F0"
	noDataPos := textOffset{x: 16, y: 20}
	displayText := noDataText
	textPos := randomOffset(rng, displayText)
	lastOffsetMs := millis()
	lastBounds := rect{}
	lastText := ""
	lastDrawPos := textOffset{}
	jiggleCounter := 0
	var lastTelemetryMs int64
	hasTelemetry := false

	for {
		now := millis()

		for xbee != nil {
			line, ok := xbee.PollLine()
			if !ok {
				break
			}
			tempF, _, valid := parseTelemetryLine(line)
			if !valid || len(tempF) == 0 {
				continue
			}
			if !hasTelemetry {
				textPos = randomOffset(rng, tempF)
			}
			displayText = tempF
			hasTelemetry = true
			lastTelemetryMs = millis()
			jiggleCounter = 0
			textPos = clampOffsetX(textPos, displayText)
			if xbeeBlinkLEDOnRx && xbeeBlinkDurationMs > 0 {
				blinkOnce(led, xbeeBlinkDurationMs)
			}
		}

		if hasTelemetry && telemetryStaleTimeoutMs > 0 && now-lastTelemetryMs >= telemetryStaleTimeoutMs {
			hasTelemetry = false
			displayText = noDataText
			textPos = clampOffsetX(textPos, displayText)
		}

		drawPos := noDataPos
		if hasTelemetry {
			if now-lastOffsetMs >= offsetIntervalMs {
				jiggleCounter++
				if textJiggleStride > 0 && jiggleCounter >= textJiggleStride {
					textPos = randomOffset(rng, displayText)
					jiggleCounter = 0
				}
				lastOffsetMs = now
			}
			textPos = clampOffsetX(textPos, displayText)
			drawPos = textPos
		}

		currentBounds := textBoundsAt(drawPos, displayText)
		if displayText != lastText || drawPos != lastDrawPos || !currentBounds.valid() {
			if lastBounds.valid() {
				clearRect(display, lastBounds)
			}
			if currentBounds.valid() {
				drawText(display, drawPos.x, drawPos.y, displayText)
				lastBounds = currentBounds
				lastText = displayText
				lastDrawPos = drawPos
			} else {
				lastBounds = rect{}
				lastText = ""
				lastDrawPos = textOffset{}
			}
		}

		flushDirtyPages(display, i2c)
		sleepIdle(telemetryIdleDelayMs)
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
