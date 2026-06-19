package app

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
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
		cfg.CodexCommand,
		"exec",
		"--ephemeral",
		"--skip-git-repo-check",
		"--image",
		imagePath,
		prompt,
	)
	cmd.Stderr = &stderr

	output, err := cmd.Output()
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
