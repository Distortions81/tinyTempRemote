#!/bin/bash

# Favor minimal firmware size per https://tinygo.org/docs/guides/optimizing-binaries/
if [ "$1" = "flash" ]; then
	# Flash directly to the board (triggers bootloader mode)
	tinygo flash \
		-target=nicenano \
		-opt=z \
		-panic=trap \
		-scheduler=none \
		-gc=leaking \
		.
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
	ls -lh firmware.uf2
fi
