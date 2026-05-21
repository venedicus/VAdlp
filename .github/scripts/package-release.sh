#!/usr/bin/env bash
set -euo pipefail

# Usage: package-release.sh <format> <binary> <archive>
# format: tar | zip | dmg | appimage

FORMAT="${1:?format required (tar|zip|dmg|appimage)}"
BINARY="${2:?binary path required}"
ARCHIVE="${3:?archive output path required}"

if [[ ! -f "$BINARY" ]]; then
  echo "binary not found: $BINARY" >&2
  exit 1
fi

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
STAGING="$(mktemp -d)"
cleanup() { rm -rf "$STAGING"; }
trap cleanup EXIT

cp "$BINARY" "$STAGING/"
[[ -f "$ROOT/README.md" ]] && cp "$ROOT/README.md" "$STAGING/"
[[ -f "$ROOT/LICENSE" ]] && cp "$ROOT/LICENSE" "$STAGING/"

case "$FORMAT" in
  tar)
    tar -czf "$ARCHIVE" -C "$STAGING" .
    ;;
  zip)
    (cd "$STAGING" && zip -r "$ROOT/$ARCHIVE" .)
    ;;
  dmg)
    hdiutil create -volname "VAdlp" -srcfolder "$STAGING" -ov -format UDZO "$ARCHIVE"
    ;;
  appimage)
    APPDIR="$STAGING/VAdlp.AppDir"
    mkdir -p "$APPDIR/usr/bin" "$APPDIR/usr/share/applications"
    cp "$BINARY" "$APPDIR/usr/bin/vadlp"
    cat >"$APPDIR/usr/share/applications/vadlp.desktop" <<'EOF'
[Desktop Entry]
Name=VAdlp
Comment=Desktop GUI for yt-dlp
Exec=vadlp
Icon=vadlp
Type=Application
Categories=AudioVideo;Network;
EOF
    cat >"$APPDIR/AppRun" <<'EOF'
#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
exec "$HERE/usr/bin/vadlp" "$@"
EOF
    chmod +x "$APPDIR/AppRun" "$APPDIR/usr/bin/vadlp"

    LINUXDEPLOY="$STAGING/linuxdeploy-x86_64.AppImage"
    curl -fsSL -o "$LINUXDEPLOY" \
      https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
    chmod +x "$LINUXDEPLOY"

    GTK_PLUGIN="$STAGING/linuxdeploy-plugin-gtk-x86_64.AppImage"
    curl -fsSL -o "$GTK_PLUGIN" \
      https://github.com/linuxdeploy/linuxdeploy-plugin-gtk/releases/download/continuous/linuxdeploy-plugin-gtk-x86_64.AppImage
    chmod +x "$GTK_PLUGIN"

    export ARCH=x86_64 APPIMAGE_EXTRACT_AND_RUN=1
    "$LINUXDEPLOY" --appdir "$APPDIR" --plugin gtk --output appimage
    mv "$STAGING"/VAdlp*.AppImage "$ARCHIVE"
    ;;
  *)
    echo "unknown format: $FORMAT" >&2
    exit 1
    ;;
esac

echo "created $ARCHIVE"
