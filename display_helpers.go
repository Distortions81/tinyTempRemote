package main

import (
	"machine"
	"time"
)

type randomIntn interface {
	Intn(n int) int
}

func blinkOnce(pin machine.Pin, durationMs int64) {
	pin.High()
	time.Sleep(time.Duration(durationMs) * time.Millisecond)
	pin.Low()
}

type textOffset struct {
	x int16
	y int16
}

func randomOffset(r randomIntn, text string) textOffset {
	width := textPixelWidth(text)
	if width > displayWidth {
		width = displayWidth
	}
	maxX := int16(displayWidth) - width
	if maxX < 0 {
		maxX = 0
	}
	x := int16(r.Intn(int(maxX) + 1))

	minY := int16(charHeight)
	maxY := int16(displayHeight - 1)
	if maxY <= minY {
		maxY = minY + 1
	}
	yRange := int(maxY - minY)
	y := minY
	if yRange > 0 {
		y += int16(r.Intn(yRange))
	}

	return textOffset{x: x, y: y}
}

func clampOffsetX(offset textOffset, text string) textOffset {
	width := textPixelWidth(text)
	if width > displayWidth {
		width = displayWidth
	}
	maxX := int16(displayWidth) - width
	if maxX < 0 {
		maxX = 0
	}
	if offset.x > maxX {
		offset.x = maxX
	}
	return offset
}

func blinkError(pin machine.Pin) {
	for i := 0; i < 3; i++ {
		blinkOnce(pin, 75)
		sleepMs(75)
	}
}

func resetDisplay(pin machine.Pin) {
	if pin == machine.NoPin {
		return
	}

	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	sleepMs(100)
	pin.High()
	sleepMs(100)
	pin.Low()
	sleepMs(100)
	pin.High()
	blinkOnce(machine.LED, 2)
}
