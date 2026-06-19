//go:build !darwin || !cgo

package app

import "fmt"

func CopyImageToClipboard(path string) error {
	return fmt.Errorf("image clipboard support requires macOS with cgo: %s", path)
}
