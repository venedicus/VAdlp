# Contributing

## Setup

Requirements are in [README.md](README.md): Go, yt-dlp and ffmpeg for manual tests, gcc when building the UI.

```bash
go build -o vadlp ./cmd/vadlp
./vadlp
```

Or `task run` if you use Task.

## Layout

| Path | |
|------|--|
| `internal/core` | Config, yt-dlp args, profiles, session |
| `internal/downloader` | Subprocess, logs, progress |
| `internal/updater` | Binary checks and downloads |
| `internal/i18n` | Strings |
| `internal/ui/fyne` | UI only |

Keep download logic out of the UI package when you can.

## Pull requests

- One change per PR when possible
- `go test ./...` and `go build -o vadlp ./cmd/vadlp` must pass
- Note what you tested manually
- Update README if user-visible behaviour changes

## Commits

Short imperative subject (`Fix queue cancel`). No drive-by refactors mixed with fixes.

## Style

- `gofmt` before push (`task fmt`)
- Don't swallow errors without a reason
- UI strings: add keys to `internal/i18n/locales/en.json` and `ru.json`
