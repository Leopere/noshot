//go:build !darwin || !cgo

package app

import (
	"context"
	"fmt"
)

func Notify(title, message string) {}

func AskCustomPrompt(ctx context.Context) (string, error) {
	return "", fmt.Errorf("custom prompt requires macOS with cgo")
}

func OpenPath(path string) error {
	return fmt.Errorf("open path requires macOS with cgo: %s", path)
}

func EditPath(path string) error {
	return fmt.Errorf("edit path requires macOS with cgo: %s", path)
}

func CopyTextToClipboard(text string) error {
	return fmt.Errorf("text clipboard support requires macOS with cgo")
}
