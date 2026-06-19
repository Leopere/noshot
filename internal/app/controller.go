package app

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type Controller struct {
	cfg        Config
	configPath string
	mu         sync.Mutex
	busy       bool
}

func NewController(cfg Config, configPath string) *Controller {
	return &Controller{cfg: cfg, configPath: configPath}
}

func (c *Controller) HandleHotkey(id int) {
	if !c.tryStart() {
		Notify("NoShot", "Capture already in progress")
		return
	}
	go func() {
		defer c.done()
		if err := c.handleHotkey(context.Background(), id); err != nil {
			log.Printf("hotkey %d failed: %v", id, err)
			Notify("NoShot", err.Error())
		}
	}()
}

func (c *Controller) handleHotkey(ctx context.Context, id int) error {
	switch id {
	case 1:
		return c.captureOnly(ctx, CaptureRegion)
	case 2:
		return c.captureOnly(ctx, CaptureWindow)
	case 3:
		return c.captureOnly(ctx, CaptureFullscreen)
	case 4:
		return c.captureAndAskCodex(ctx, ExplainPrompt)
	case 5:
		return c.captureAndAskCustomCodex(ctx)
	default:
		return fmt.Errorf("unknown hotkey %d", id)
	}
}

func (c *Controller) captureOnly(ctx context.Context, mode CaptureMode) error {
	path, err := Capture(ctx, c.cfg, mode)
	if err != nil {
		return err
	}
	Notify("NoShot", "Saved and copied "+path)
	return nil
}

func (c *Controller) captureAndAskCodex(ctx context.Context, prompt string) error {
	path, err := Capture(ctx, c.cfg, CaptureRegion)
	if err != nil {
		return err
	}
	Notify("NoShot", "Screenshot saved. Asking Codex...")

	result, err := RunCodexOnImage(ctx, c.cfg, path, prompt)
	if err != nil {
		return err
	}
	Notify("NoShot", "Codex answer copied and saved to "+result.AnswerPath)
	return nil
}

func (c *Controller) captureAndAskCustomCodex(ctx context.Context) error {
	path, err := Capture(ctx, c.cfg, CaptureRegion)
	if err != nil {
		return err
	}
	prompt, err := AskCustomPrompt(ctx)
	if err != nil {
		return fmt.Errorf("custom prompt cancelled; screenshot saved to %s", path)
	}
	Notify("NoShot", "Screenshot saved. Asking Codex...")

	result, err := RunCodexOnImage(ctx, c.cfg, path, prompt)
	if err != nil {
		return err
	}
	Notify("NoShot", "Codex answer copied and saved to "+result.AnswerPath)
	return nil
}

func (c *Controller) OpenScreenshotsFolder() error {
	dir, err := ExpandPath(c.cfg.ScreenshotsDir)
	if err != nil {
		return err
	}
	return OpenPath(dir)
}

func (c *Controller) EditConfig() error {
	return EditPath(c.configPath)
}

func (c *Controller) tryStart() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.busy {
		return false
	}
	c.busy = true
	return true
}

func (c *Controller) done() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.busy = false
}
