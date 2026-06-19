//go:build darwin && cgo

package main

/*
#cgo darwin LDFLAGS: -framework Cocoa -framework Carbon
void noshot_run(void);
*/
import "C"

import (
	"log"

	"github.com/Leopere/noshot/internal/app"
)

const (
	menuOpenScreenshots = 1
	menuEditConfig      = 2
)

var controller *app.Controller

func main() {
	cfg, configPath, err := app.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	controller = app.NewController(cfg, configPath)
	app.Notify("NoShot", "Running in the menu bar")
	C.noshot_run()
}

//export goHandleHotkey
func goHandleHotkey(id C.int) {
	controller.HandleHotkey(int(id))
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
	default:
		app.Notify("NoShot", "Unknown menu action")
	}
}
