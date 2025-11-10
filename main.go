package main

import (
	"fmt"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

// ---- Bit-banged I2C ----
type SoftI2C struct {
	SDA, SCL machine.Pin
	delay    time.Duration
}

func (i2c *SoftI2C) Configure(freqHz int) {
	i2c.SDA.Configure(machine.PinConfig{Mode: machine.PinOutput})
	i2c.SCL.Configure(machine.PinConfig{Mode: machine.PinOutput})
	i2c.SDA.High()
	i2c.SCL.High()
	i2c.delay = time.Second / time.Duration(freqHz*2)
}

func (i2c *SoftI2C) tick() { time.Sleep(i2c.delay) }

func (i2c *SoftI2C) start() {
	i2c.SDA.High()
	i2c.SCL.High()
	i2c.tick()
	i2c.SDA.Low()
	i2c.tick()
	i2c.SCL.Low()
}

func (i2c *SoftI2C) stop() {
	i2c.SDA.Low()
	i2c.SCL.High()
	i2c.tick()
	i2c.SDA.High()
	i2c.tick()
}

func (i2c *SoftI2C) writeByte(b byte) bool {
	for i := 0; i < 8; i++ {
		if b&0x80 != 0 {
			i2c.SDA.High()
		} else {
			i2c.SDA.Low()
		}
		b <<= 1
		i2c.tick()
		i2c.SCL.High()
		i2c.tick()
		i2c.SCL.Low()
	}
	i2c.SDA.Configure(machine.PinConfig{Mode: machine.PinInput})
	i2c.tick()
	i2c.SCL.High()
	ack := !i2c.SDA.Get()
	i2c.SCL.Low()
	i2c.SDA.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return ack
}

func (i2c *SoftI2C) readByte(ack bool) byte {
	var b byte
	i2c.SDA.Configure(machine.PinConfig{Mode: machine.PinInput})
	for i := 0; i < 8; i++ {
		i2c.SCL.High()
		i2c.tick()
		b = (b << 1)
		if i2c.SDA.Get() {
			b |= 1
		}
		i2c.SCL.Low()
		i2c.tick()
	}
	i2c.SDA.Configure(machine.PinConfig{Mode: machine.PinOutput})
	if ack {
		i2c.SDA.Low()
	} else {
		i2c.SDA.High()
	}
	i2c.SCL.High()
	i2c.tick()
	i2c.SCL.Low()
	i2c.SDA.High()
	return b
}

// ---- I2C-compatible adapter for SSD1306 ----
type softI2CWrapper struct {
	bus *SoftI2C
}

func (s *softI2CWrapper) Tx(addr uint16, w, r []byte) error {
	s.bus.start()
	s.bus.writeByte(byte(addr<<1) | 0)
	for _, b := range w {
		s.bus.writeByte(b)
	}
	if r != nil {
		s.bus.start()
		s.bus.writeByte(byte(addr<<1) | 1)
		for i := range r {
			r[i] = s.bus.readByte(i < len(r)-1)
		}
	}
	s.bus.stop()
	return nil
}

// ---- MCP9808 ----
const MCP9808_ADDR = 0x18

func readTempC(i2c *softI2CWrapper) (float32, bool) {
	buf := []byte{0x05}
	i2c.Tx(MCP9808_ADDR, buf, nil)
	data := make([]byte, 2)
	i2c.Tx(MCP9808_ADDR, nil, data)

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
	i2c := SoftI2C{SDA: machine.D18, SCL: machine.D19}
	i2c.Configure(100000)
	bus := &softI2CWrapper{bus: &i2c}

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	display := ssd1306.NewI2C(bus)
	display.Configure(ssd1306.Config{Width: 128, Height: 32})
	display.ClearDisplay()
	display.Display()

	white := color.RGBA{255, 255, 255, 255}

	for {
		temp, ok := readTempC(bus)
		display.ClearDisplay()
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
