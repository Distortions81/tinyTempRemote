// +build !debug

package main

// No-op debug functions when debug is not enabled
func debugPrint(s string)     {}
func debugPrintln(s string)   {}
func debugPrintInt(val int)   {}
func debugPrintHex(val uint8) {}
func debugPrintBool(val bool) {}
