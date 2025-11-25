package main

import "machine"

type xbeeRadio struct {
	uart    *machine.UART
	lineBuf [xbeeLineMaxLen]byte
	lineLen int
}

func newXBeeRadio() *xbeeRadio {
	if xbeeUART == nil {
		return nil
	}

	cfg := machine.UARTConfig{
		BaudRate: xbeeBaudRate,
	}
	if xbeeTxPin != machine.NoPin {
		cfg.TX = xbeeTxPin
	}
	if xbeeRxPin != machine.NoPin {
		cfg.RX = xbeeRxPin
	}
	xbeeUART.Configure(cfg)

	if xbeeResetPin != machine.NoPin {
		xbeeResetPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		xbeeResetPin.High()
		if xbeeResetPulseMs > 0 {
			xbeeResetPin.Low()
			sleepMs(xbeeResetPulseMs)
			xbeeResetPin.High()
			if xbeeBootDelayMs > 0 {
				sleepMs(xbeeBootDelayMs)
			}
		}
	}
	if xbeeSleepPin != machine.NoPin {
		xbeeSleepPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		xbeeSleepPin.Low()
	}

	return &xbeeRadio{uart: xbeeUART}
}

func (x *xbeeRadio) PollLine() (string, bool) {
	if x == nil {
		return "", false
	}

	activity := false
	for x.uart.Buffered() > 0 {
		b, err := x.uart.ReadByte()
		if err != nil {
			break
		}
		activity = true
		switch b {
		case ';', '\n':
			if x.lineLen == 0 {
				continue
			}
			line := string(x.lineBuf[:x.lineLen])
			x.lineLen = 0
			return line, true
		case '\r':
			continue
		default:
			if x.lineLen >= len(x.lineBuf) {
				x.lineLen = 0
				continue
			}
			x.lineBuf[x.lineLen] = b
			x.lineLen++
		}
	}
	if activity && xbeeBlinkDurationMs > 0 {
		blinkOnce(machine.LED, xbeeBlinkDurationMs)
	}
	return "", false
}
