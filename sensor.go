package main

import (
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/mcp9808"
)

func newSensor(bus drivers.I2C) (*mcp9808.Device, bool) {
	sensor := mcp9808.New(bus)
	sensor.Address = mcp9808Addr
	if !sensor.Connected() {
		return nil, false
	}
	if err := sensor.SetResolution(mcp9808.Maximum); err != nil {
		return nil, false
	}
	return &sensor, true
}
