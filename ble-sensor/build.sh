#!/bin/bash

# Parse arguments
ACTION="$1"
DEBUG=""

if [ "$ACTION" = "debug" ]; then
    DEBUG="-tags=debug"
    echo "Building with DEBUG enabled (serial logging active)"
fi

echo "Building and flashing firmware to nice!nano... [plug in device now]"

# Run tinygo flash and capture exit status
if tinygo flash -target=nicenano $DEBUG .; then
    echo "Flash succeeded."

    # Open serial only if DEBUG mode was activated
    if [ "$ACTION" = "debug" ]; then
        # Use minicom instead of screen
        PORT="/dev/ttyACM0"
        BAUD=115200

        echo "Opening serial port with minicom..."
        echo "(Exit with Ctrl+A, then X)"

		sleep 3

        # Automatically configure minicom for this session
        minicom -D "$PORT" -b "$BAUD"
    fi
else
    echo "Flash failed — NOT opening serial port."
    exit 1
fi
