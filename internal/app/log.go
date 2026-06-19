package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func Logf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	log.Print(message)

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir := filepath.Join(home, "Library/Logs/NoShot")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	file, err := os.OpenFile(filepath.Join(dir, "noshot.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()
	_, _ = fmt.Fprintf(file, "%s %s\n", time.Now().Format(time.RFC3339), message)
}
