package main

import (
	"fmt"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/i2csoft"
	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

// ---- MCP9808 ----
const MCP9808_ADDR = 0x18

func readTempC(bus drivers.I2C) (float32, bool) {
	data := []byte{0, 0}
	if err := bus.Tx(MCP9808_ADDR, []byte{0x05}, data); err != nil {
		return 0, false
	}

	raw := uint16(data[0])<<8 | uint16(data[1])
	temp := float32(raw&0x0FFF) * 0.0625
	if (raw & 0x1000) != 0 {
		temp -= 256
	}
	return temp, true
}

func blinkOnce(pin machine.Pin, duration time.Duration) {
	pin.High()
	time.Sleep(duration)
	pin.Low()
}

// ---- Main ----
func main() {
	i2c := i2csoft.New(machine.D19, machine.D18)
	i2c.Configure(i2csoft.I2CConfig{Frequency: 100000, SCL: machine.D19, SDA: machine.D18})

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	display := ssd1306.NewI2C(i2c)
	time.Sleep(100 * time.Millisecond) // let the OLED power rails settle before init
	display.Configure(ssd1306.Config{Width: 128, Height: 32, Address: 0x3C})
	display.ClearDisplay()

	white := color.RGBA{255, 255, 255, 255}

	for {
		temp, ok := readTempC(i2c)
		display.ClearBuffer()
		if ok {
			blinkOnce(led, 100*time.Millisecond)
			tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 0, 16,
				fmt.Sprintf("Temp: %.2f C", temp), white)
		} else {
			tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, 0, 16,
				"Sensor NACK", white)
		}
		display.Display()
		time.Sleep(2 * time.Second)
	}
}
