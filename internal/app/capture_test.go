package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNextScreenshotPathUsesTemplate(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 6, 19, 7, 30, 45, 0, time.Local)

	got, err := nextScreenshotPath(dir, defaultFilenameLayout, now)
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(dir, "greenshot_2026-06-19_07-30-45.png")
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

func TestNextScreenshotPathAvoidsCollisions(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 6, 19, 7, 30, 45, 0, time.Local)
	first := filepath.Join(dir, "greenshot_2026-06-19_07-30-45.png")
	if err := os.WriteFile(first, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := nextScreenshotPath(dir, defaultFilenameLayout, now)
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(dir, "greenshot_2026-06-19_07-30-45_1.png")
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}
