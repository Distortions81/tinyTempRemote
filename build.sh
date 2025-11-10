#!/bin/bash

# Favor minimal firmware size per https://tinygo.org/docs/guides/optimizing-binaries/
tinygo build \
	-target=teensy36 \
	-opt=z \
	-panic=trap \
	-scheduler=none \
	-gc=leaking \
	-size short \
	-o firmware.hex . &&
SIZE=$(python3 - <<'PY'
max_end=0
with open('firmware.hex') as f:
    for line in f:
        if not line.startswith(':'):
            continue
        count = int(line[1:3], 16)
        addr = int(line[3:7], 16)
        rectype = line[7:9]
        if rectype != '00':
            continue
        end = addr + count
        if end > max_end:
            max_end = end
print(max_end)
PY
) && echo "Firmware size: ${SIZE} bytes" &&
teensy_loader_cli -mmcu=TEENSY36 -w firmware.hex
