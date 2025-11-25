package main

func parseTelemetryLine(line string) (float64, bool) {
	const prefix = "TEMP,"
	if len(line) <= len(prefix) || line[:len(prefix)] != prefix {
		return 0, false
	}

	value := line[len(prefix):]
	return parseTemperatureValue(value)
}

func parseTemperatureValue(text string) (float64, bool) {
	if len(text) == 0 {
		return 0, false
	}

	idx := 0
	negative := false
	if text[idx] == '-' {
		negative = true
		idx++
	} else if text[idx] == '+' {
		idx++
	}
	if idx >= len(text) {
		return 0, false
	}

	var whole int64
	digits := 0
	for idx < len(text) && text[idx] != '.' {
		ch := text[idx]
		if ch < '0' || ch > '9' {
			return 0, false
		}
		whole = whole*10 + int64(ch-'0')
		idx++
		digits++
	}
	if digits == 0 {
		return 0, false
	}

	var frac int64
	var fracDiv float64 = 1
	if idx < len(text) {
		if text[idx] != '.' {
			return 0, false
		}
		idx++
		if idx >= len(text) {
			return 0, false
		}
		for idx < len(text) {
			ch := text[idx]
			if ch < '0' || ch > '9' {
				return 0, false
			}
			frac = frac*10 + int64(ch-'0')
			fracDiv *= 10
			idx++
		}
	}

	value := float64(whole)
	if fracDiv > 1 {
		value += float64(frac) / fracDiv
	}
	if negative {
		value = -value
	}
	return value, true
}
