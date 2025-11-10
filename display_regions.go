package main

import "tinygo.org/x/drivers/ssd1306"

type rect struct {
	x      int16
	y      int16
	width  int16
	height int16
}

func (r rect) valid() bool {
	return r.width > 0 && r.height > 0
}

func textBoundsAt(pos textOffset, text string) rect {
	width := textPixelWidth(text)
	if width <= 0 {
		return rect{}
	}

	x := clampInt16(pos.x, 0, displayWidth-1)
	if x >= int16(displayWidth) {
		return rect{}
	}

	if x+width > int16(displayWidth) {
		width = int16(displayWidth) - x
	}
	if width <= 0 {
		return rect{}
	}

	bottom := clampInt16(pos.y, 0, displayHeight-1)
	top := bottom - int16(charHeight) + 1
	if top < 0 {
		top = 0
	}
	height := bottom - top + 1

	return rect{
		x:      x,
		y:      top,
		width:  width,
		height: height,
	}
}

func clearRect(display *ssd1306.Device, area rect) {
	fillRect(display, area, false)
}

func fillRect(display *ssd1306.Device, area rect, on bool) {
	if !area.valid() {
		return
	}
	xEnd := area.x + area.width
	yEnd := area.y + area.height
	if xEnd > int16(displayWidth) {
		xEnd = int16(displayWidth)
	}
	if yEnd > int16(displayHeight) {
		yEnd = int16(displayHeight)
	}
	for y := area.y; y < yEnd; y++ {
		for x := area.x; x < xEnd; x++ {
			setPixel(display, x, y, on)
		}
	}
}

func clampInt16(v, min, max int16) int16 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
