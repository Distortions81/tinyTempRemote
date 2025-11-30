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
		return nil
	}

	radio := &bleRadio{
		adapter: bleAdapter,
	}

	// Enable BLE stack
	err := radio.adapter.Enable()
	if err != nil {
		return nil
	}

	radio.enabled = true

	// Configure advertising
	adv := radio.adapter.DefaultAdvertisement()
	adv.Configure(bluetooth.AdvertisementOptions{
		LocalName: bleDeviceName,
	})

	// Start advertising
	err = adv.Start()
	if err != nil {
		return nil
	}

	return radio
}

func (b *bleRadio) SendTelemetry(tempC float64) {
	if b == nil || !b.enabled {
		return
	}

	// For now, we're just setting up BLE advertising
	// In the future, we can add a GATT service to expose temperature readings
	// The temperature will be readable via BLE characteristics
}

func (b *bleRadio) SendTextLine(line string) {
	if b == nil || !b.enabled {
		return
	}
	// Placeholder for future BLE GATT notifications
}
