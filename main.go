package main

import (
	"fmt"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/mcp9808"
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
	i2c.SDA.Configure(machine.PinConfig{Mode: machine.PinOutputOpenDrain})
	i2c.SCL.Configure(machine.PinConfig{Mode: machine.PinOutputOpenDrain})
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
	i2c.SDA.Configure(machine.PinConfig{Mode: machine.PinOutputOpenDrain})
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
	i2c.SDA.Configure(machine.PinConfig{Mode: machine.PinOutputOpenDrain})
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

// adapters SoftI2C to drivers.I2C
type softI2CBus struct {
	bus *SoftI2C
}

func (s *softI2CBus) Tx(addr uint16, w, r []byte) error {
	if len(w) > 0 {
		s.bus.start()
		if !s.bus.writeByte(byte(addr<<1) | 0) {
			s.bus.stop()
			return fmt.Errorf("i2c addr 0x%02X NACK on write", addr)
		}
		for _, b := range w {
			if !s.bus.writeByte(b) {
				s.bus.stop()
				return fmt.Errorf("i2c write NACK")
			}
		}
		if len(r) == 0 {
			s.bus.stop()
		}
	}
	if len(r) > 0 {
		s.bus.start()
		if !s.bus.writeByte(byte(addr<<1) | 1) {
			s.bus.stop()
			return fmt.Errorf("i2c addr 0x%02X NACK on read", addr)
		}
		for i := range r {
			r[i] = s.bus.readByte(i < len(r)-1)
		}
		s.bus.stop()
	}
	return nil
}

// ---- MCP9808 ----
const mcp9808Addr = 0x18

func newSensor(bus drivers.I2C) (*mcp9808.Device, error) {
	sensor := mcp9808.New(bus)
	sensor.Address = mcp9808Addr
	if !sensor.Connected() {
		return nil, fmt.Errorf("mcp9808 not detected")
	}
	if err := sensor.SetResolution(mcp9808.Maximum); err != nil {
		return nil, err
	}
	return &sensor, nil
}

func blinkOnce(pin machine.Pin, duration time.Duration) {
	pin.High()
	time.Sleep(duration)
	pin.Low()
}

// ---- Main ----
func main() {
	bitBang := &SoftI2C{SDA: machine.D18, SCL: machine.D19}
	bitBang.Configure(100000)
	i2c := &softI2CBus{bus: bitBang}

	sensor, err := newSensor(i2c)
	if err != nil {
		panic(err)
	}

	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	display := ssd1306.NewI2C(i2c)
	time.Sleep(100 * time.Millisecond) // let the OLED power rails settle before init
	display.Configure(ssd1306.Config{Width: 128, Height: 32, Address: 0x3C})
	display.ClearDisplay()

	white := color.RGBA{255, 255, 255, 255}

	for {
		temp, tempErr := sensor.ReadTemperature()
		display.ClearBuffer()
		if tempErr == nil {
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
