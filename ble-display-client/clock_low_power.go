package main

// Clock management for nRF52840
// The nRF52840 has a simpler clock architecture than NXP Kinetis chips

// forceLowSpeedClock reduces power consumption by lowering the clock speed.
// On nRF52840, the clock system is managed by the CLOCK peripheral and uses
// a high-frequency crystal oscillator (HFXO) or internal RC oscillator (HFINT).
//
// Unlike the NXP Kinetis chips with their complex PLL and FLL configuration,
// nRF52840 clock management is simpler and mostly automatic in TinyGo.
func forceLowSpeedClock() {
	// For nRF52840, clock speed reduction is not as critical for power savings
	// as it is on NXP chips. The main power savings come from sleep modes (WFI/WFE).
	//
	// If you need to reduce power consumption, consider:
	// 1. Using the low-frequency clock for timers when possible
	// 2. Putting the CPU to sleep (handled by sleepIdle)
	// 3. Disabling unused peripherals
	//
	// For now, this is a no-op to maintain API compatibility with the original code.
}
