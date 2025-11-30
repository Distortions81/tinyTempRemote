#!/bin/bash

# Favor minimal firmware size per https://tinygo.org/docs/guides/optimizing-binaries/
if [ "$1" = "flash" ]; then
	# Build the firmware first
	echo "Building firmware..."
	tinygo build \
		-target=nicenano \
		-opt=z \
		-panic=trap \
		-scheduler=none \
		-gc=leaking \
		-o firmware.uf2 . || exit 1

	echo "Triggering bootloader mode..."

	# Trigger bootloader (this will fail to flash but will reset the device)
	tinygo flash -target=nicenano /dev/null 2>/dev/null || true

	# Wait for device to appear
	sleep 2

	# Find the USB device (should be sda, sdb, etc. with usb transport)
	DEVICE=""
	for dev in /sys/block/sd*; do
		if [ -f "$dev/device/vendor" ]; then
			VENDOR=$(cat "$dev/device/vendor" 2>/dev/null | tr -d ' ')
			if [[ "$VENDOR" == *"Adafruit"* ]] || [[ "$VENDOR" == *"nRF"* ]]; then
				DEVICE="/dev/$(basename $dev)"
				break
			fi
		fi
	done

	# If vendor check didn't work, just find first USB mass storage device
	if [ -z "$DEVICE" ]; then
		DEVICE=$(lsblk -d -o NAME,TRAN | grep usb | head -1 | awk '{print "/dev/"$1}')
	fi

	if [ -z "$DEVICE" ]; then
		echo "ERROR: Could not find nice!nano USB device"
		echo "Make sure the board is connected via USB"
		exit 1
	fi

	echo "Found device: $DEVICE"

	# Mount the device
	MOUNT_POINT="/tmp/nicenano_$$"
	mkdir -p "$MOUNT_POINT"

	echo "Mounting bootloader..."
	sudo mount "$DEVICE" "$MOUNT_POINT" 2>/dev/null || sudo mount "${DEVICE}1" "$MOUNT_POINT" || {
		echo "ERROR: Failed to mount device"
		rmdir "$MOUNT_POINT"
		exit 1
	}

	# Copy firmware
	echo "Copying firmware..."
	sudo cp firmware.uf2 "$MOUNT_POINT/"
	sync

	# Unmount
	echo "Flashing..."
	sudo umount "$MOUNT_POINT"
	rmdir "$MOUNT_POINT"

	echo ""
	echo "✓ Flash successful! Board should be running new firmware."
else
	# Just build the firmware
	tinygo build \
		-target=nicenano \
		-opt=z \
		-panic=trap \
		-scheduler=none \
		-gc=leaking \
		-size short \
		-o firmware.uf2 . &&
	ls -lh firmware.uf2 &&
	echo "" &&
	echo "To flash: Double-tap reset button, then copy firmware.uf2 to NICENANO drive"
fi
