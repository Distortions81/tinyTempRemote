package main

func parseTelemetryLine(line string) (string, string, bool) {
	const prefix = "TEMP,"
	if len(line) <= len(prefix) || line[:len(prefix)] != prefix {
		return "", "", false
	}

	firstComma := -1
	for i := len(prefix); i < len(line); i++ {
		if line[i] == ',' {
			firstComma = i
			break
		}
	}
	if firstComma == -1 {
		value := line[len(prefix):]
		return value, "", len(value) > 0
	}

	fahrenheit := line[len(prefix):firstComma]
	if len(fahrenheit) == 0 {
		return "", "", false
	}

	secondComma := -1
	for i := firstComma + 1; i < len(line); i++ {
		if line[i] == ',' {
			secondComma = i
			break
		}
	}

	var celsius string
	if secondComma == -1 {
		if firstComma+1 < len(line) {
			celsius = line[firstComma+1:]
		}
	} else {
		celsius = line[firstComma+1 : secondComma]
	}

	return fahrenheit, celsius, true
}
