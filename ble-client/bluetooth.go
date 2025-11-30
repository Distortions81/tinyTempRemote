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
		debugPrintln("  BLE is disabled in config")
		return nil
	}

	debugPrintln("  Creating BLE client instance...")
	client := &bleClient{
		adapter: bleAdapter,
	}

	debugPrintln("  Enabling BLE stack...")
	err := client.adapter.Enable()
	if err != nil {
		debugPrint("  ERROR: Failed to enable BLE: ")
		debugPrintln(err.Error())
		return nil
	}
	debugPrintln("  BLE stack enabled")

	client.enabled = true

	debugPrintln("  BLE client ready for scanning")
	// Note: BLE scanning will be implemented in the main loop
	// since we're running with scheduler=none (no goroutines)

	return client
}

func (b *bleClient) startScanning() {
	if b == nil || !b.enabled {
		debugPrintln("  BLE startScanning: client not available")
		return
	}

	debugPrintln("  BLE: Starting scan for devices...")
	// Start scanning for devices
	// This is a placeholder for BLE scanning functionality
	// In the future, we'll:
	// 1. Scan for devices with name matching remoteSensorName
	// 2. Connect to the device
	// 3. Subscribe to temperature characteristic notifications
	// 4. Update lastTempText when we receive data
	debugPrintln("  BLE: Scanning not yet implemented")
}

func (b *bleClient) GetLatestTemp() string {
	if !receivedUpdate {
		debugPrintln("  BLE GetLatestTemp: no update available")
		return ""
	}
	debugPrint("  BLE GetLatestTemp: returning ")
	debugPrintln(lastTempText)
	return lastTempText
}

func (b *bleClient) HasUpdate() bool {
	hasUpdate := receivedUpdate
	if hasUpdate {
		debugPrintln("  BLE HasUpdate: true")
	}
	return hasUpdate
}

func (b *bleClient) ClearUpdate() {
	debugPrintln("  BLE ClearUpdate: clearing update flag")
	receivedUpdate = false
}
