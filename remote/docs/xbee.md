# XBee Pro 900 RPSMA Integration

This firmware now exposes a simple UART-backed transport so the Teensy 3.6 can stream temperature readings over a Digi XBee Pro 900 RPSMA in transparent (AT) mode. The defaults target UART1 (Teensy RX2/TX2 on D9/D10, which is the MK66F18’s UART1) so the OLED and MCP9808 sensor can stay on the soft I²C bus.

## Wiring
 - `D10` → XBee DIN (pin 3) – Teensy TX2 (UART2) pin, 3.3 V logic only.
 - `D09` → XBee DOUT (pin 2) – Teensy RX2 (UART2) pin, 3.3 V logic only.
- `3V3` → XBee VCC (pin 1). The radio can draw ~215 mA on transmit, so use the Teensy VIN header only if your supply can source it.
- `GND` → XBee GND (pin 10).
- Optional: assign `xbeeResetPin` to any spare GPIO to drive the module’s `RESET` pin (pin 5). Set `xbeeSleepPin` if you want to toggle `SLEEP_RQ` (pin 9) for low-power cycles.

Pins and baud rate live in `config.go`:
```go
var (
	xbeeTxPin machine.Pin = machine.D10
	xbeeRxPin machine.Pin = machine.D09
	xbeeResetPin machine.Pin = machine.NoPin
	xbeeSleepPin machine.Pin = machine.NoPin
	xbeeUART = machine.TeensyUART2 // UART1 on the MK66F18
)
const xbeeBaudRate = 9600
const xbeeBlinkLEDOnTx = true
const xbeeBlinkDurationMs int64 = 15
```
Move the module to another UART or change the baud rate by editing those values and rebuilding. Set `xbeeBlinkLEDOnTx` to `false` if you don’t want the Teensy’s LED to blink briefly on every transmission; adjust `xbeeBlinkDurationMs` for longer or shorter pulses.

Disable the OLED when you only need a headless transmitter by toggling `enableOLED`.
```go
const (
	enableOLED = true
	testTxModeEnabled = false
	testTxIntervalMs = 1000
	testTxStartTempC = 20.0
	testTxMaxTempC = 30.0
	testTxStepTempC = 0.5
)
```
Flip `enableOLED` to `false` to keep the MCP9808/XBee running but skip the SSD1306 for lower power. Turn on `testTxModeEnabled` if you want the radio to emit a rolling Celsius value (between `testTxStartTempC` and `testTxMaxTempC`) once per second for link testing when the sensor is absent; it still blinks the LED each packet like the normal path.

## Data format
Each successful temperature sample is emitted as:

```
TEMP,<Fahrenheit>,<Celsius>\r\n
```

Example: `TEMP,72.4 F,22.5 C`. You can also call `SendTextLine("debug...")` from within the firmware to push ad-hoc diagnostics through the same link.

## Commissioning checklist
1. Program both radios with matching PAN ID, channel, and destination addresses using Digi’s XCTU or AT commands (transparent mode is assumed).
2. Power the XBee from a stable 3.3 V source before the Teensy boots so the automatic reset pulse sees a valid VCC.
3. Build/flash the firmware (`./build.sh`). On boot the Teensy pulses `RESET` (if wired) and keeps `SLEEP_RQ` low so the radio is immediately awake.
4. On the receiving side, connect another XBee to a PC/MCU at the same baud rate, or plug into USB via an FTDI cable. You should see ASCII `TEMP,...` lines every sensor update.

If you don’t receive traffic, verify RSSI with the remote XBee, confirm UART wiring, and optionally enable the `xbeeResetPin` to guarantee the module is brought out of reset during startup.
