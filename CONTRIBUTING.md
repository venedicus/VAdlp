# Contributing

## Setup

Requirements are in [README.md](README.md): Go 1.25+, Node.js 18+, yt-dlp and ffmpeg for manual tests, gcc when building the UI (Wails needs CGO).

```bash
git clone https://github.com/venedicus/VAdlp.git
cd VAdlp
task dev      # hot-reload dev mode
task run      # build into build/bin/ and run it
```

Without [Task](https://taskfile.dev): `cd frontend && npm install && cd ..` once, then `go run github.com/wailsapp/wails/v2/cmd/wails@latest dev` (or `build -nopackage`) — see [Taskfile.yml](Taskfile.yml) for the exact commands Task wraps.

### Linux build deps

```bash
sudo apt-get install gcc libgtk-3-dev libwebkit2gtk-4.1-dev
```

## Layout

| Path | |
|------|--|
| `internal/app` | Wails-bound backend API, queue/journal state, system tray |
| `internal/core` | Config, yt-dlp args, profiles, session |
| `internal/downloader` | Subprocess, logs, progress |
| `internal/updater` | Binary checks/downloads for yt-dlp, ffmpeg, deno, and VAdlp itself |
| `internal/health` | Startup health checks (Health button in the UI) |
| `internal/i18n` | Strings (`locales/en.json`, `locales/ru.json`) |
| `internal/version` | Version injected at build time (ldflags) |
| `frontend/` | React UI — components in `src/components`, app shell in `src/App.tsx` |
| `tools/` | Build/release helpers used by CI and Taskfile |

Keep download logic out of `internal/app` and the UI when you can — it should stay a thin binding layer over `internal/core`/`internal/downloader`/`internal/service`.

## Before a pull request

```bash
task check
```

Or step by step:

| Task | Command |
|------|---------|
| Format | `task fmt` |
| Vet | `task vet` |
| Lint | `task lint` (needs [golangci-lint](https://golangci-lint.run/)) |
| Test | `task test` |
| Build | `task build:app` |

CI runs the same checks plus `govulncheck` and cross-platform builds. Match CI locally when you can.

## Pull requests

- One change per PR when possible
- Fill in the PR template (summary + test plan)
- Note what you tested manually (OS, download scenario)
- Update README if user-visible behaviour changes
- Add `en.json` and `ru.json` keys for new UI strings

Labels are applied automatically (`ui`, `core`, `ci`, …) when possible. Create missing labels in the repo if the labeler workflow warns.

## Commits

Short imperative subject (`Fix queue cancel`). No drive-by refactors mixed with fixes.

## Style

- `gofmt` before push (`task fmt`)
- Don't swallow errors without a reason
- Prefer fixing linter findings over disabling rules

## Releases (maintainers)

See [RELEASE.md](RELEASE.md). Tag `v*` on `main` after CI is green.

## Security

See [SECURITY.md](SECURITY.md). Do not file public issues for unpatched vulnerabilities.
