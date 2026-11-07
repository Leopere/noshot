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
- Retention: automatically remove screenshots older than 60 days
- Codex integration: uses the installed `codex` CLI and its current local auth
- Logs: `~/Library/Logs/NoShot/noshot.log`

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

The app targets macOS and captures through Apple's supported `screencapture` utility.

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
  "retention_days": 60,
  "codex_command": "codex",
  "codex_work_dir": ""
}
```

The filename template is a Go time layout. Set `retention_days` to `0` to keep
screenshots indefinitely. Cleanup recognizes NoShot's configured filename
template and the legacy Greenshot filename format, while leaving unrelated files
alone. Cleanup runs at launch and once a day while NoShot is running, keeping
directory scans out of the capture path.

The menu includes capture and Codex self-tests. Use those first when macOS
permissions or Codex auth are unclear.

By default, Codex runs from `~/Library/Application Support/NoShot/codex-work`
against a copied screenshot, so it does not need direct access to your
screenshots folder or this development repo.

## License

Licensed under CC-BY-4.0. Attribute original work to Colin Knapp and
https://colinknapp.com.

## Notarized public Mac releases

Commit and normal-push the existing `main` branch, then run
`scripts/release-macos`. This command submits the exact source revision to
Jenkins job `public-macos-noshot`. Jenkins runs the recorded checks and
build, signs a separate candidate with Colin Knapp’s Developer ID and secure
timestamps, retains the Apple notarization submission, staples the accepted
ticket, and checks Gatekeeper before publishing a verified public download.
The development repository keeps its existing visibility and license.

Apple review can continue between bounded Jenkins jobs. The existing candidate
and submission are reused; a pending review does not rebuild or reupload the
app. Observe the current phase with:

```
python3 ~/.local/share/jenkins-local/manage-macos-distribution.py observe --app noshot
```

Local package/build commands remain available for development. Public Mac
publication uses the notarized Jenkins path above. The process preserves the
installed app, local data, permissions, and working development signature.
Secure timestamps allow validation at signing time after certificate expiry;
revocation, restricted-capability profiles, and platform policy still apply.
