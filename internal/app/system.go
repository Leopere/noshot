package app

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func Notify(title, message string) {
	_ = exec.Command(
		"osascript",
		"-e",
		fmt.Sprintf("display notification %s with title %s", appleScriptQuote(message), appleScriptQuote(title)),
	).Run()
}

func AskCustomPrompt(ctx context.Context) (string, error) {
	script := `text returned of (display dialog "Ask Codex about this screenshot:" default answer "" buttons {"Cancel", "Run"} default button "Run" cancel button "Cancel")`
	out, err := exec.CommandContext(ctx, "osascript", "-e", script).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func OpenPath(path string) error {
	return exec.Command("open", path).Start()
}

func EditPath(path string) error {
	return exec.Command("open", "-t", path).Start()
}

func CopyTextToClipboard(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func appleScriptQuote(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}
