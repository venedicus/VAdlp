# Architecture

VAdlp is a desktop GUI around [yt-dlp](https://github.com/yt-dlp/yt-dlp). Layers stay separate so the UI does not spawn processes directly.

## Layers

```
cmd/vadlp          → entry (Fyne app)
internal/ui/fyne   → widgets, tabs, queue UI, i18n bind
internal/service   → download, probe, history, session (orchestration)
internal/core      → Config, BuildCommand, profiles, history files
internal/downloader→ yt-dlp process, progress regex, probe JSON
internal/settings  → settings.json (versioned)
internal/configdir → ~/.config/vadlp (or OS equivalent)
internal/i18n      → en/ru locales (embedded)
internal/applog    → vadlp.log in config dir
```

## Data flow (download)

1. User edits `core.Config` in the UI (live command preview via `core.PreviewCommand`).
2. **Download** calls `service.Download(ctx, cfg, jobID, onEvent)`.
3. Service runs `cfg.ValidateForDownload()`, then `downloader.RunCtx`.
4. Events update progress/log badges; on finish, `core.AppendHistory` records URL, status, duration.
5. Optional `core.SaveSession` throttles playlist/queue snapshot to disk.

## On disk

| Path | Content |
|------|---------|
| `settings.json` | App settings v4, embedded config, window layout, UI scale (0 = auto), queue workers |
| `profiles/*.json` | Named configs (no URL) |
| `history.json` | Last downloads |
| `vadlp.log` | Debug/info log (optional debug mode in Tools) |

## Cancellation

- **Stop** → `context.CancelFunc` + `downloader.CancelJob(id)`.
- **Window close** → `CancelAll()` + last session save if a job was running.

## i18n

Locales are embedded JSON arrays. `LocaleBinder` refreshes labels on language change without restart. CI checks en/ru key parity (`internal/i18n/locale_parity_test.go`).
