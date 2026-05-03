# Contributing

Thank you for taking the time to improve VAdlp. The guidelines below keep reviews predictable and the codebase maintainable.

## Getting started

1. Install the [requirements](README.md#requirements) from the main README (Go, `yt-dlp`, `ffmpeg` for manual tests, and a C compiler when building the Fyne UI).
2. Clone the repository and verify the application runs:

   ```bash
   go run ./cmd/vadlp
   ```

## Architecture

| Path | Responsibility |
|------|----------------|
| `internal/core` | Configuration model, `yt-dlp` argument construction, presets, session persistence |
| `internal/downloader` | Process execution, log streaming, heuristics for progress and stage detection |
| `internal/updater` | yt-dlp / ffmpeg availability probing and automatic download |
| `internal/ui/fyne` | Presentation only; avoid embedding download protocol details here |

Prefer small, testable helpers in `internal/core` and `internal/downloader` over large UI callbacks.

## Pull requests

1. **Scope** — One logical change per pull request when practical.
2. **Build** — `go build ./...` must succeed.
3. **Manual check** — Describe what you exercised in the PR.
4. **Documentation** — Update `README.md` when behaviour or requirements visible to users change.

## Commits

Use short, imperative subject lines (e.g. `Fix playlist progress label`). Avoid bundling unrelated refactors with functional fixes.

## Style

- Run `gofmt` (or `task fmt`) before submitting.
- Prefer explicit error handling over silent failure.
- Keep user-facing copy in English.
- Source comments and identifiers remain ASCII.
