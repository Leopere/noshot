//go:build !darwin || !cgo

package app

import "fmt"

func nativeCapture(path string, mode CaptureMode) error {
	return fmt.Errorf("capture requires macOS: %s", path)
}
