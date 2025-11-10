#!/bin/bash

tinygo build -target=teensy36 -o firmware.hex main.go && teensy_loader_cli -mmcu=TEENSY36 -w firmware.hex