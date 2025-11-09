// TinyGo target: nrf52840
// tinygo flash -target=nrf52840
package main

import (
	"device/arm"
	"device/nrf"
	"machine"
	"time"
)

// ==== Pin map (adjust if needed) ====
var (
	// I2C (MCP9808)
	I2C = machine.I2C0 // SDA/SCL default to board pins; override below if needed.
	SDA = machine.SDA_PIN
	SCL = machine.SCL_PIN

	// SPI (CC1101)
	SPI      = machine.SPI0
	PIN_SCK  = machine.SCK_PIN // P0.13 on nRF52840 DK
	PIN_MOSI = machine.SDO_PIN // P0.15
	PIN_MISO = machine.SDI_PIN // P0.14
	PIN_CSN  = machine.Pin(12) // P0.12
	PIN_GDO0 = machine.Pin(11) // P0.11 (IRQ optional)

	// Hourly wake
	wakeSeconds uint32 = 3600
)

// ==== MCP9808 (I2C) ====
const mcpAddr = 0x18

func mcpWrite(reg byte, data ...byte) error {
	buf := append([]byte{reg}, data...)
	return I2C.Tx(mcpAddr, buf, nil)
}
func mcpRead(reg byte, n int) ([]byte, error) {
	out := make([]byte, n)
	if err := I2C.Tx(mcpAddr, []byte{reg}, out); err != nil {
		return nil, err
	}
	return out, nil
}
func mcpInit() {
	I2C.Configure(machine.I2CConfig{SCL: SCL, SDA: SDA})
	// Shutdown sensor initially, set faster conversion (0.25°C)
	_ = mcpWrite(0x01, 0x01, 0x00) // config MSB: SHDN=1
	_ = mcpWrite(0x08, 0x01)       // resolution 0.25°C
}
func mcpOneShotC() (float32, error) {
	// Clear shutdown and set one-shot
	_ = mcpWrite(0x01, 0x00, 0x01)
	time.Sleep(6 * time.Millisecond)
	d, err := mcpRead(0x05, 2)
	if err != nil {
		return 0, err
	}
	raw := uint16(d[0])<<8 | uint16(d[1])
	t := float32(raw&0x0FFF) * 0.0625
	if (raw & 0x1000) != 0 {
		t -= 256
	}
	_ = mcpWrite(0x01, 0x01, 0x00) // back to shutdown
	return t, nil
}

// ==== CC1101 (SPI) ====
const (
	CC1101_WRITE_BURST = 0x40
	CC1101_READ_SINGLE = 0x80
	CC1101_READ_BURST  = 0xC0

	// Strobes
	SRES    = 0x30
	SFSTXON = 0x31
	SXOFF   = 0x32
	SCAL    = 0x33
	SRX     = 0x34
	STX     = 0x35
	SIDLE   = 0x36
	SWOR    = 0x38
	SPWD    = 0x39
	SFRX    = 0x3A
	SFTX    = 0x3B
	SNOP    = 0x3D

	// FIFOs
	TXFIFO = 0x3F
	RXFIFO = 0x3F

	// PATABLE addr
	PATABLE = 0x3E
)

