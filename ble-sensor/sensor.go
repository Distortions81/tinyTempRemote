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
	debugPrintln("  Creating MCP9808 sensor instance...")
	sensor := mcp9808.New(bus)
	sensor.Address = mcp9808Addr
	debugPrint("  Checking sensor connection at address 0x")
	debugPrintHex(uint8(mcp9808Addr))
	debugPrintln("...")
	if !sensor.Connected() {
		debugPrintln("  ERROR: Sensor not connected!")
		return nil, false
	}
	debugPrintln("  Sensor connected successfully")
	debugPrint("  Setting resolution to: ")
	debugPrintInt(int(sensorResolution))
	debugPrintln("")
	if err := sensor.SetResolution(sensorResolution); err != nil {
		debugPrint("  ERROR: Failed to set resolution: ")
		debugPrintln(err.Error())
		return nil, false
	}
	debugPrintln("  Entering shutdown mode (low power)...")
	if err := sensorEnterShutdown(&sensor); err != nil {
		debugPrint("  ERROR: Failed to enter shutdown: ")
		debugPrintln(err.Error())
		return nil, false
	}
	debugPrintln("  Sensor initialization complete")
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
		debugPrintln("ERROR: readSensorTemperature called with nil sensor")
		return 0, newError("sensor missing")
	}
	debugPrintln("  Waking sensor from shutdown...")
	if err := sensorExitShutdown(sensor); err != nil {
		debugPrint("  ERROR: Failed to wake sensor: ")
		debugPrintln(err.Error())
		return 0, err
	}
	debugPrint("  Waiting ")
	debugPrintInt(int(sensorWakeDelayMs))
	debugPrintln("ms for sensor to stabilize...")
	sleepMs(sensorWakeDelayMs)
	debugPrintln("  Reading temperature...")
	temp, err := sensor.ReadTemperature()
	if err != nil {
		debugPrint("  ERROR: Failed to read temperature: ")
		debugPrintln(err.Error())
		_ = sensorEnterShutdown(sensor)
		return 0, err
	}
	debugPrint("  Raw temp: ")
	debugPrintFloat(temp)
	debugPrintln(" C")
	debugPrintln("  Returning sensor to shutdown mode...")
	if err := sensorEnterShutdown(sensor); err != nil {
		debugPrint("  ERROR: Failed to shutdown sensor: ")
		debugPrintln(err.Error())
		return 0, err
	}
	return temp, nil
}
