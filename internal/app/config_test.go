package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigApplyDefaults(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ApplyDefaults()

	if cfg.ScreenshotsDir != defaultScreenshotsDir {
		t.Fatalf("ScreenshotsDir = %q, want %q", cfg.ScreenshotsDir, defaultScreenshotsDir)
	}
	if cfg.FilenameTemplate != defaultFilenameLayout {
		t.Fatalf("FilenameTemplate = %q, want %q", cfg.FilenameTemplate, defaultFilenameLayout)
	}
	if !cfg.CopyImageToClipboard {
		t.Fatal("CopyImageToClipboard should default to true")
	}
	if cfg.CodexCommand != defaultCodexCommand {
		t.Fatalf("CodexCommand = %q, want %q", cfg.CodexCommand, defaultCodexCommand)
	}
}

func TestSaveConfigCreatesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.json")

	if err := SaveConfig(path, Config{}); err != nil {
		t.Fatalf("SaveConfig returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
}

func TestExpandPathHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	got, err := ExpandPath("~/Pictures/Greenshot")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "Pictures/Greenshot")
	if got != want {
		t.Fatalf("ExpandPath = %q, want %q", got, want)
	}
}
