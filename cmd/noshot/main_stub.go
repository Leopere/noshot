//go:build !darwin || !cgo

package main

import "log"

func main() {
	log.Fatal("NoShot currently requires macOS with cgo enabled")
}
