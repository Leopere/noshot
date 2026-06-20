#!/usr/bin/env bash
set -euo pipefail

PREFS="$HOME/Library/Preferences/com.apple.symbolichotkeys.plist"
BACKUP="$PREFS.noshot-backup-$(date +%Y%m%d-%H%M%S)"

if [[ -f "$PREFS" ]]; then
  cp -p "$PREFS" "$BACKUP"
  echo "Backed up $PREFS to $BACKUP"
fi

disable_standard_hotkey() {
  local id="$1"
  local ascii="$2"
  local keycode="$3"
  local modifiers="$4"

  defaults write com.apple.symbolichotkeys AppleSymbolicHotKeys -dict-add "$id" \
    "{ enabled = 0; value = { parameters = ($ascii, $keycode, $modifiers); type = standard; }; }"
}

# Shift-Command-3: save picture of screen as a file.
disable_standard_hotkey 28 51 20 1179648

# Shift-Command-4: save picture of selected area as a file.
disable_standard_hotkey 30 52 21 1179648

# Shift-Command-5: screenshot and recording options.
disable_standard_hotkey 184 53 23 1179648

# If Apple's screenshot toolbar was left in recording mode, reset that state so it
# stops asking for audio permissions when the toolbar is triggered elsewhere.
defaults write com.apple.screencapture video -bool false

if [[ -x /System/Library/PrivateFrameworks/SystemAdministration.framework/Resources/activateSettings ]]; then
  /System/Library/PrivateFrameworks/SystemAdministration.framework/Resources/activateSettings -u || true
fi

killall SystemUIServer 2>/dev/null || true
killall cfprefsd 2>/dev/null || true

echo "Disabled Apple screenshot hotkeys for Shift-Command-3, Shift-Command-4, and Shift-Command-5."
