# Remote Sensor

Temperature sensor firmware for the nice!nano (nRF52840) board with optional OLED display and Bluetooth Low Energy wireless telemetry.

## Hardware

- **Board**: nice!nano (nRF52840-based)
- **Temperature Sensor**: MCP9808 (I2C address 0x18)
- **Display** (optional): SSD1306 OLED (128x32, I2C address 0x3C)
- **Wireless**: Bluetooth Low Energy (BLE) for telemetry

## Pin Configuration

### I2C (Hardware I2C0)
- **SDA**: P0_17
- **SCL**: P0_20

### Other Pins
- **LED**: P0_15 (built-in LED)
- **Display Reset**: P0_02
- **Random Seed**: P0_31 (analog pin for entropy)

## Building

### Build firmware only
```bash
./build.sh
```

This creates `firmware.uf2` file.

### Build and flash to board
```bash
./build.sh flash
```

This automatically triggers bootloader mode and flashes the firmware.

## Configuration

Edit [config.go](config.go) to customize:

- **Display settings**: Enable/disable OLED, contrast, dimensions
- **Sensor polling**: Adjust `sensorPollDelayMs` (default: 5000ms)
- **Bluetooth settings**: Device name (`bleDeviceName`), enable/disable BLE, LED blink on transmit
- **Low power mode**: Configure `idleLowPowerMode`
- **Test mode**: Enable `testTxModeEnabled` for testing without sensor

## Features

- **Temperature Sensing**: Reads temperature from MCP9808 sensor in Celsius, displays in Fahrenheit
- **OLED Display**: Shows temperature with animated text positioning to prevent burn-in
- **Bluetooth Low Energy**: Advertises as "TempSensor" for wireless connectivity (pairing support coming soon)
- **Low Power Mode**: Optimized for battery operation using ARM WFI sleep modes
- **Error Handling**: LED blink patterns indicate sensor errors

## Firmware Size

Current build size:
- **Flash**: ~13.3 KB
- **RAM**: ~4.7 KB

## Dependencies

- [TinyGo](https://tinygo.org/) compiler
- [tinygo.org/x/drivers/ssd1306](https://pkg.go.dev/tinygo.org/x/drivers/ssd1306) - OLED display driver
- [tinygo.org/x/bluetooth](https://pkg.go.dev/tinygo.org/x/bluetooth) - Bluetooth Low Energy stack

## Flashing

### Method 1: Using build script (Recommended)
```bash
./build.sh flash
```

### Method 2: Manual UF2 copy
1. Double-tap the reset button on the nice!nano
2. Board appears as USB drive named "NICENANO"
3. Copy `firmware.uf2` to the drive
4. Board automatically resets and runs new firmware

## License

See the main repository for license information.
