package main

import (
	"tinygo.org/x/bluetooth"
)

var (
	bleAdapter     = bluetooth.DefaultAdapter
	lastTempText   string
	receivedUpdate bool
)

type bleClient struct {
	adapter *bluetooth.Adapter
	enabled bool
}

func newBLEClient() *bleClient {
	if !bleEnabled {
		return nil
	}

	client := &bleClient{
		adapter: bleAdapter,
	}

	// Enable BLE stack
	err := client.adapter.Enable()
	if err != nil {
		return nil
	}

	client.enabled = true

	// Note: BLE scanning will be implemented in the main loop
	// since we're running with scheduler=none (no goroutines)

	return client
}

func (b *bleClient) startScanning() {
	if b == nil || !b.enabled {
		return
	}

	// Start scanning for devices
	// This is a placeholder for BLE scanning functionality
	// In the future, we'll:
	// 1. Scan for devices with name matching remoteSensorName
	// 2. Connect to the device
	// 3. Subscribe to temperature characteristic notifications
	// 4. Update lastTempText when we receive data
}

func (b *bleClient) GetLatestTemp() string {
	if !receivedUpdate {
		return ""
	}
	return lastTempText
}

func (b *bleClient) HasUpdate() bool {
	return receivedUpdate
}

func (b *bleClient) ClearUpdate() {
	receivedUpdate = false
}
