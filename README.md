# VAdlp — Video-Audio dlp

A simple and powerful desktop GUI for [**yt-dlp**](https://github.com/yt-dlp/yt-dlp), written in Go with the [Fyne](https://fyne.io/) toolkit.

Configure downloads, preview the generated command, and manage queues — all in one place.

## Features

- Form-driven configuration with a live command preview
- **Automatic yt-dlp installer** — if yt-dlp is not found at startup, VAdlp offers to download it automatically from the official GitHub release
- Sequential download queue
- Built-in presets (YouTube playlist, audio-only)
- Per-file and aggregate (playlist or queue) progress
- Session export/import for resuming work after a restart
- Adaptive log panel that fills all available vertical space
- Custom Tokyo Night–inspired theme

## Requirements

| Component | Notes |
|-----------|--------|
| **Go** | 1.22 or newer (for building from source) |
| **yt-dlp** | Required at runtime; installed automatically on first launch if missing |
| **ffmpeg** | Strongly recommended for merge, remux, and audio extraction |
| **C toolchain** | **gcc** on `PATH` when building with Fyne (CGO) |

## Locating yt-dlp

The application resolves the `yt-dlp` executable in this order:

1. `<app dir>/bin/yt-dlp[.exe]`
2. `<app dir>/yt-dlp[.exe]`
3. A sibling `../bin/` layout
4. `./bin/` under the current working directory
5. **`PATH`** via the standard library lookup
6. **Auto-install** — if none of the above succeed, a dialog appears at startup offering to download yt-dlp from `github.com/yt-dlp/yt-dlp/releases/latest` and place it next to the application

## Quick start

```bash
git clone https://github.com/youruser/vadlp
cd vadlp
go run ./cmd/vadlp
```

With [Task](https://taskfile.dev/):

```bash
task run
```

## Building

```bash
go build -o vadlp ./cmd/vadlp
```

```bash
task build      # go build ./...
task build:app  # go build -o vadlp ./cmd/vadlp
```

## Windows: CGO / gcc

Fyne uses CGO. On Windows, install a MinGW-w64 toolchain (for example via [MSYS2](https://www.msys2.org/)), add the compiler `bin` directory to your system `PATH`, and confirm `gcc --version` and `go env CGO_ENABLED` (expect `1`) in a **new** shell before building.

## Project layout

```text
cmd/vadlp/             Application entrypoint
internal/core/         Configuration, command construction, presets, session file
internal/downloader/   Subprocess runner, stdout parsing, progress events
internal/updater/      yt-dlp / ffmpeg availability check and auto-installer
internal/ui/fyne/      Fyne UI, theme, status widgets
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT License.
