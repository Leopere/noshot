package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const ExplainPrompt = "Explain this screenshot clearly and concisely."

type CodexResult struct {
	AnswerPath string
	Output     string
}

func RunCodexOnImage(ctx context.Context, cfg Config, imagePath, prompt string) (CodexResult, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		prompt = ExplainPrompt
	}

	runCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	var stderr bytes.Buffer
	cmd := exec.CommandContext(
		runCtx,
		codexCommand(cfg),
		"exec",
		"--ephemeral",
		"--skip-git-repo-check",
		"--cd",
		codexWorkDir(cfg),
		"--image",
		imagePath,
		"-",
	)
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(prompt)

	output, err := cmd.Output()
	Logf("codex exec image=%q prompt=%q stderr=%q", imagePath, prompt, stderr.String())
	if runCtx.Err() != nil {
		return CodexResult{}, runCtx.Err()
	}
	if err != nil {
		return CodexResult{}, fmt.Errorf("codex failed: %w: %s", err, stderr.String())
	}

	answer := strings.TrimSpace(string(output))
	answerPath := imagePath + ".codex.md"
	if err := os.WriteFile(answerPath, []byte(answer+"\n"), 0o644); err != nil {
		return CodexResult{}, err
	}
	if err := CopyTextToClipboard(answer); err != nil {
		return CodexResult{}, fmt.Errorf("saved Codex answer but could not copy text: %w", err)
	}

	return CodexResult{AnswerPath: answerPath, Output: answer}, nil
}

func CodexSelfTest(ctx context.Context, cfg Config) (CodexResult, string, error) {
	dir, err := ExpandPath(cfg.ScreenshotsDir)
	if err != nil {
		return CodexResult{}, "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return CodexResult{}, "", err
	}
	imagePath := filepath.Join(dir, "noshot_codex_selftest.png")
	const png = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAIAAAD8GO2jAAAAK0lEQVR4nO3NMQ0AAAgEsVeJUQSCBkaSXm5vUn16jgUAAAAAAAAAAAB8ARZ8dAx5CUbFQQAAAABJRU5ErkJggg=="
	data, err := base64.StdEncoding.DecodeString(png)
	if err != nil {
		return CodexResult{}, "", err
	}
	if err := os.WriteFile(imagePath, data, 0o644); err != nil {
		return CodexResult{}, "", err
	}
	result, err := RunCodexOnImage(ctx, cfg, imagePath, "Reply with one short sentence describing this test image.")
	return result, imagePath, err
}

func codexWorkDir(cfg Config) string {
	if strings.TrimSpace(cfg.CodexWorkDir) != "" {
		if dir, err := ExpandPath(cfg.CodexWorkDir); err == nil {
			return dir
		}
		return cfg.CodexWorkDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}

func codexCommand(cfg Config) string {
	command := strings.TrimSpace(cfg.CodexCommand)
	if command == "" || command == defaultCodexCommand {
		for _, candidate := range []string{
			"/Users/aedev/.local/bin/codex",
			"/opt/homebrew/bin/codex",
			"/usr/local/bin/codex",
		} {
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate
			}
		}
		if path, err := exec.LookPath(defaultCodexCommand); err == nil {
			return path
		}
		return defaultCodexCommand
	}
	return command
}
