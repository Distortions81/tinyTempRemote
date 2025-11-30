package main

import (
	"tinygo.org/x/bluetooth"
)

var (
	bleAdapter = bluetooth.DefaultAdapter
)

type bleRadio struct {
	adapter *bluetooth.Adapter
	enabled bool
}

func newBLERadio() *bleRadio {
	if !bleEnabled {
		debugPrintln("  BLE is disabled in config")
		return nil
	}

	debugPrintln("  Creating BLE radio instance...")
	radio := &bleRadio{
		adapter: bleAdapter,
	}

	debugPrintln("  Enabling BLE stack...")
	err := radio.adapter.Enable()
	if err != nil {
		debugPrint("  ERROR: Failed to enable BLE: ")
		debugPrintln(err.Error())
		return nil
	}
	debugPrintln("  BLE stack enabled")

	radio.enabled = true

	debugPrint("  Configuring BLE advertisement (device name: ")
	debugPrint(bleDeviceName)
	debugPrintln(")...")
	adv := radio.adapter.DefaultAdvertisement()
	adv.Configure(bluetooth.AdvertisementOptions{
		LocalName: bleDeviceName,
	})

	debugPrintln("  Starting BLE advertising...")
	err = adv.Start()
	if err != nil {
		debugPrint("  ERROR: Failed to start advertising: ")
		debugPrintln(err.Error())
		return nil
	}
	debugPrintln("  BLE advertising started successfully")

	return radio
}

func (b *bleRadio) SendTelemetry(tempC float64) {
	if b == nil || !b.enabled {
		debugPrintln("  BLE SendTelemetry: radio not available")
		return
	}

	debugPrint("  BLE: Telemetry data ready (")
	debugPrintFloat(tempC)
	debugPrintln(" C)")
	// For now, we're just setting up BLE advertising
	// In the future, we can add a GATT service to expose temperature readings
	// The temperature will be readable via BLE characteristics
	debugPrintln("  BLE: GATT service not yet implemented")
}

func (b *bleRadio) SendTextLine(line string) {
	if b == nil || !b.enabled {
		return
	}
	// Placeholder for future BLE GATT notifications
}
