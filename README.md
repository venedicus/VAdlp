# yt-dlp Desktop

Cross-platform desktop client for [**yt-dlp**](https://github.com/yt-dlp/yt-dlp), written in Go with the [Fyne](https://fyne.io/) toolkit. It focuses on a clear workflow: configure a download, inspect the generated command, run it, and monitor progress and logs in one place.

## Features

- Form-driven configuration with a live command preview
- Sequential download queue
- Built-in presets (for example YouTube playlist and audio-only)
- Per-file and aggregate (playlist or queue) progress
- Session export/import for resuming work after a restart
- Custom Tokyo Night–inspired theme

## Requirements

| Component | Notes |
|-----------|--------|
| **Go** | 1.22 or newer (for building from source) |
| **yt-dlp** | Required at runtime; see [Locating yt-dlp](#locating-yt-dlp) |
| **ffmpeg** | Strongly recommended for merge, remux, and audio extraction |
| **C toolchain** | **gcc** on `PATH` when building with Fyne (CGO) |

## Locating yt-dlp

The application resolves the `yt-dlp` executable in this order:

1. `<directory of this app>/bin/yt-dlp` (or `yt-dlp.exe` on Windows)
2. `<directory of this app>/yt-dlp` (same file next to the GUI binary)
3. A sibling `../bin/` layout (some install or `go install` arrangements)
4. `./bin/` under the **current working directory** (typical when developing with `go run`)
5. **`PATH`** via the standard library lookup

GUI launches (for example double-clicking an `.exe` on Windows) often use a different working directory and a shorter `PATH` than an interactive shell. Bundling `yt-dlp` next to the application or under `bin/` beside it avoids “works in the terminal only” behaviour.

## Quick start

```bash
git clone <repository-url>
cd yt-dlp-desktop-go-gui
go run ./cmd/ytgui
```

With [Task](https://taskfile.dev/):

```bash
task run
```

## Building

```bash
go build -o ytgui ./cmd/ytgui
```

On Windows the output is commonly named `ytgui.exe`. Cross-compilation follows normal Go `GOOS` / `GOARCH` conventions; ensure a suitable C toolchain when the target platform requires CGO.

```bash
task build      # go build ./...
task build:app  # go build ./cmd/ytgui
```

## Windows: CGO / gcc

Fyne uses CGO. On Windows, install a MinGW-w64 toolchain (for example via [MSYS2](https://www.msys2.org/)), add the compiler `bin` directory to your system `PATH`, and confirm `gcc --version` and `go env CGO_ENABLED` (expect `1`) in a **new** shell before building.

## Project layout

```text
cmd/ytgui/           # Application entrypoint
internal/core/      # Configuration, command construction, presets, session file format
internal/downloader/# Subprocess runner, stdout parsing, progress events
internal/ui/fyne/   # Fyne UI, theme, status widgets
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This project is released under the [MIT License](LICENSE).
