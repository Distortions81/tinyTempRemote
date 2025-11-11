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
	if !area.valid() {
		return
	}

	x0 := clampInt16(area.x, 0, displayWidth-1)
	x1 := clampInt16(area.x+area.width-1, x0, displayWidth-1)
	y0 := clampInt16(area.y, 0, displayHeight-1)
	y1 := clampInt16(area.y+area.height-1, y0, displayHeight-1)

	buf := display.GetBuffer()
	width := int(displayWidth)
	pageStart := y0 / 8
	pageEnd := y1 / 8
	for page := pageStart; page <= pageEnd; page++ {
		mask := pageMaskForRange(page, y0, y1)
		if mask == 0 {
			continue
		}
		inv := ^mask
		rowOffset := page * int16(width)
		for x := x0; x <= x1; x++ {
			idx := int(rowOffset) + int(x)
			buf[idx] &= inv
		}
	}

	markDirtyRect(rect{
		x:      x0,
		y:      y0,
		width:  x1 - x0 + 1,
		height: y1 - y0 + 1,
	})
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

func pageMaskForRange(page, yStart, yEnd int16) byte {
	pageBase := page * 8
	mask := byte(0)
	for bit := int16(0); bit < 8; bit++ {
		y := pageBase + bit
		if y < yStart || y > yEnd {
			continue
		}
		mask |= 1 << uint(bit)
	}
	return mask
}
