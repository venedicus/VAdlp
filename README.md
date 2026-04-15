# yt-dlp-desktop-go-gui (Fyne, Go)

Cross-platform desktop GUI for `yt-dlp` with:
- command builder and live preview
- queue (sequential downloads)
- presets (`YouTube Playlist`, `Audio Only`)
- progress and logs
- TokyoNight-inspired theme

## Project structure

```text
yt-dlp-desktop-go-gui/
  cmd/ytgui/main.go             # app entrypoint
  internal/core/                # config, command builder, presets
  internal/downloader/          # yt-dlp runner + progress parsing
  internal/ui/fyne/             # Fyne UI and TokyoNight theme
  bin/                          # optional local yt-dlp binary
  go.mod
```

## Requirements

### Runtime tools
- `yt-dlp` (either in `PATH` or in `bin/yt-dlp(.exe)`)
- `ffmpeg` (recommended for merging/remux/audio workflows)

### Build tools (Windows + Fyne)
- Go 1.22+
- C compiler (`gcc`) available in `PATH` (required by Fyne/cgo)

## Windows setup (recommended)

1. Install [MSYS2](https://www.msys2.org/).
2. Open **MSYS2 UCRT64** terminal and run:
   - `pacman -Syu`
   - reopen terminal
   - `pacman -S --needed mingw-w64-ucrt-x86_64-gcc`
3. Add this folder to your system `PATH`:
   - `C:\msys64\ucrt64\bin`
4. Open a new PowerShell and verify:
   - `gcc --version`
   - `go env CGO_ENABLED` (should be `1`)

## Run

From project root:

```powershell
go run ./cmd/ytgui
```

Or with [Task](https://taskfile.dev/):

```powershell
task run
```

## Build

```powershell
go build ./cmd/ytgui
```

Output binary appears in current directory (for example `ytgui.exe` on Windows).

With Task:

```powershell
task build
task build:app
```

### Manual binary builds for testing

Build current OS binary into `build/bin`:

```powershell
mkdir build\bin -Force
go build -o .\build\bin\ytgui.exe .\cmd\ytgui
```

Cross-build examples (Windows host):

```powershell
# Linux
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o .\build\bin\ytgui-linux-amd64 .\cmd\ytgui

# macOS (Intel)
$env:GOOS="darwin"; $env:GOARCH="amd64"; go build -o .\build\bin\ytgui-darwin-amd64 .\cmd\ytgui

# macOS (Apple Silicon)
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -o .\build\bin\ytgui-darwin-arm64 .\cmd\ytgui
```

After cross-build commands, reset env for local builds:

```powershell
Remove-Item Env:GOOS -ErrorAction Ignore
Remove-Item Env:GOARCH -ErrorAction Ignore
```

## Local yt-dlp binary

If you prefer project-local binary, place it here:

- Windows: `bin/yt-dlp.exe`
- Linux/macOS: `bin/yt-dlp`

The app uses local binary first, then fallback to system `yt-dlp` from `PATH`.

## Troubleshooting

### `cgo: C compiler "gcc" not found`
- `gcc` is not visible in current shell `PATH`.
- Fix: add MinGW/MSYS2 gcc folder to `PATH`, then restart terminal/IDE.
- Verify with: `where gcc` and `gcc --version`.

### `yt-dlp` not found
- Put `yt-dlp` in `PATH` or copy binary into `bin/`.

### Slow/failed post-processing
- Install `ffmpeg` and ensure it is available in `PATH`.

## Notes

- Current queue mode is sequential (one task at a time).
- Config persistence and custom presets can be added next.
