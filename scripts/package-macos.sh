#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="$ROOT/dist/NoShot.app"
CONTENTS="$APP_DIR/Contents"
MACOS="$CONTENTS/MacOS"
RESOURCES="$CONTENTS/Resources"

rm -rf "$APP_DIR" "$ROOT/dist/NoShot-macos-arm64.zip"
mkdir -p "$MACOS" "$RESOURCES"

go build -o "$MACOS/NoShot" "$ROOT/cmd/noshot"
cp "$ROOT/assets/menubar/noshot-status.pdf" "$RESOURCES/noshot-status.pdf"

cat > "$CONTENTS/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleDevelopmentRegion</key>
  <string>en</string>
  <key>CFBundleExecutable</key>
  <string>NoShot</string>
  <key>CFBundleIdentifier</key>
  <string>com.colinknapp.noshot</string>
  <key>CFBundleInfoDictionaryVersion</key>
  <string>6.0</string>
  <key>CFBundleName</key>
  <string>NoShot</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleShortVersionString</key>
  <string>0.2.5</string>
  <key>CFBundleVersion</key>
  <string>8</string>
  <key>LSMinimumSystemVersion</key>
  <string>12.0</string>
  <key>LSUIElement</key>
  <true/>
  <key>NSHumanReadableCopyright</key>
  <string>CC-BY-4.0 Colin Knapp</string>
</dict>
</plist>
PLIST

codesign --force --deep --sign - "$APP_DIR" >/dev/null

(
  cd "$ROOT/dist"
  ditto -c -k --sequesterRsrc --keepParent "NoShot.app" "NoShot-macos-arm64.zip"
)

echo "$APP_DIR"
echo "$ROOT/dist/NoShot-macos-arm64.zip"
