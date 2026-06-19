# NoShot

NoShot is a dirt-simple, clean-room macOS screenshot helper inspired by the
workflow of Greenshot. It saves screenshots to disk, copies the image to the
clipboard by default, and can hand a screenshot to local Codex for explanation
or a custom prompt.

NoShot does not use Greenshot code, assets, names, or trademarks. Greenshot is
a separate project.

## Defaults

- Screenshots folder: `~/Pictures/Greenshot`
- Filename template: `greenshot_2006-01-02_15-04-05.png`
- Clipboard: copy the captured image after saving
- Codex integration: uses the installed `codex` CLI and its current local auth

## Hotkeys

- `Command+Shift+1`: capture a selected region
- `Command+Shift+2`: capture a window
- `Command+Shift+3`: capture the full screen
- `Command+Shift+4`: capture a selected region and ask Codex to explain it
- `Command+Shift+5`: capture a selected region and ask Codex with a custom prompt

macOS may already bind some of these shortcuts. Disable conflicting shortcuts in
System Settings before running NoShot.

## Build

```sh
make verify
make run
```

The app targets macOS and uses Apple's built-in `screencapture` command.

## Config

NoShot creates its config at:

```text
~/Library/Application Support/NoShot/config.json
```

Example:

```json
{
  "screenshots_dir": "~/Pictures/Greenshot",
  "filename_template": "greenshot_2006-01-02_15-04-05.png",
  "copy_image_to_clipboard": true,
  "codex_command": "codex"
}
```

The filename template is a Go time layout.

## License

Licensed under CC-BY-4.0. Attribute original work to Colin Knapp and
https://colinknapp.com.

