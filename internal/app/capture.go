package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type CaptureMode int

const (
	CaptureRegion CaptureMode = iota + 1
	CaptureWindow
	CaptureFullscreen
)

func Capture(ctx context.Context, cfg Config, mode CaptureMode) (string, error) {
	dir, err := ExpandPath(cfg.ScreenshotsDir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	path, err := nextScreenshotPath(dir, cfg.FilenameTemplate, time.Now())
	if err != nil {
		return "", err
	}

	args := []string{"-x", "-t", "png"}
	switch mode {
	case CaptureRegion:
		args = append(args, "-i", "-s")
	case CaptureWindow:
		args = append(args, "-i", "-w")
	case CaptureFullscreen:
	default:
		return "", fmt.Errorf("unknown capture mode %d", mode)
	}
	args = append(args, path)

	cmd := exec.CommandContext(ctx, "screencapture", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("screencapture failed: %w: %s", err, string(output))
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() == 0 {
		_ = os.Remove(path)
		return "", fmt.Errorf("screencapture produced an empty file")
	}

	if cfg.CopyImageToClipboard {
		if err := CopyImageToClipboard(path); err != nil {
			return path, fmt.Errorf("saved screenshot but could not copy image: %w", err)
		}
	}

	return path, nil
}

func nextScreenshotPath(dir, layout string, now time.Time) (string, error) {
	name := now.Format(layout)
	if filepath.Ext(name) == "" {
		name += ".png"
	}
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path, nil
	} else if err != nil {
		return "", err
	}

	ext := filepath.Ext(name)
	base := name[:len(name)-len(ext)]
	for i := 1; i < 1000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		} else if err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("could not find an available filename in %s", dir)
}
