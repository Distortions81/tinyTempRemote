# Minimal Bluetooth Test for Nice!Nano

This is a minimal test project based on the official TinyGo Bluetooth examples to verify basic BLE functionality on the Nice!Nano (nRF52840) board.

## Purpose

Test if basic Bluetooth works in isolation, without any other project dependencies that might be causing issues.

## What It Does

- Enables the BLE stack
- Advertises as "Go Bluetooth"
- Prints MAC address every second
- Handles device connections/disconnections
- Stops advertising when a device disconnects

## Hardware

- Nice!Nano (nRF52840-based board)
- USB cable for flashing and serial monitoring

## Building and Flashing

### Quick Start

```bash
# Build and flash to nice!nano
./build.sh

# Build, flash, and open serial monitor (debug mode)
./build.sh debug
```

The script uses the `nicenano` target and will automatically open minicom for serial monitoring if you use debug mode.

## Testing

1. Flash the firmware to your Nice!Nano
2. Open the serial monitor to see debug output
3. Use a BLE scanner app on your phone (e.g., nRF Connect) to look for "Go Bluetooth"
4. Try connecting to the device
5. Observe connection/disconnection messages in the serial monitor

## Expected Output

```
advertising...
Go Bluetooth / XX:XX:XX:XX:XX:XX
Go Bluetooth / XX:XX:XX:XX:XX:XX
device connected: YY:YY:YY:YY:YY:YY
device disconnected: YY:YY:YY:YY:YY:YY
```

## Troubleshooting

If the build fails:
- Ensure TinyGo is installed: `tinygo version`
- Verify Go version: `go version` (should be 1.19+)
- Download dependencies: `go mod download`

If flashing fails:
- Verify USB connection
- Try pressing reset button twice quickly to enter bootloader mode
- Check that the board appears as a USB device

If you can't see BLE advertisements:
- Run with debug mode: `./build.sh debug`
- Check serial output for error messages
- Verify the BLE stack enabled successfully
- Try using a different BLE scanner app

## Based On

Official TinyGo Bluetooth example:
https://github.com/tinygo-org/bluetooth/tree/release/examples/advertisement
