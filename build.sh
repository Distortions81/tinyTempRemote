#!/bin/bash

set -euo pipefail

tinygo build -target=teensy36 -o firmware.hex . &&
teensy_loader_cli -mmcu=TEENSY36 -w firmware.hex
