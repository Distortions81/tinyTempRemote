package main

import "device/nxp"

const (
	lowPowerModeDisabled = iota
	lowPowerModeStop
	lowPowerModeVLPS
)

func init() {
	if idleLowPowerMode != lowPowerModeDisabled {
		// Allow Very-Low-Power modes; this register is write-once after reset.
		nxp.SMC.SetPMPROT_AVLP(1)
	}
	disableUSBClock()
	forceLowSpeedClock()
}

func disableUSBClock() {
	nxp.SIM.SetSCGC4_USBOTG(0)
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
	var stopMode uint8
	switch idleLowPowerMode {
	case lowPowerModeStop:
		stopMode = 0x0
	case lowPowerModeVLPS:
		stopMode = 0x2
	default:
		return func() {}
	}

	prevStop := nxp.SMC.GetPMCTRL_STOPM()
	if prevStop != stopMode {
		nxp.SMC.SetPMCTRL_STOPM(stopMode)
	}

	nxp.SystemControl.SetSCR_SLEEPDEEP(1)

	return func() {
		if prevStop != stopMode {
			nxp.SMC.SetPMCTRL_STOPM(prevStop)
		}
		nxp.SystemControl.SetSCR_SLEEPDEEP(0)
	}
}
