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
	const (
		noDataText  = "FF"
		initialText = "00"
	)
	noDataPos := textOffset{x: 16, y: 20}
	displayText := initialText
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
				tempC, valid := parseTelemetryLine(line)
				if !valid {
						if len(line) > 0 {
							hasTelemetry = true
							displayText = "FF"
							lastTelemetryMs = millis()
							textPos = clampOffsetX(textPos, displayText)
							if xbeeBlinkDurationMs > 0 {
								blinkOnce(led, xbeeBlinkDurationMs)
							}
						}
						continue
					}

				tempF := tempC*9/5 + 32
				tempText := formatTemp(tempF)
				if !hasTelemetry {
					textPos = randomOffset(rng, tempText)
				}
				displayText = tempText
				hasTelemetry = true
				lastTelemetryMs = millis()
				jiggleCounter = 0
				textPos = clampOffsetX(textPos, displayText)
				if xbeeBlinkDurationMs > 0 {
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
