//go:build darwin

package app

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func nativeCapture(path string, mode CaptureMode) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	args := []string{"-x", "-t", "png"}
	switch mode {
	case CaptureRegion:
		args = append(args, "-i", "-s")
	case CaptureWindow:
		args = append(args, "-i", "-w")
	case CaptureFullscreen:
	default:
		return fmt.Errorf("unknown capture mode %d", mode)
	}
	args = append(args, path)

	cmd := exec.CommandContext(ctx, "/usr/sbin/screencapture", args...)
	output, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		message := strings.TrimSpace(string(output))
		Logf("screencapture failed mode=%d path=%q err=%v output=%q", mode, path, err, message)
		if strings.Contains(strings.ToLower(message), "could not create image from display") {
			return fmt.Errorf("macOS denied screen capture; grant Screen & System Audio Recording permission to NoShot")
		}
		if message == "" {
			message = err.Error()
		}
		return fmt.Errorf("screencapture failed: %s", message)
	}
	return nil
}
