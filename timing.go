package main

import _ "unsafe"

const microsecondsPerMillisecond int64 = 1000

//go:linkname runtimeSleepTicks runtime.sleepTicks
func runtimeSleepTicks(duration int64)

//go:linkname runtimeTicks runtime.ticks
func runtimeTicks() int64

func sleepMicros(us int64) {
	if us <= 0 {
		return
	}
	deadline := runtimeTicks() + us
	for {
		now := runtimeTicks()
		remaining := deadline - now
		if remaining <= 0 {
			return
		}
		runtimeSleepTicks(remaining)
	}
}

func sleepMs(ms int64) {
	sleepMicros(ms * microsecondsPerMillisecond)
}

func millis() int64 {
	return runtimeTicks() / microsecondsPerMillisecond
}
