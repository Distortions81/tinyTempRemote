package main

import (
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ssd1306"
)

type dirtyPage struct {
	active bool
	minX   int16
	maxX   int16
}

var (
	oledDirtyPages [displayHeight / 8]dirtyPage
	oledXferBuf    [displayWidth + 1]byte
)

func markDirtyPixel(x, y int16) {
	markDirtyRect(rect{x: x, y: y, width: 1, height: 1})
}

func markDirtyRect(area rect) {
	if !area.valid() {
		return
	}
	x0 := clampInt16(area.x, 0, displayWidth-1)
	x1 := clampInt16(area.x+area.width-1, x0, displayWidth-1)
	y0 := clampInt16(area.y, 0, displayHeight-1)
	y1 := clampInt16(area.y+area.height-1, y0, displayHeight-1)

	pageStart := y0 / 8
	pageEnd := y1 / 8
	for page := pageStart; page <= pageEnd; page++ {
		dp := &oledDirtyPages[page]
		if !dp.active {
			dp.active = true
			dp.minX = x0
			dp.maxX = x1
			continue
		}
		if x0 < dp.minX {
			dp.minX = x0
		}
		if x1 > dp.maxX {
			dp.maxX = x1
		}
	}
}

func flushDirtyPages(display *ssd1306.Device, bus drivers.I2C) {
	buffer := display.GetBuffer()
	pageWidth := int(displayWidth)
	for page := range oledDirtyPages {
		dp := &oledDirtyPages[page]
		if !dp.active {
			continue
		}
		start := clampInt16(dp.minX, 0, displayWidth-1)
		end := clampInt16(dp.maxX, start, displayWidth-1)

		display.Command(ssd1306.COLUMNADDR)
		display.Command(uint8(start))
		display.Command(uint8(end))
		display.Command(ssd1306.PAGEADDR)
		display.Command(uint8(page))
		display.Command(uint8(page))

		pageOffset := page*pageWidth + int(start)
		length := int(end-start) + 1
		writeOLEDData(bus, buffer[pageOffset:pageOffset+length])

		dp.active = false
	}
}

func writeOLEDData(bus drivers.I2C, data []byte) {
	if len(data) == 0 {
		return
	}
	oledXferBuf[0] = 0x40
	copy(oledXferBuf[1:], data)
	// Ignore transmission errors for now; retries handled at higher level if needed.
	_ = bus.Tx(displayI2CAddr, oledXferBuf[:len(data)+1], nil)
}
