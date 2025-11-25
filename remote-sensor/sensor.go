package main

import (
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/mcp9808"
)

const (
	sensorWakeDelayMs int64 = 150
)

var (
	sensorResolution     = mcp9808.High
	sensorWakeConfig     = [...]byte{0x00, 0x00}
	sensorShutdownConfig = [...]byte{0x01, 0x00}
)

func newSensor(bus drivers.I2C) (*mcp9808.Device, bool) {
	sensor := mcp9808.New(bus)
	sensor.Address = mcp9808Addr
	if !sensor.Connected() {
		return nil, false
	}
	if err := sensor.SetResolution(sensorResolution); err != nil {
		return nil, false
	}
	if err := sensorEnterShutdown(&sensor); err != nil {
		return nil, false
	}
	return &sensor, true
}

func sensorEnterShutdown(sensor *mcp9808.Device) error {
	if sensor == nil {
		return newError("sensor missing")
	}
	return sensor.Write(mcp9808.MCP9808_REG_CONFIG, sensorShutdownConfig[:])
}

func sensorExitShutdown(sensor *mcp9808.Device) error {
	if sensor == nil {
		return newError("sensor missing")
	}
	return sensor.Write(mcp9808.MCP9808_REG_CONFIG, sensorWakeConfig[:])
}

func readSensorTemperature(sensor *mcp9808.Device) (float64, error) {
	if sensor == nil {
		return 0, newError("sensor missing")
	}
	if err := sensorExitShutdown(sensor); err != nil {
		return 0, err
	}
	sleepMs(sensorWakeDelayMs)
	temp, err := sensor.ReadTemperature()
	if err != nil {
		_ = sensorEnterShutdown(sensor)
		return 0, err
	}
	if err := sensorEnterShutdown(sensor); err != nil {
		return 0, err
	}
	return temp, nil
}
