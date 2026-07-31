# Releases

## Versioning

Tags use [Semantic Versioning](https://semver.org/): `vMAJOR.MINOR.PATCH` (example: `v0.1.0`).

## Cut a release

1. Ensure `main` is green in [CI](https://github.com/venedicus/VAdlp/actions/workflows/ci.yml).
2. Update changelog notes if you keep them manually (optional; GitHub can generate release notes).
3. Tag and push:

```bash
git tag v0.9.0
git push origin v0.9.0
```

The [Release workflow](.github/workflows/release.yml) builds assets and publishes a GitHub Release.

## Assets

| File | Platform | Notes |
|------|----------|--------|
| `vadlp-linux-amd64.tar.gz` | Linux x86_64 | portable tarball |
| `vadlp-linux-amd64.AppImage` | Linux x86_64 | no install; needs FUSE on host |
| `vadlp-linux-arm64.tar.gz` | Linux ARM64 | portable tarball |
| `vadlp-windows-amd64.zip` | Windows x86_64 | portable zip |
| `vadlp-windows-arm64.zip` | Windows ARM64 | not in CI yet (build from source) |
| `vadlp-darwin-arm64.tar.gz` | macOS Apple Silicon | portable tarball |
| `vadlp-darwin-arm64.dmg` | macOS Apple Silicon | `VAdlp.app` disk image, ad-hoc signed (proper `.app` bundle from v0.9.2+; was a bare unsigned binary before) |
| `vadlp-darwin-amd64.tar.gz` | macOS Intel | portable tarball |
| `vadlp-darwin-amd64.dmg` | macOS Intel | `VAdlp.app` disk image, ad-hoc signed (proper `.app` bundle from v0.9.2+; was a bare unsigned binary before) |

Each primary archive has a `.sha256` sidecar. `checksums.txt` on the release lists all payloads.

## Verify download

```bash
sha256sum -c vadlp-linux-amd64.tar.gz.sha256
```

## Code signing (optional, not enabled by default)

To sign release binaries in CI, add repository secrets and extend the workflow:

| Secret | Use |
|--------|-----|
| `APPLE_CERTIFICATE_BASE64` + `APPLE_CERTIFICATE_PASSWORD` | macOS `.dmg` / binary signing |
| `WINDOWS_CERTIFICATE_BASE64` + `WINDOWS_CERTIFICATE_PASSWORD` | Authenticode for `.exe` |

Without secrets, assets are unsigned (typical for open-source nightlies).

## Not in automated releases (yet)

- Windows **MSI** / macOS notarization pipeline
- Store packages (Microsoft Store, Homebrew cask)

## Release history

- **`v0.10.0`** — fixes a download hang and adds an optional Browse tab.

  **Fixes.** Downloads could freeze indefinitely (reported roughly every
  20–30 videos): stdout and stderr were read serially, so a full stderr pipe
  deadlocked the process against the app, and cancelling could not get
  through. Both streams are now drained concurrently and the subprocess is
  killed and reaped on cancel. Cancelling one queue task no longer cancels
  every running task. Playlist progress no longer miscounts on filenames
  like `Part 1 of 3.mp4`. yt-dlp, ffmpeg and deno downloads are now verified
  against the publisher's SHA-256 manifest before use, and the install
  aborts when integrity cannot be confirmed. Probe subprocesses have
  timeouts. A slow or blocked `api.github.com` no longer freezes startup or
  leaves a permanent error banner.

  **New.** Optional Browse tab (off by default, enable in Settings): search
  YouTube, open channels and playlists, preview in an embedded player, and
  download at the resolution you pick — using your existing cookies and
  proxy settings. See [docs/feature-youtube-browser.md](docs/feature-youtube-browser.md),
  which also plans the account-based viewing that is not built yet.

  **Internal.** `App.tsx` split into per-tab components, the hand-written
  DTO copy replaced by the generated Go models, and ~400 lines of new tests
  covering queue orchestration, cancellation and the pipe deadlock.
- **`v0.9.2`** — fixes the macOS `.dmg`: it now contains a real `VAdlp.app` bundle (ad-hoc signed) with an `/Applications` symlink for drag-install, instead of a bare unsigned executable plus README/LICENSE. Also fixes a packaging bug that stripped the executable bit from binaries staged into `tar`/`zip` archives. No functional app changes.
- **`v0.9.1`** — fixes the macOS build, which failed to link (`getlantern/systray`'s Darwin `AppDelegate` class collided with Wails' own); switched to `energye/systray`. Docs cleanup only otherwise.
- **`v0.9.0`** — full rewrite of the UI from Fyne to Wails v2 + React (Bubble Tea–styled). Adds: system tray with background downloads, drag-and-drop queue reordering, clipboard URL detection, scheduled queue start, per-task queue editing, history search/filter, light/dark/auto theme, VAdlp self-update check, settings export/import, multi-instance detection and management, and 11 UI languages (English, Russian, Spanish, Portuguese, Japanese, German, French, Polish, Korean, Traditional Chinese, Simplified Chinese). The old Fyne UI is removed; see [ARCHITECTURE.md](ARCHITECTURE.md) for the current layout. Pre-1.0, so expect the occasional rough edge — file an issue if you hit one.
- `v0.1.1` was the first successful public release on the old Fyne UI. `v0.1.0` failed in CI before assets were published.
