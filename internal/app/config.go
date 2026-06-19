package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultScreenshotsDir = "~/Pictures/Greenshot"
	defaultFilenameLayout = "greenshot_2006-01-02_15-04-05.png"
	defaultCodexCommand   = "codex"
	defaultCodexWorkDir   = ""
	applicationSupportDir = "Library/Application Support/NoShot"
	defaultConfigFilename = "config.json"
)

type Config struct {
	ScreenshotsDir       string `json:"screenshots_dir"`
	FilenameTemplate     string `json:"filename_template"`
	CopyImageToClipboard bool   `json:"copy_image_to_clipboard"`
	CodexCommand         string `json:"codex_command"`
	CodexWorkDir         string `json:"codex_work_dir"`
}

func DefaultConfig() Config {
	return Config{
		ScreenshotsDir:       defaultScreenshotsDir,
		FilenameTemplate:     defaultFilenameLayout,
		CopyImageToClipboard: true,
		CodexCommand:         defaultCodexCommand,
		CodexWorkDir:         defaultCodexWorkDir,
	}
}

func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, applicationSupportDir, defaultConfigFilename), nil
}

func LoadConfig() (Config, string, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, "", err
	}

	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, path, SaveConfig(path, cfg)
		}
		return Config{}, path, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, path, err
	}
	cfg.ApplyDefaults()
	return cfg, path, nil
}

func SaveConfig(path string, cfg Config) error {
	cfg.ApplyDefaults()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

func (c *Config) ApplyDefaults() {
	if strings.TrimSpace(c.ScreenshotsDir) == "" {
		c.ScreenshotsDir = defaultScreenshotsDir
	}
	if strings.TrimSpace(c.FilenameTemplate) == "" {
		c.FilenameTemplate = defaultFilenameLayout
	}
	if strings.TrimSpace(c.CodexCommand) == "" {
		c.CodexCommand = defaultCodexCommand
	}
	if strings.TrimSpace(c.CodexWorkDir) == "" {
		c.CodexWorkDir = defaultCodexWorkDir
	}
}

func ExpandPath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}
