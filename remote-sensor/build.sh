#!/bin/bash
# Parse arguments
ACTION="$1"
DEBUG=""

if [ "$1" = "debug" ]; then
	DEBUG="-tags=debug"
	echo "Building with DEBUG enabled (serial logging active)"
fi

echo "Building and flashing firmware to nice!nano..."

tinygo flash -target=nicenano $DEBUG .