package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReapOldScreenshots(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 7, 11, 12, 0, 0, 0, time.Local)
	old := now.Add(-61 * 24 * time.Hour)
	recent := now.Add(-59 * 24 * time.Hour)

	writeAgedFile(t, dir, "greenshot_2026-05-01_10-00-00.png", old)
	writeAgedFile(t, dir, "greenshot_2026-05-01_10-00-00.png.codex.md", recent)
	writeAgedFile(t, dir, "greenshot_2026-05-01_10-00-00_1.png", old)
	writeAgedFile(t, dir, "Greenshot 2026-05-01 10.00.00.png", old)
	writeAgedFile(t, dir, "greenshot_2026-07-01_10-00-00.png", recent)
	writeAgedFile(t, dir, "keep-me.png", old)

	cfg := DefaultConfig()
	cfg.ScreenshotsDir = dir
	reaped, err := ReapOldScreenshots(cfg, now)
	if err != nil {
		t.Fatal(err)
	}
	if reaped != 3 {
		t.Fatalf("reaped = %d, want 3", reaped)
	}

	for _, name := range []string{
		"greenshot_2026-05-01_10-00-00.png",
		"greenshot_2026-05-01_10-00-00.png.codex.md",
		"greenshot_2026-05-01_10-00-00_1.png",
		"Greenshot 2026-05-01 10.00.00.png",
	} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("expected %q to be removed, stat error = %v", name, err)
		}
	}
	for _, name := range []string{"greenshot_2026-07-01_10-00-00.png", "keep-me.png"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("expected %q to remain: %v", name, err)
		}
	}
}

func TestReapOldScreenshotsCanBeDisabled(t *testing.T) {
	dir := t.TempDir()
	name := "greenshot_2025-01-01_10-00-00.png"
	writeAgedFile(t, dir, name, time.Now().Add(-365*24*time.Hour))

	cfg := DefaultConfig()
	cfg.ScreenshotsDir = dir
	cfg.RetentionDays = 0
	reaped, err := ReapOldScreenshots(cfg, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if reaped != 0 {
		t.Fatalf("reaped = %d, want 0", reaped)
	}
	if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
		t.Fatalf("expected screenshot to remain: %v", err)
	}
}

func writeAgedFile(t *testing.T, dir, name string, modified time.Time) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatal(err)
	}
}
