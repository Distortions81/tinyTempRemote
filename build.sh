#!/bin/bash

tinygo build \
	-target=teensy36 \
	-opt=z \
	-panic=trap \
	-o firmware.hex . &&
teensy_loader_cli -mmcu=TEENSY36 -w firmware.hex
