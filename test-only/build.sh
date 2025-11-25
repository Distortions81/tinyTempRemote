#!/bin/bash
set -euo pipefail

# Build the goTempTest CLI that waits for a device to emit data before exiting.
go build -o goTempTest ./...
