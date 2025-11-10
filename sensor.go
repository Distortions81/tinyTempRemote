package main

import (
	"fmt"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/mcp9808"
)

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
