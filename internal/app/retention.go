package app

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const legacyGreenshotFilenameLayout = "Greenshot 2006-01-02 15.04.05.png"

func ReapOldScreenshots(cfg Config, now time.Time) (int, error) {
	if cfg.RetentionDays <= 0 {
		return 0, nil
	}

	dir, err := ExpandPath(cfg.ScreenshotsDir)
	if err != nil {
		return 0, err
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	cutoff := now.Add(-time.Duration(cfg.RetentionDays) * 24 * time.Hour)
	reaped := 0
	for _, entry := range entries {
		info, err := entry.Info()
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return reaped, err
		}
		if !info.Mode().IsRegular() || !info.ModTime().Before(cutoff) {
			continue
		}

		name := entry.Name()
		imageName := strings.TrimSuffix(name, ".codex.md")
		if !isManagedScreenshotName(imageName, cfg.FilenameTemplate) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, name)); err != nil {
			return reaped, err
		}
		if name == imageName {
			reaped++
			if err := os.Remove(filepath.Join(dir, name+".codex.md")); err != nil && !errors.Is(err, os.ErrNotExist) {
				return reaped, err
			}
		}
	}
	return reaped, nil
}

func isManagedScreenshotName(name, layout string) bool {
	if name == "noshot_codex_selftest.png" {
		return true
	}
	return matchesScreenshotLayout(name, layout) || matchesScreenshotLayout(name, legacyGreenshotFilenameLayout)
}

func matchesScreenshotLayout(name, layout string) bool {
	if parsesScreenshotLayout(name, layout) {
		return true
	}

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	separator := strings.LastIndexByte(base, '_')
	if separator < 0 {
		return false
	}
	collision, err := strconv.Atoi(base[separator+1:])
	if err != nil || collision < 1 || collision >= 1000 {
		return false
	}
	return parsesScreenshotLayout(base[:separator]+ext, layout)
}

func parsesScreenshotLayout(name, layout string) bool {
	if _, err := time.ParseInLocation(layout, name, time.Local); err == nil {
		return true
	}
	if strings.EqualFold(filepath.Ext(name), ".png") {
		_, err := time.ParseInLocation(layout, strings.TrimSuffix(name, filepath.Ext(name)), time.Local)
		return err == nil
	}
	return false
}