// 915 MHz, ~38.4 kbps GFSK, CRC on, variable length.
// Values mirror SmartRF Studio style defaults for 26 MHz ref.
// Adjust if your module differs.
var cc1101RegTable = [][2]byte{
	{0x00, 0x06}, // IOCFG2  (GDO2)  - not used
	{0x02, 0x06}, // IOCFG0  (GDO0)  - assert on SYNC, de-assert end of packet
	{0x03, 0x07}, // FIFOTHR - FIFO threshold
	{0x06, 0x06}, // PKTLEN  - max length when variable length used
	{0x07, 0x04}, // PKTCTRL1 - no address check
	{0x08, 0x05}, // PKTCTRL0 - CRC on, variable length
	{0x0B, 0x06}, // FSCTRL1 - IF freq
	{0x0C, 0x00}, // FSCTRL0
	{0x0D, 0x23}, // FREQ2   - 915 MHz
	{0x0E, 0x31}, // FREQ1
	{0x0F, 0x3B}, // FREQ0
	{0x10, 0xCA}, // MDMCFG4 - BW/DR settings (for ~38.4 kbps)
	{0x11, 0x83}, // MDMCFG3 - data rate mantissa
	{0x12, 0x73}, // MDMCFG2 - GFSK, 30/32 sync, no manchester
	{0x13, 0x22}, // MDMCFG1 - FEC off, 4 preamble bytes
	{0x14, 0xF8}, // MDMCFG0 - channel spacing
	{0x15, 0x35}, // DEVIATN - ~20.6 kHz
	{0x18, 0x16}, // MCSM0   - FS Autocal
	{0x19, 0x16}, // FOCCFG
	{0x1A, 0x6C}, // BSCFG
	{0x1B, 0x43}, // AGCCTRL2
	{0x1C, 0x40}, // AGCCTRL1
	{0x1D, 0x91}, // AGCCTRL0
	{0x21, 0x56}, // FREND1
	{0x22, 0x10}, // FREND0
	{0x23, 0xE9}, // FSCAL3
	{0x24, 0x2A}, // FSCAL2
	{0x25, 0x00}, // FSCAL1
	{0x26, 0x1F}, // FSCAL0
	{0x29, 0x59}, // FSTEST (typical)
	{0x2C, 0x81}, // TEST2
	{0x2D, 0x35}, // TEST1
	{0x2E, 0x09}, // TEST0
}

// PA table entries. For 915 MHz on CC1101, 0xC8 is commonly ~+5 dBm on many modules.
// 0x84 is around 0 dBm. Actual level depends on board/RF front-end.
// Keep within local regulations.
var paTable5dBm = []byte{0xC8} // +~5 dBm
var paTable0dBm = []byte{0x84} // ~0 dBm

func spiInit() {
	SPI.Configure(machine.SPIConfig{
		Frequency: 8_000_000,
		Mode:      0,
		SCK:       PIN_SCK,
		SDO:       PIN_MOSI,
		SDI:       PIN_MISO,
	})
	PIN_CSN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	PIN_CSN.High()
	PIN_GDO0.Configure(machine.PinConfig{Mode: machine.PinInput})
}

func ccWriteReg(addr, val byte) {
	PIN_CSN.Low()
	SPI.Transfer(addr)
	SPI.Transfer(val)
	PIN_CSN.High()
}
func ccBurstWrite(addr byte, data []byte) {
	PIN_CSN.Low()
	SPI.Transfer(addr | CC1101_WRITE_BURST)
	for _, b := range data {
		SPI.Transfer(b)
	}
	PIN_CSN.High()
}
func ccStrobe(cmd byte) {
	PIN_CSN.Low()
	SPI.Transfer(cmd)
	PIN_CSN.High()
}
func ccReadStatus(addr byte) byte {
	PIN_CSN.Low()
	SPI.Transfer(addr | CC1101_READ_SINGLE)
	v, _ := SPI.Transfer(0x00)
	PIN_CSN.High()
	return v
}

func ccInit(pa5dBm bool) {
	// Reset
	ccStrobe(SRES)
	time.Sleep(1 * time.Millisecond)

	// Write registers
	for _, kv := range cc1101RegTable {
		ccWriteReg(kv[0], kv[1])
	}

	// Set PA table
	if pa5dBm {
		ccBurstWrite(PATABLE, paTable5dBm)
	} else {
		ccBurstWrite(PATABLE, paTable0dBm)
	}

	// Calibrate and idle
	ccStrobe(SCAL)
	time.Sleep(1 * time.Millisecond)
	ccStrobe(SIDLE)
	ccStrobe(SFTX)
}

