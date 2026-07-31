# План улучшений (ветка `improvements`)

Результаты аудита приложения (см. анализ в задаче) → список изменений.
Каждый пункт — отдельный локальный коммит.

## 1. Безопасность загрузок: checksum + контекст + User-Agent

- `internal/updater/http.go`: `downloadFileForce` принимает `context.Context`,
  ставит `User-Agent: VAdlp`, удаляет `//nolint:noctx`.
- Новый `internal/updater/checksum.go`: загрузка списка SHA-256
  (`yt-dlp`: `SHA2-256SUMS`, `BtbN/FFmpeg-Builds`: `SHA2-256SUMS`,
  `denoland/deno`: `SHASUMS256.txt`), поиск записи по имени файла,
  проверка хэша скачанного файла до переименования.
- Если список чексумм недоступен — установка/обновление инструмента
  прерывается ошибкой (не можем подтвердить целостность).

Затрагивает: `checker.go` (yt-dlp), `ffmpeg.go`, `deno.go`.

## 2. Таймауты на probe-подпроцессы

- `internal/executil/cmd.go`: `CommandContext`/`OutputContext`.
- `internal/downloader/probe.go`: `ProbeCtx(ctx, cfg)` с таймаутом 30 c;
  `internal/service/service.go`: проброс контекста.
- `internal/updater/checker.go`: `probeExact` — таймаут 10 c.

## 3. Отмена по taskID вместо общего cancelFn

- `internal/app/app.go`: вместо одного `a.cancelFn` — map
  `jobID → context.CancelFunc` под мьютексом.
  `StopDownload` отменяет все запущенные задачи;
  `CancelQueueTask` — только свою (контекст + `downloader.CancelJob`).
- `internal/downloader/runner.go`: при `ctx.Done()` возвращать
  `ErrCancelled`, чтобы статус был `cancelled`, а не `error/context canceled`.

## 4. Уведомление о сбое запланированного запуска

- `ScheduleQueueRun`: если `RunQueue()` вернул ошибку (например,
  «download already running») — пишем в journal и шлём десктопное
  уведомление. Новая i18n-строка во всех 11 локалях.

## 5. Логирование проглоченных ошибок

- `main.go`: ошибку `settings.Load()` не терять — логировать в applog.
- `internal/app/app.go`: `_ =` → `applog.Info`/journal для
  `i18n.Init`, `applog.Init`, `instance.Register`, `settings.Save`,
  `core.SaveSession`, `MigrateToolsToConfigDir`, `core.AppendHistory`.

## 6. Progress/speed regex и фильтр логов

- `internal/downloader/runner.go`:
  - `progressRegex` требует маркер `[download]` (не ловит `%` в названиях);
  - `speedRegex` учитывает приблизительные скорости `~X/s`.
- `internal/app/app.go`: фильтр логов убирает только настоящие
  progress-строки, сохраняя `[download] Destination: ...` и строки с `%`.

## 7. GetLocales без повторного парсинга

- `internal/i18n/i18n.go`: `LocaleMap(lang)` с кэшем распарсенных файлов;
  `GetLocales` в app.go делегирует ему (убирает дублирующую структуру).

## 8. Фикс `.pre-commit-config.yaml`

- `gofmt -w ./cmd ./internal` → `gofmt -w ./internal ./tools`
  (каталога `cmd` нет).

## 9. Тесты для `internal/app`

- Юнит-тесты: очередь (Add/Remove/Reorder/Pause/Resume/Retry/Clear),
  DTO-конвертеры (round-trip), runJob на ошибке валидации (без бинарника),
  конфигурация сессии/журнала через test seams (`core.HistoryConfigDir`).

## 10. Разбить `App.tsx` ✅

- Вынести `EditQueueTaskModal` и все табы (Download/Network/Playlist/
  Extras/Queue/History/Settings) в отдельные компоненты
  `frontend/src/components/` (+`tabs/`), общие контролы и хелперы —
  в `FormControls.tsx`. `App.tsx` сокращён с ~1780 до ~1100 строк.

## 11. DTO-дедупликация: удалить ручной `types.ts` ✅

- Все DTO теперь импортируются из сгенерированного
  `frontend/src/wailsjs/go/models.ts` (`app.*`/`downloader.*`).
  Событийные типы (`DownloadProgressDTO`, `LocaleMap`), которых нет
  в генерации (не участвуют в сигнатурах методов), вынесены в
  `frontend/src/lib/eventTypes.ts`.
- `types.ts` удалён. Локальное конструирование `AppSettingsDTO`
  (defaults/depPaths/App) — через `new app.AppSettingsDTO(...)`,
  т.к. сгенерированный класс требует метод `convertValues`.
