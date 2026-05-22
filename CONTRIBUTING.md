# Contributing

## Setup

Requirements are in [README.md](README.md): Go 1.25+, yt-dlp and ffmpeg for manual tests, gcc when building the UI.

```bash
git clone https://github.com/venedicus/VAdlp.git
cd VAdlp
task run
```

Without Task: `./build.sh` or `.\build.ps1` (wraps `scripts/build-metadata.sh` + `go build -o bin/...`).

### Linux build deps

```bash
sudo apt-get install gcc libgl1-mesa-dev xorg-dev libxkbcommon-dev
```

## Layout

| Path | |
|------|--|
| `internal/core` | Config, yt-dlp args, profiles, session |
| `internal/downloader` | Subprocess, logs, progress |
| `internal/updater` | Binary checks and downloads |
| `internal/i18n` | Strings |
| `internal/ui/fyne` | UI only |
| `internal/version` | Version injected at build time |

Keep download logic out of the UI package when you can.

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
| Build | `task build` |

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
