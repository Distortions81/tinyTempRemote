package main

import (
	"fmt"
	"machine"
	"time"
)

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
