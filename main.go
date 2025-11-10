package main

import (
	"machine"

	"tinygo.org/x/drivers/ssd1306"
)

func main() {
	bitBang := &SoftI2C{SDA: machine.D18, SCL: machine.D19}
	bitBang.Configure(400000)
	i2c := &softI2CBus{bus: bitBang}

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	sensor, ok := newSensor(i2c)
	if !ok {
		sensor = nil
		blinkError(led)
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

	resetDisplay(displayResetPin)

	display := ssd1306.NewI2C(i2c)

	sleepMs(oledSettleDelayMs) // let the OLED power rails settle before init
	display.Configure(ssd1306.Config{Width: displayWidth, Height: displayHeight, Address: 0x3C})
	display.ClearDisplay()

	rng := newTinyRNG(seedEntropy())
	textPos := randomOffset(rng, "00.0 F")
	lastOffsetMs := millis()

	drawNoData := func() {
		drawText(display, 16, 20, "F0")
	}

	for {
		display.ClearBuffer()

		if sensor != nil {
			tempC, tempErr := sensor.ReadTemperature()
			if tempErr == nil {
				tempF := tempC*9/5 + 32

				tempText := formatTemp(tempF)
				nowMs := millis()
				if nowMs-lastOffsetMs >= offsetIntervalMs {
					textPos = randomOffset(rng, tempText)
					lastOffsetMs = nowMs
				}
				textPos = clampOffsetX(textPos, tempText)
				drawText(display, textPos.x, textPos.y, tempText)
			} else {
				blinkError(led)
				reinitSensor()
				drawNoData()
			}
		} else {
			reinitSensor()
			drawNoData()
		}
		display.Display()
		sleepMs(sensorPollDelayMs)
	}
}

func formatTemp(temp float64) string {
	scaledValue := temp * 10
	negative := scaledValue < 0
	if negative {
		scaledValue = -scaledValue
	}
	scaled := int32(scaledValue + 0.5)

	whole := scaled / 10
	frac := scaled % 10

	var buf [16]byte
	pos := len(buf)

	pos--
	buf[pos] = 'F'
	pos--
	buf[pos] = ' '
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
