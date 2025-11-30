package main

// Power management for nRF52840
// The nRF52840 uses ARM Cortex-M4 with different power modes than NXP Kinetis

const (
	lowPowerModeDisabled = iota
	lowPowerModeStop
	lowPowerModeVLPS
)

func init() {
	// nRF52840 initialization
	// Power management is simpler on nRF52840 - it uses ARM's WFI/WFE instructions
	// No special initialization needed like the NXP Kinetis chips
}

func disableUSBClock() {
	// nRF52840 USB clock management is handled automatically
	// No manual clock gating needed
}

func sleepIdle(ms int64) {
	if ms <= 0 {
		return
	}
	if idleLowPowerMode == lowPowerModeDisabled || ms < idleLowPowerMinMs {
		sleepMs(ms)
		return
	}

	restore := enterIdleLowPowerMode()
	sleepMs(ms)
	restore()
}

func enterIdleLowPowerMode() func() {
	// For nRF52840, low power mode is achieved through WFI (Wait For Interrupt)
	// This is automatically handled by TinyGo's time.Sleep implementation
	// No special register configuration needed like NXP chips

	// Return a no-op restore function
	return func() {}
}
