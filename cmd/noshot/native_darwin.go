//go:build darwin && cgo

package main

/*
#cgo darwin LDFLAGS: -framework Cocoa -framework Carbon
void noshot_run(void);
*/
import "C"

import (
	"fmt"
	"log"
	"time"

	"github.com/Leopere/noshot/internal/app"
)

const (
	menuOpenScreenshots = 1
	menuEditConfig      = 2
	menuCaptureSelfTest = 3
	menuCodexSelfTest   = 4
)

var controller *app.Controller

func main() {
	cfg, configPath, err := app.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	reapScreenshots(cfg, time.Now(), "startup")
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for now := range ticker.C {
			reapScreenshots(cfg, now, "daily")
		}
	}()
	controller = app.NewController(cfg, configPath)
	app.Notify("NoShot", "Running in the menu bar")
	C.noshot_run()
}

func reapScreenshots(cfg app.Config, now time.Time, reason string) {
	if count, err := app.ReapOldScreenshots(cfg, now); err != nil {
		app.Logf("%s screenshot cleanup failed: %v", reason, err)
	} else if count > 0 {
		app.Logf("reaped %d old screenshots during %s cleanup", count, reason)
	}
}

//export goHandleHotkey
func goHandleHotkey(id C.int) {
	controller.HandleHotkey(int(id))
}

//export goHandleHotkeyRegistration
func goHandleHotkeyRegistration(id C.int, status C.int) {
	shortcut := fmt.Sprintf("Cmd+Shift+%d", int(id))
	if status == 0 {
		app.Logf("hotkey registered id=%d shortcut=%q", int(id), shortcut)
		return
	}

	statusName := hotkeyRegistrationStatusName(int(status))
	app.Logf("hotkey registration failed id=%d shortcut=%q status=%d statusName=%q", int(id), shortcut, int(status), statusName)
	app.Notify("NoShot", fmt.Sprintf("%s unavailable: %s", shortcut, statusName))
}

func hotkeyRegistrationStatusName(status int) string {
	switch status {
	case -9878:
		return "eventHotKeyExistsErr"
	case -9879:
		return "eventHotKeyInvalidErr"
	default:
		return fmt.Sprintf("OSStatus %d", status)
	}
}

//export goHandleMenu
func goHandleMenu(action C.int) {
	switch int(action) {
	case menuOpenScreenshots:
		if err := controller.OpenScreenshotsFolder(); err != nil {
			app.Notify("NoShot", err.Error())
		}
	case menuEditConfig:
		if err := controller.EditConfig(); err != nil {
			app.Notify("NoShot", err.Error())
		}
	case menuCaptureSelfTest:
		controller.RunCaptureSelfTest()
	case menuCodexSelfTest:
		controller.RunCodexSelfTest()
	default:
		app.Notify("NoShot", "Unknown menu action")
	}
}
