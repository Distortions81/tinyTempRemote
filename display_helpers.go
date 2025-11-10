package main

import (
	"machine"
	"math/rand"
)

func blinkOnce(pin machine.Pin, durationMs int64) {
	go func(d int64) {
		pin.High()
		sleepMs(d)
		pin.Low()
	}(durationMs)
}

type textOffset struct {
	x int16
	y int16
}

func randomOffset(r *rand.Rand, text string) textOffset {
	width := int16(len(text) * charWidth)
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
	width := int16(len(text) * charWidth)
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
}
