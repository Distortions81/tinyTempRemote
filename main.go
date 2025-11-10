package main

import (
	"fmt"
	"image/color"
	"machine"
	"math/rand"
	"time"

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
		panic(err)
	}

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	resetDisplay(displayResetPin)

	display := ssd1306.NewI2C(i2c)
	time.Sleep(100 * time.Millisecond) // let the OLED power rails settle before init
	display.Configure(ssd1306.Config{Width: displayWidth, Height: displayHeight, Address: 0x3C})
	display.ClearDisplay()

	white := color.RGBA{255, 255, 255, 255}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	textPos := randomOffset(rng, "0.00 F")
	lastOffset := time.Now()

	for {
		temp, tempErr := sensor.ReadTemperature()
		tempF := temp*9/5 + 32

		display.ClearBuffer()
		if tempErr == nil {
			if enableBlink {
				blinkOnce(led, 100*time.Millisecond)
			}

			tempText := fmt.Sprintf("%.2f F", tempF)
			now := time.Now()
			if now.Sub(lastOffset) >= offsetInterval {
				textPos = randomOffset(rng, tempText)
				lastOffset = now
			}
			textPos = clampOffsetX(textPos, tempText)
			tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, textPos.x, textPos.y, tempText, white)
		} else {
			tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 0, 16,
				"Sensor NACK", white)
		}
		display.Display()
		time.Sleep(2 * time.Second)
	}
}
