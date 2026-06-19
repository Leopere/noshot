package app

import (
	"context"
	"fmt"
	"os"
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

	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := nativeCapture(path, mode); err != nil {
		_ = os.Remove(path)
		return "", err
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() == 0 {
		_ = os.Remove(path)
		return "", fmt.Errorf("native capture produced an empty file")
	}

	if cfg.CopyImageToClipboard {
		if err := CopyImageToClipboard(path); err != nil {
			Logf("clipboard image copy failed path=%q: %v", path, err)
			return path, fmt.Errorf("saved screenshot but could not copy image: %w", err)
		}
		Logf("clipboard image copy succeeded path=%q", path)
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
