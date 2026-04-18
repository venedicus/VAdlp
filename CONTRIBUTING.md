# Contributing

Thanks for contributing to `yt-dlp-desktop-go-gui`.

## Development setup

1. Install dependencies from `README.md`:
   - Go 1.22+
   - `gcc` in `PATH` (for Fyne/cgo)
   - `yt-dlp` and `ffmpeg` for runtime testing
2. Clone repository and run:
   - `go run ./cmd/ytgui`

## Project layout

- `cmd/ytgui`: app entrypoint
- `internal/core`: config, presets, command builder
- `internal/downloader`: yt-dlp process runner and progress events
- `internal/ui/fyne`: UI and theming

## Coding guidelines

- Keep business logic in `internal/core` and `internal/downloader`.
- Keep UI-specific code in `internal/ui/fyne`.
- Prefer small pure functions for command-building logic.
- Keep changes ASCII unless existing file requires Unicode.

## Pull request checklist

- [ ] Code builds locally with `go build ./...`
- [ ] App runs with `go run ./cmd/ytgui`
- [ ] New behavior is manually verified
- [ ] README/docs updated for user-visible changes
- [ ] No generated binaries committed

## Commit style

- Use concise, imperative commit messages.
- Prefer one logical change per commit.
