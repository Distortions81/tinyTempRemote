# BLE Display Client

OLED display client firmware for the nice!nano (nRF52840) board that receives temperature data via Bluetooth Low Energy from a remote sensor.

## Hardware

- **Board**: nice!nano (nRF52840-based)
- **Display**: SSD1306 OLED (128x32, I2C address 0x3C)
- **Wireless**: Bluetooth Low Energy (BLE) for receiving telemetry

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

- **Display settings**: Contrast, dimensions, update interval
- **Bluetooth settings**: Remote sensor name (`remoteSensorName`), LED blink on receive
- **Low power mode**: Configure `idleLowPowerMode`

## Features

- **BLE Client**: Scans for and receives temperature data from remote sensor
- **OLED Display**: Shows received temperature with animated text positioning to prevent burn-in
- **Low Power Mode**: Optimized for battery operation using ARM WFI sleep modes
- **LED Feedback**: Blinks LED when temperature data is received

## Firmware Size

Current build size:
- **Flash**: ~14.3 KB
- **RAM**: ~5.7 KB

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

## Usage

1. Flash this firmware to a nice!nano with an OLED display
2. Flash the [remote-sensor](../remote-sensor/) firmware to another nice!nano with a temperature sensor
3. Power on both devices
4. The display client will automatically scan for and connect to the "TempSensor" device
5. Temperature readings will appear on the OLED display

## Note

The BLE scanning and pairing functionality is currently a placeholder. Full BLE GATT client implementation is coming soon, which will include:
- Active scanning for devices with name "TempSensor"
- Automatic connection to remote sensor
- Subscription to temperature characteristic notifications
- Display of received temperature data

## License

See the main repository for license information.