func ccSend(pkt []byte) {
	// Variable-length: write length first
	ccStrobe(SIDLE)
	ccStrobe(SFTX)
	PIN_CSN.Low()
	SPI.Transfer(TXFIFO | CC1101_WRITE_BURST)
	SPI.Transfer(byte(len(pkt)))
	for _, b := range pkt {
		SPI.Transfer(b)
	}
	PIN_CSN.High()

	// TX
	ccStrobe(STX)

	// Option A: wait on GDO0 de-assert (end-of-packet)
	// Simple polling with timeout.
	deadline := time.Now().Add(200 * time.Millisecond)
	// Wait for assert
	for !PIN_GDO0.Get() {
		if time.Now().After(deadline) {
			break
		}
	}
	// Wait for de-assert
	for PIN_GDO0.Get() {
		if time.Now().After(deadline) {
			break
		}
	}
	ccStrobe(SIDLE)
}

// ==== RTC1 timed sleep (System-ON idle with WFI) ====
//
// LFCLK: start 32.768 kHz
// RTC1: prescaler 32767 => 1 Hz tick. Compare = now + seconds.
func rtcInit() {
	// Start LFCLK (LFXO if present, else LFRC)
	nrf.CLOCK.LFCLKSRC.Set(nrf.CLOCK_LFCLKSRC_SRC_Synth) // fallback safe for dev boards; use Xtal if fitted
	nrf.CLOCK.EVENTS_LFCLKSTART.Set(0)
	nrf.CLOCK.TASKS_LFCLKSTART.Set(1)
	for nrf.CLOCK.EVENTS_LFCLKSTART.Get() == 0 {
		// wait
	}
	// Configure RTC1
	nrf.RTC1.TASKS_STOP.Set(1)
	nrf.RTC1.PRESCALER.Set(32767) // 1 Hz
	nrf.RTC1.COUNTER.Set(0)
	// Enable interrupt on COMPARE[0]
	nrf.RTC1.EVTENSET.Set(nrf.RTC_EVTENSET_COMPARE0_Enabled)
	nrf.RTC1.INTENSET.Set(nrf.RTC_INTENSET_COMPARE0_Enabled)
	arm.EnableIRQ(nrf.IRQ_RTC1)
	nrf.RTC1.TASKS_START.Set(1)
}

var rtcFired volatileBool

type volatileBool struct{ v uint32 }

func (b *volatileBool) Set(x bool) {
	if x {
		b.v = 1
	} else {
		b.v = 0
	}
}
func (b *volatileBool) Get() bool { return b.v != 0 }

func sleepSeconds(sec uint32) {
	// Set compare = now + sec
	now := nrf.RTC1.COUNTER.Get()
	nrf.RTC1.CC[0].Set((now + sec) & 0x00FFFFFF)
	nrf.RTC1.EVENTS_COMPARE[0].Set(0)
	rtcFired.Set(false)

	// WFI loop until compare event
	for !rtcFired.Get() {
		arm.Asm("wfi")
	}
}

func main() {
	// Buses
	I2C.Configure(machine.I2CConfig{SCL: SCL, SDA: SDA})
	spiInit()
	mcpInit()
	ccInit(true) // true => +~5 dBm PA entry

	// RTC for timed sleep
	rtcInit()

	for {
		// Sensor read
		t, err := mcpOneShotC()
		if err == nil {
			// Payload: [type=0x01, temp_q4.4 int16 big-endian]
			tempQ4 := int16(t * 16)
			pkt := []byte{0x01, byte(tempQ4 >> 8), byte(tempQ4)}
			ccSend(pkt)
		}
		// Hour sleep
		sleepSeconds(wakeSeconds)
	}
}

// RTC1 IRQ
//
//go:extern IRQ_RTC1
func irqRTC1()
func init() {
	arm.SetPriority(nrf.IRQ_RTC1, 3)
}

//export RTC1_IRQHandler
func RTC1_IRQHandler() {
	// Clear COMPARE0 event
	if nrf.RTC1.EVENTS_COMPARE[0].Get() != 0 {
		nrf.RTC1.EVENTS_COMPARE[0].Set(0)
		// Also clear the compare to avoid retrigger
		nrf.RTC1.CC[0].Set(0)
		rtcFired.Set(true)
	}
}
