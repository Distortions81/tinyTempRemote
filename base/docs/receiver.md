# Base Station Receiver

This Teensy 3.6 firmware listens to the same Digi XBee Pro 900 RPSMA telemetry stream produced by the remote node and mirrors the reported Fahrenheit temperature on an SSD1306 128×32 OLED. It reuses the same soft-I²C wiring and font assets so both ends look identical on the bench.

## Wiring
- `D18` → OLED SDA, `D19` → OLED SCL (soft I²C, 400 kHz).
- `D03` → OLED reset line (optional but recommended).
- `D10` → XBee DIN (pin 3), `D09` → XBee DOUT (pin 2).
- `3V3` and `GND` → OLED and XBee power rails. Budget at least 215 mA headroom for the radio.

All logic is 3.3 V, so no level shifting is needed when pairing with an XBee.

## Firmware notes
- Build/flash with `cd base && ./build.sh` (requires TinyGo + `teensy_loader_cli`).
- `config.go` exposes `telemetryIdleDelayMs`, `telemetryStaleTimeoutMs`, and the UART pinout if you need to tune responsiveness or move to a different serial port.
- On boot and before the first packet arrives the OLED shows `00`, then each valid `TEMP,<Fahrenheit>,<Celsius>` line updates the display and briefly blinks the onboard LED (disable via `xbeeBlinkLEDOnRx`).
- If no fresh telemetry arrives within 20 s the display falls back to `FF` at a fixed position so you can tell the link is stale.

Because the OLED font only contains digits, `.` and `F`, the receiver shows the Fahrenheit string exactly as transmitted. Change the remote formatter if you need additional units or text on the base display.
