# VAdlp

[![CI](https://github.com/venedicus/VAdlp/actions/workflows/ci.yml/badge.svg)](https://github.com/venedicus/VAdlp/actions/workflows/ci.yml)

Desktop GUI for [yt-dlp](https://github.com/yt-dlp/yt-dlp). Go backend + **React** UI ([Wails](https://wails.io/)) in Bubble Tea style.

Build the command, run downloads, keep a queue, save profiles.

## Features

- Live preview of the yt-dlp command
- Download tab: URL, batch list, output path and filename template
- Format: custom `-f` string, merge container, quick presets (1080p, 4K, audio-only, …)
- Persistent profile bar (visible on every tab): save, save-as, rename, delete (URL is not stored in a profile)
- Queue: drag-and-drop reorder, pause/resume queued items, edit a queued item's config (e.g. rate limit) before it runs, retry failed, cancel running, parallel workers, scheduled (delayed) start
- Clipboard watcher: offers to queue a supported URL it sees copied while the app is focused
- Network: cookies (browser or file), proxy, rate limit, login
- Playlist limits, session save/load/resume (resume hints applied on load)
- Extras: subtitles, thumbnails, SponsorBlock, extra flags
- Format list from `yt-dlp -J` with thumbnails
- History tab with search and status filter
- Dependencies tab: yt-dlp/ffmpeg/deno status, install/update, prefers a VAdlp-managed copy over a system PATH copy, warns about outdated tools
- Light/dark/auto theme, adjustable UI scale, 11 languages (English, Russian, Spanish, Portuguese, Japanese, German, French, Polish, Korean, Traditional Chinese, Simplified Chinese)
- System tray: closing the window minimizes to tray and downloads keep running in the background; desktop notification when a download or queue run finishes
- VAdlp itself checks GitHub for newer releases on startup
- Export/import all settings and profiles as one backup file
- yt-dlp auto-install on first start if missing

## Requirements

| | |
|---|---|
| Go | 1.25+ |
| Node.js | 18+ (frontend build) |
| yt-dlp | runtime; offered for install if missing |
| ffmpeg | recommended for merge and `-x` |
| gcc | required to build Wails (CGO) |

## Download

Pre-built binaries: [Releases](https://github.com/venedicus/VAdlp/releases) (tags `v*`, e.g. `v0.1.2`).

Platforms: Linux (amd64, arm64), Windows (amd64), macOS (amd64, arm64; `.dmg` + tarball). Optional AppImage may appear when the CI step succeeds. See [RELEASE.md](RELEASE.md) for checksums and verification.

## Build and run

Wails needs CGO (`CGO_ENABLED=1`) and a C compiler. The app is built with **Wails** (not plain `go build`); output goes to `build/bin/` (see [Taskfile.yml](Taskfile.yml)).

**Recommended** — [Task](https://taskfile.dev):

```bash
git clone https://github.com/venedicus/VAdlp.git
cd VAdlp
task run      # Wails build into build/bin/ and run
task dev      # Wails dev mode with hot reload
task check    # fmt, vet, lint, tests, build
```

Without Task:

```bash
cd frontend && npm install && cd ..
go run github.com/wailsapp/wails/v2/cmd/wails@latest build -nopackage
build/bin/vadlp.exe   # Windows
# build/bin/vadlp     # Linux/macOS
```

For development with live reload:

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@latest dev
```

`main.go` embeds `frontend/dist` via `go:embed`, so a plain `go build .` only produces a working binary if `frontend/dist` was already built (`npm run build` inside `frontend/`) — otherwise it embeds a stale or empty UI. `wails build`/`task build:app` runs that frontend build step for you; use it instead of building `main.go` directly.

The first Wails build on a machine needs a C compiler, Node.js, and can take several minutes. Rebuilds are normally faster.

### Windows (gcc)

Install MinGW-w64 ([MSYS2](https://www.msys2.org/) is fine), put `gcc` on `PATH`, open a new shell, check `go env CGO_ENABLED` is `1`.

### Linux (Wails / WebView deps)

Debian/Ubuntu:

```bash
sudo apt-get install gcc libgtk-3-dev libwebkit2gtk-4.1-dev
```

## yt-dlp lookup

1. `%AppData%/vadlp/tools/` or `~/.config/vadlp/tools/` (permanent install location)
2. `<app>/bin/yt-dlp[.exe]`
3. `<app>/yt-dlp[.exe]`
4. `./bin/` from cwd
5. `PATH`
6. Install from Dependencies tab (GitHub latest release)

## Layout

```text
main.go                 Wails entry + embedded frontend
internal/app/           Wails backend API (React bindings), system tray
internal/core/          config, command builder, profiles, session, history
internal/downloader/    yt-dlp process, progress parsing
internal/updater/       yt-dlp, ffmpeg, deno, VAdlp's own release check
internal/health/        startup health checks surfaced in the UI
internal/settings/      settings.json
internal/i18n/          11 locales (en, ru, es, pt, ja, de, fr, pl, ko, zh-Hans, zh-Hant)
frontend/               React UI (Bubble Tea style)
internal/version/       build version (ldflags)
tools/                  CI/local build and packaging helpers
```

Profiles on disk: `%AppData%\vadlp\profiles\` (Windows) or `~/.config/vadlp/profiles/`.

## CI and quality

On every push/PR to `main`:

- `golangci-lint`, `gofmt`, `go vet`, tests, `govulncheck`
- Native builds on Linux (amd64 + arm64), Windows, macOS (arm64 + amd64)

Details: [CONTRIBUTING.md](CONTRIBUTING.md), workflows in [.github/workflows](.github/workflows).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Security: [SECURITY.md](SECURITY.md).

## License

MIT
