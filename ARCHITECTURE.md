# Architecture

VAdlp is a desktop GUI around [yt-dlp](https://github.com/yt-dlp/yt-dlp), built with [Wails v2](https://wails.io/) (Go backend, React frontend). Layers stay separate so the UI never spawns processes directly — everything goes through `internal/app` → `internal/service` → `internal/downloader`.

## Layers

```
main.go               Wails entry: embeds frontend/dist, window options, OnBeforeClose (hide to tray)
internal/app          Wails-bound API the React frontend calls; owns in-memory queue/journal state, system tray
internal/service      download, probe, history, session (orchestration)
internal/core         Config, BuildCommand, profiles, history files, presets
internal/downloader   yt-dlp process execution, progress regex parsing, format probing
internal/updater      yt-dlp/ffmpeg/deno resolution & install/update, VAdlp's own GitHub release check
internal/health       startup health checks (missing deps, network) surfaced in the Health button
internal/settings     settings.json (versioned, migrated on load)
internal/configdir    OS-appropriate config directory (~/.config/vadlp or %AppData%\vadlp)
internal/i18n         locales (embedded JSON, served to the frontend by id)
internal/applog       vadlp.log in the config dir (opt-in debug logging)
internal/executil     process spawning helpers (hide console window on Windows, etc.)
frontend/             React UI (Bubble Tea / Charm-inspired styling), Wails JS bindings under src/wailsjs
tools/                CI/local build helpers (buildmeta, vadlp-build, releasepack)
```

`internal/app/app.go` is the only place that holds mutable application state (queue, journal, cached dependency info, scheduled queue timer). All exported methods on `*App` are bound to the frontend via `Bind` in `main.go` and called from `frontend/src/wailsjs/runtime.ts`'s `AppAPI` wrapper.

## Data flow (single download)

1. User edits `ConfigDTO` in the React UI; `PreviewCommand` round-trips it through `core.PreviewCommand` for the live yt-dlp command preview.
2. `RunDownload` calls `app.runJob`, which calls `service.Download(ctx, cfg, jobID, onEvent)`.
3. The service validates (`cfg.ValidateForDownload()`) and runs `downloader.RunCtx`, which spawns yt-dlp and parses its stdout for progress/playlist events.
4. Events come back through `onEvent` and are converted into `DownloadProgressDTO` pushed to the frontend via Wails events (`download:progress`, `download:log`).
5. On finish, `core.AppendHistory` records URL/status/duration, and `runJob` fires a desktop notification (`internal/app/tray.go`) if running as a standalone download (not part of a queue run).

## Data flow (queue run)

`RunQueue` snapshots all `queued` tasks, then either runs them sequentially or with `QueueParallel` workers (a semaphore-bounded goroutine pool). Each task runs through the same `runJob` as a single download, but with a `taskID` so progress is also tracked per task (`taskProgress` map in the frontend) and the task's status (`queued`/`paused`/`running`/`completed`/`error`/`cancelled`) is updated via `setTaskStatus` + a `queue:update` event. `runJob`'s `computeOverall()` nests a task's own playlist progress inside its slot in the queue, so the overall bar always spans the whole run. After the run, `notifyQueueDone` sends one summary notification (X done, Y failed).

A queue run can be delayed with `ScheduleQueueRun(atUnixMillis)`, which arms a `time.AfterFunc` and emits `queue:scheduled` so the frontend can show a countdown; `GetState`/`GetScheduledQueueRun` let the frontend recover that countdown after a reload.

## System tray and background operation

`internal/app/tray.go` starts [getlantern/systray](https://github.com/getlantern/systray) on its own goroutine from `Startup`. `main.go`'s `OnBeforeClose` hides the window instead of letting Wails quit, unless the user picked **Quit** from the tray menu (`App.ShouldQuit`/`allowQuit`). The tray icon itself is generated at runtime (`internal/app/trayicon.go`, a tiny PNG-in-ICO) so no binary asset needs to be tracked in the repo. Notifications use [gen2brain/beeep](https://github.com/gen2brain/beeep).

## Dependency resolution

`internal/updater.resolveTool` searches, in order: an explicit custom path → VAdlp's managed tools directory (`configdir.ToolsDir()`) → paths next to the executable → the current working directory's `bin/` → the system `PATH`. The same order applies to yt-dlp, ffmpeg, and deno, so a managed copy always wins over a stale system/pip install once one is installed via the Dependencies tab. `DependencyInfo.Status` can be `missing`, `found`, `unknown` (binary found but its version couldn't be parsed), `outdated`, `checking` (frontend-only, optimistic), or `error`.

## On disk (config dir)

| Path | Content |
|------|---------|
| `settings.json` | App settings (versioned, migrated), embedded `Config`, window layout, UI scale, theme, queue workers |
| `profiles/*.json` | Named configs (no URL) |
| `history.json` | Last downloads (capped) |
| `tools/` | Managed yt-dlp/ffmpeg/deno binaries installed via the Dependencies tab |
| `vadlp.log` | Debug/info log (opt-in via Settings) |

## Cancellation and shutdown

- **Stop** → `context.CancelFunc` stored on `*App` + `downloader.CancelJob(id)` for queue tasks.
- **Window close** → hidden to tray (see above), downloads keep running.
- **Real shutdown** (tray Quit) → `Shutdown` stops the tray, cancels any scheduled queue run, cancels all jobs, saves an in-flight session snapshot if one exists, and persists settings.

## i18n

Locales are embedded JSON arrays (`internal/i18n/locales/*.json`, 11 languages), served to the frontend as an id→string map via `GetLocales`. The frontend's `t()`/`tf()` helpers look up ids and interpolate `{{.Param}}` placeholders. CI checks that every locale file has the same keys as `en.json` (`internal/i18n/locale_parity_test.go`).
