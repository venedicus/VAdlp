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

archive_path="$ARCHIVE"
if [[ "$archive_path" != /* && ! "$archive_path" =~ ^[A-Za-z]:[/\\] ]]; then
  archive_path="$ROOT/$archive_path"
fi

cp "$BINARY" "$STAGING/"
[[ -f "$ROOT/README.md" ]] && cp "$ROOT/README.md" "$STAGING/"
[[ -f "$ROOT/LICENSE" ]] && cp "$ROOT/LICENSE" "$STAGING/"

if [[ "${RUNNER_OS:-}" == "Windows" ]]; then
  export PATH="/mingw64/bin:/usr/bin:${PATH}"
fi

case "$FORMAT" in
  tar)
    tar -czf "$archive_path" -C "$STAGING" .
    ;;
  zip)
    if command -v zip >/dev/null 2>&1; then
      (cd "$STAGING" && zip -qr "$archive_path" .)
    elif [[ "${RUNNER_OS:-}" == "Windows" ]]; then
      staging_win="$(cygpath -wa "$STAGING")"
      archive_win="$(cygpath -wa "$archive_path")"
      powershell.exe -NoProfile -Command "
        \$d = '$archive_win'
        if (Test-Path \$d) { Remove-Item -Force \$d }
        Compress-Archive -Path (Join-Path '$staging_win' '*') -DestinationPath \$d -CompressionLevel Optimal
      "
    else
      echo "zip not found; install zip or use Windows runner" >&2
      exit 1
    fi
    ;;
  dmg)
    hdiutil create -volname "VAdlp" -srcfolder "$STAGING" -ov -format UDZO "$archive_path"
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
    mv "$STAGING"/VAdlp*.AppImage "$archive_path"
    ;;
  *)
    echo "unknown format: $FORMAT" >&2
    exit 1
    ;;
esac

echo "created $archive_path"
