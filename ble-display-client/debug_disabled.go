// +build !debug

package main

// No-op debug functions when debug is not enabled
func debugPrint(s string)   {}
func debugPrintln(s string) {}
