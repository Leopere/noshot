//go:build !darwin || !cgo

package app

import "fmt"

func nativeCapture(path string, mode CaptureMode) error {
	return fmt.Errorf("native capture requires macOS with cgo")
}
