package main

import "machine"

type xbeeRadio struct {
	uart *machine.UART
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

func (x *xbeeRadio) SendTelemetry(tempC float64) {
	if x == nil {
		return
	}

	tempCText := formatTempValue(tempC)
	x.sendRecord(tempCText)
}

func (x *xbeeRadio) SendTextLine(line string) {
	if x == nil || len(line) == 0 {
		return
	}
	x.writeAll([]byte(line))
	x.writeAll([]byte(";"))
}

func (x *xbeeRadio) sendRecord(tempCText string) {
	var buf [96]byte
	idx := copy(buf[:], "TEMP,")
	idx += copyAndClamp(buf[idx:], tempCText)
	buf[idx] = ';'
	idx++
	x.writeAll(buf[:idx])
}

func (x *xbeeRadio) writeAll(payload []byte) {
	for len(payload) > 0 {
		written, _ := x.uart.Write(payload)
		if written <= 0 {
			return
		}
		payload = payload[written:]
	}
}

func copyAndClamp(dst []byte, src string) int {
	if len(src) > len(dst) {
		src = src[:len(dst)]
	}
	return copy(dst, src)
}
