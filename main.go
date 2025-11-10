package main

import (
	"image/color"
	"machine"
	"math/rand"
	"strconv"

	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

func main() {
	bitBang := &SoftI2C{SDA: machine.D18, SCL: machine.D19}
	bitBang.Configure(400000)
	i2c := &softI2CBus{bus: bitBang}

	sensor, err := newSensor(i2c)
	if err != nil {
		sensor = nil
	}

	reinitSensor := func(reason string) {
		if newSensorInstance, sensorErr := newSensor(i2c); sensorErr == nil {
			sensor = newSensorInstance
		} else {
			sensor = nil
		}
	}

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	resetDisplay(displayResetPin)

	display := ssd1306.NewI2C(i2c)
	sleepMs(oledSettleDelayMs) // let the OLED power rails settle before init
	display.Configure(ssd1306.Config{Width: displayWidth, Height: displayHeight, Address: 0x3C})
	display.ClearDisplay()

	white := color.RGBA{255, 255, 255, 255}
	seed := millis()
	if seed == 0 {
		seed = int64(machine.CPUFrequency())
	}
	rng := rand.New(rand.NewSource(seed))
	textPos := randomOffset(rng, "0.00 F")
	lastOffsetMs := millis()

	drawNoData := func() {
		tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 24, 16, "NO DATA", white)
	}

	for {
		display.ClearBuffer()

		if sensor != nil {
			tempC, tempErr := sensor.ReadTemperature()
			if tempErr == nil {
				tempF := tempC*9/5 + 32

				tempText := strconv.FormatFloat(float64(tempF), 'f', 1, 32) + " F"
				nowMs := millis()
				if nowMs-lastOffsetMs >= offsetIntervalMs {
					textPos = randomOffset(rng, tempText)
					lastOffsetMs = nowMs
				}
				textPos = clampOffsetX(textPos, tempText)
				tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, textPos.x, textPos.y, tempText, white)
			} else {
				reinitSensor("read failure")
				drawNoData()
			}
		} else {
			reinitSensor("not initialized")
			drawNoData()
		}
		display.Display()
		sleepMs(sensorPollDelayMs)
	}
}
