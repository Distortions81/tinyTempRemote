package main

import (
	"device/nxp"
)

// forceLowSpeedClock reduces the core clocking by switching to the internal FLL and
// turning off the PLL path. TinyGo expects the core to keep running, so we leave
// the FLL/IRC enabled but disable the PLL and slow the dividers.
func forceLowSpeedClock() {
	// widen dividers before changing the clock source to avoid brief overdrive.
	nxp.SIM.SetCLKDIV1_OUTDIV1(7)
	nxp.SIM.SetCLKDIV1_OUTDIV2(7)
	nxp.SIM.SetCLKDIV1_OUTDIV3(7)
	nxp.SIM.SetCLKDIV1_OUTDIV4(7)

	// Select the internal reference clock (FEI) as the core clock source.
	nxp.MCG.SetC1_IREFSTEN(0)
	nxp.MCG.SetC1_FRDIV(0)
	nxp.MCG.SetC1_CLKS(0) // FLL engaged internal reference

	// Disable PLL output.
	nxp.MCG.SetC5_PLLCLKEN(0)
	nxp.MCG.SetC5_PLLSTEN(0)
	nxp.MCG.SetC6_PLLS(0)

	// disable clock monitor if enabled.
	nxp.MCG.SetC6_CME0(0)

	// Gate down USB clock (redundant if already done) and leave other buses off in sleep.
	nxp.SIM.SetSCGC4_USBOTG(0)
}
