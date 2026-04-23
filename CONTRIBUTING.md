# Contributing

Thank you for taking the time to improve this project. The guidelines below keep reviews predictable and the codebase maintainable.

## Getting started

1. Install the [requirements](README.md#requirements) from the main README (Go, `yt-dlp`, `ffmpeg` for manual tests, and a C compiler when building the Fyne UI).
2. Clone the repository and verify the application runs:

   ```bash
   go run ./cmd/ytgui
   ```

## Architecture

| Path | Responsibility |
|------|----------------|
| `internal/core` | Configuration model, `yt-dlp` argument construction, presets, session persistence |
| `internal/downloader` | Process execution, log streaming, heuristics for progress and stage detection |
| `internal/ui/fyne` | Presentation only; avoid embedding download protocol details here |

Prefer small, testable helpers in `internal/core` and `internal/downloader` over large UI callbacks.

## Pull requests

1. **Scope** — One logical change per pull request when practical.
2. **Build** — `go build ./...` must succeed.
3. **Manual check** — Describe what you exercised in the PR (for example: single URL, playlist, queue, session load).
4. **Documentation** — Update `README.md` when behaviour or requirements visible to users change.

## Commits

Use short, imperative subject lines (for example: `Fix playlist progress label`). Avoid bundling unrelated refactors with functional fixes unless necessary.

## Style

- Run `gofmt` (or `task fmt`) before submitting.
- Prefer explicit error handling over silent failure.
- Keep user-facing copy in English unless the project adds a dedicated localization layer.
- Source comments and identifiers remain ASCII unless an existing API or upstream convention requires otherwise.
