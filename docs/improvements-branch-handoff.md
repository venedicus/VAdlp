# Ветка `improvements` — описание работ для следующего разработчика/AI

Документ описывает, что сделано в ветке `improvements` (база: `main` @ e085a88).
Цель — чтобы другой ИИ-агент мог быстро понять состояние кода, ключевые
решения и подводные камни, не разбирая каждый коммит заново.

## Общая картина

Ветка реализует план из `docs/improvements-plan.md` (11 пунктов, каждый —
отдельный локальный коммит) плюс внеплановый фикс зависания загрузки
(`3b977a1`, раздел 12). Все коммиты локальные, не запушены.

```
2574985 docs: plan improvements for the improvements branch
25d09e1 fix(updater): verify SHA-256 checksums, add context and User-Agent to tool downloads
6918219 fix: add timeouts to probe subprocesses
b73ee49 fix(app): cancel downloads per task instead of sharing one cancel func
438f4c1 fix(app): surface failures of scheduled queue runs
63dceff fix: log swallowed errors instead of discarding them
537315e fix(downloader): only treat [download] lines as progress, keep useful log lines
faf71e6 refactor(i18n): cache parsed locale maps for GetLocales
c249555 chore: fix stale gofmt path in pre-commit config
86a2ae4 test(app): cover queue orchestration, cancellation and DTO conversion
049ea12 refactor(frontend): split App.tsx into components
249dcb1 refactor(frontend): use generated Go models, drop manual types.ts
3b977a1 fix(downloader): drain stdout and stderr concurrently, kill subprocess on cancel
```

Проверка: `go test ./...`, `go vet ./...`, `gofmt -l ./internal ./tools`,
`go test -race ./internal/app/`, `npm run build` + `npm test` в `frontend/`
(46 vitest-тестов) — всё зелёное.

## Backend (Go)

### 1. Безопасность загрузок инструментов — `25d09e1`
- `internal/updater/http.go`: `downloadFileForce` принимает `context.Context`,
  `User-Agent: VAdlp`, снят `//nolint:noctx`.
- Новый `internal/updater/checksum.go`: SHA-256-верификация скачанных
  yt-dlp/ffmpeg/deno до переименования `.part` → финал.
  Ключевые символы: `fetchChecksum`, `verifyChecksum`, `verifyDownload`,
  `parseChecksumLine`, `ytDlpChecksumURL`/`ffmpegChecksumURL`/`denoChecksumURL`,
  лимит `maxChecksumManifestSize` (4 MiB). Список чексумм недоступен →
  установка прерывается ошибкой.
- `checker.go`, `ffmpeg.go`, `deno.go` прогоняют загрузку через
  `verifyDownload`.

### 2. Таймауты probe-подпроцессов — `6918219`
- `internal/executil/cmd.go`: `CommandContext`/`OutputContext` (контекст +
  таймаут, без `nolint`).
- `internal/downloader/probe.go`: `ProbeCtx(ctx, cfg)` — таймаут 30 c;
  `internal/service/service.go` пробрасывает контекст.
- `internal/updater/checker.go`: `probeExact` — таймаут 10 c.

### 3. Отмена по taskID — `b73ee49`
- `internal/app/app.go`: вместо одного `cancelFn` — map
  `runCancels[jobID] → context.CancelFunc` под мьютексом.
  `jobID` = `taskID` либо `"main"` (для одиночного скачивания).
- `StopDownload` отменяет все запущенные задачи; `CancelQueueTask(id)`
  отменяет только свою (контекст + фолбэк на `downloader.CancelJob(id)`).
  Возвращает `true`, если контекст найден и отменён.
- `internal/downloader/runner.go`: при `ctx.Done()` — `ErrCancelled`
  (статус `cancelled`, а не `error/context canceled`).

### 4. Ошибки запланированного запуска — `438f4c1`
- `ScheduleQueueRun`: ошибка `RunQueue()` → запись в journal + десктопное
  уведомление. Новый i18n-ключ `err.schedule_failed` — во всех 11 локалях
  (`i18n/locales/*.json`).

### 5. Логирование проглоченных ошибок — `63dceff`
- `main.go`: ошибка `settings.Load()` больше не теряется.
- `internal/app/app.go`: `_ =` → `applog.Info`/journal для `i18n.Init`,
  `applog.Init`, `instance.Register`, `settings.Save`, `core.SaveSession`,
  `MigrateToolsToConfigDir`, `core.AppendHistory`.
- `tray.go`: ошибка notify логируется (test seam `notifyFn`).

### 6. Progress/speed regex и фильтр логов — `537315e`
- `internal/downloader/runner.go`:
  - `progressRegex` требует маркер `[download]` (не ловит `%` в названиях);
  - `speedRegex` учитывает `~X/s`;
  - экспортирован `IsProgressLine(line string) bool`.
- `internal/app/app.go`: фильтр логов удаляет только настоящие
  progress-строки; строки вроде `[download] Destination: ...` и строки с `%`
  сохраняются.

### 7. Кэш локалей — `faf71e6`
- `internal/i18n/i18n.go`: `LocaleMap(lang)` с кэшем распарсенных файлов
  (глобальный sync.Map, инвалидация не требуется — файлы статичны).
  `GetLocales` в app.go делегирует ему.

### 8. Фикс pre-commit — `c249555`
- `.pre-commit-config.yaml`: `gofmt -w ./cmd ./internal` →
  `gofmt -w ./internal ./tools` (каталога `cmd` нет).

### 9. Тесты `internal/app` — `86a2ae4`
- `internal/app/app_test.go`: 12+ тестов — очередь
  (Add/Remove/Reorder/Pause/Resume/Retry/Clear), DTO round-trip,
  runJob на ошибке валидации (без бинарника), сессии/журнал.
- **Test seams** (важно для новых тестов):
  - `App.emitEvent` (в `New()` обёрнут вокруг `runtime.EventsEmit` —
    вне Wails-контекста `runtime.EventsEmit` делает `log.Fatalf`,
    поэтому в тестах ОБЯЗАТЕЛЬНО подменять `emitEvent`);
  - `App.notifyFn` (по умолчанию `defaultNotify` из `tray.go`);
  - `core.SetHistoryDir(dir string)` — переопределяет каталог истории;
  - `settings.SetConfigDir(dir string)` — переопределяет каталог настроек.
- Найден и исправлен мелкий баг: `CancelQueueTask` при jobID-коллизии
  (в `b73ee49` jobID уже уникален; в `86a2ae4` добавлен ранний выход из
  runJob при уже отменённом контексте).

## Frontend (TypeScript/React)

### 10. Разбивка `App.tsx` — `049ea12`
- `App.tsx` сокращён ~1780 → ~1100 строк; остались только состояние,
  эффекты, шапка, tab-bar, modals и роутинг табов.
- Новые файлы:
  - `components/FormControls.tsx` — `Field`, `Check`, `extractDroppedURL`,
    `looksLikeDownloadableURL`, `formatCountdown`;
  - `components/EditQueueTaskModal.tsx`;
  - `components/tabs/{DownloadTab,NetworkTab,PlaylistTab,ExtrasTab,QueueTab,HistoryTab,SettingsTab}.tsx`;
  - `components/Modal.tsx` содержит собственный `Field` (не путать с
    FormControls — у него класс `modal-field`).
- Состояние, переехавшее в компоненты: `qualityPresetKey` → DownloadTab,
  `selectedQueueId`/drag-состояние → QueueTab, `historyQuery`/
  `historyStatusFilter` → HistoryTab. `scheduleInput`/`now`/`scheduledQueueAt`
  остались в App (нужны эффекту таймера).
- Реorder очереди: App отдаёт `onReorderQueue(ids)` — оптимистичный
  `setQueue` + `AppAPI.ReorderQueue` (поведение как в оригинале).

### 11. DTO-дедупликация — `249dcb1`
- `frontend/src/types.ts` удалён. Все DTO импортируются из сгенерированного
  `frontend/src/wailsjs/go/models.ts`:
  - `app.ConfigDTO`, `app.AppSettingsDTO`, `app.QueueTaskDTO`,
    `app.DependencyDTO`, `app.HistoryItemDTO`, `app.HealthIssueDTO`,
    `app.AppUpdateDTO`, `app.InstanceDTO`, `app.InstallGuardDTO` и др.;
  - `downloader.Format`, `downloader.MediaEntry`, `downloader.ProbeResult`
    (бывшие `FormatDTO`, `MediaEntryDTO`, `ProbeResultDTO`).
- Событийные типы, которых НЕТ в генерации (не участвуют в сигнатурах
  Go-методов): `DownloadProgressDTO`, `LocaleMap` — переехали в
  `frontend/src/lib/eventTypes.ts`.
- `wailsjs/runtime.ts` — рукописная обёртка (НЕ генерируется) — теперь
  импортирует типы из `go/models` + `lib/eventTypes`, реэкспортирует
  `DownloadProgressDTO`.

**Подводный камень (важно!):** сгенерированные классы с вложенными
объектами (`AppSettingsDTO`, `QueueTaskDTO`, `AppStateDTO`, `ProfileDTO`,
`downloader.MediaEntry`, `downloader.ProbeResult`) содержат обязательный
метод `convertValues`. Обычные объектные литералы им структурно НЕ
присваиваются. Поэтому локальное конструирование настроек делается через
`new app.AppSettingsDTO({...})`:
- `lib/defaults.ts` — `defaultSettings()`, `normalizeSettings()`;
- `lib/depPaths.ts` — `mergeSettingsPatch()`;
- `App.tsx` — `updateConfig`/`updateSettings`, `handleLanguageChange`,
  `resetWindowSize`;
- `lib/depPaths.test.ts` — фабрики тестовых настроек.
- `ConfigDTO` и прочие «плоские» классы методов не имеют — литералы ок.
- Реальные значения из `window.go.app.App.*` приходят как plain JSON
  (bindings в `wailsjs/go/app/App.js` не делают `createFrom`), т.е.
  `convertValues` в рантайме нет — конфликтов нет, т.к. он нигде не
  вызывается.

### 12. Исправление зависания загрузки (deadlock) — `3b977a1`
- **Симптом (репорт пользователя):** загрузка «встаёт» — раньше каждые
  20–30 видео, после пунктов 1–11 редко, но всё ещё случается (~665 видео
  за сессию). Отмена НЕ работала как операция — только перезапуск
  приложения.
- **Причина:** `RunCtx` читал stdout и stderr ПОСЛЕДОВАТЕЛЬНО через
  `io.MultiReader(stdout, stderr)` — пока stdout не закрыт, stderr не
  читается вообще. Если yt-dlp забивает буфер stderr-пайпа (4 КБ), а
  stdout молчит, процесс блокируется на записи в stderr, а приложение —
  на чтении stdout → вечный взаимный блок. Главный цикл был заблокирован
  внутри `scanner.Scan()`, поэтому ни `ctx.Done()`, ни `CancelJob` не
  доходили до процесса.
- **Исправление (`internal/downloader/runner.go`):**
  - stdout/stderr читаются двумя pump-горутинами в общий канал `lines`
    (буфер 256 строк) — deadlock невозможен;
  - киллер-горутина убивает субпроцесс сразу при `ctx.Done()` — даже
    когда главный цикл заблокирован в pipe `Read` или `cmd.Wait()`;
  - главный цикл — `select` на `lines`/`ctx.Done()`; потоковая часть
    вынесена в тестируемый `runCommand(ctx, jobID, cmd, onEvent)`
    (`cmd *exec.Cmd`), `RunCtx` делает резолюцию бинарника и аргументы;
  - `splitLines` — пакетная функция (та же семантика, что у старого
    inline-Split в сканере);
  - `playlistRegex`: убрана пустая альтернатива в группе префикса
    `(?:Downloading ... |)`. Теперь требуется
    `[download] Downloading (item|video) N of M` — строки вроде
    «[download] Destination: Part 1 of 3.mp4» больше не дают ложных
    playlist-событий (искажали текущий/общий счётчик и общий прогресс).
- **Тесты:**
  - `internal/downloader/pipe_test.go` — helper-процесс
    (`TestHelperProcess`, env `VADLP_TEST_HELPER`), перезапуск тест-бинарника
    через `os.Args[0] -test.run=TestHelperProcess`:
    - `stderr-burst`: 4 МиБ в stderr при пустом stdout → старый код
      завис бы навсегда, новый завершается (15-сек. watchdog в тесте);
    - `stderr-forever`: вечная запись в stderr; отмена контекста →
      процесс убивается, возвращается `ErrCancelled` (старый код —
      бесконечное ожидание);
  - `runner_test.go`: позитивные (item/video) и негативные
    (`TestPlaylistRegexNoFalsePositive`) кейсы playlistRegex.
- Проверка: `go vet ./...`, `go test ./...` (все пакеты), `wails build` —
  зелёные; привязки не изменились (только Go-сторона).

### 13. Reap субпроцесса при отмене — `a37f8a6`
- Найдено при ревью раздела 12: оба пути отмены в `runCommand` убивали
  процесс и возвращались **без** `cmd.Wait()`, т.е. дочерний процесс не
  reaped — зомби на Unix, утечка handle на Windows, накапливается за
  сессию при каждой отмене задачи.
- `terminate()`: kill → слив канала `lines` → `cmd.Wait()`.
  Порядок важен: `cmd.Wait` закрывает пайпы, из которых читают pump'ы,
  поэтому Wait не может идти первым; слив канала заодно разблокирует
  pump, застрявший на отправке при полном буфере.

## Что НЕ сделано / на заметку
- Ветка не запушена, PR не создавался.
- `docs/improvements-plan.md` помечен ✅ для выполненных пунктов.
- LF/CRLF предупреждения git при коммитах — ожидаемо (`.gitattributes`
  не настроен), не ошибка.
- При следующем `wails generate module/models` модели перегенерируются —
  проверять, что `models.ts` не потерял типы; `eventTypes.ts` остаётся
  ручным.
- Если понадобится новый DTO в событиях — он не появится в `models.ts`
  автоматически, только через сигнатуру bound-метода.

## Местный план для ветки

Статус работ в ветке `improvements` и ближайшие шаги. План — живой
документ: пункты помечаются при выполнении.

### Выполнено
- [x] Пункты 1–11 плана `docs/improvements-plan.md` (коммиты см. в начале
  документа) — каждый пункт отдельным коммитом, всё проверено
  (`go test ./...`, `go vet ./...`, `gofmt -l`, `go test -race
  ./internal/app/`, `npm run build` + `npm test` (46 vitest-тестов),
  `wails build`).
- [x] Устранён deadlock загрузки — `3b977a1` (см. раздел 12): параллельный
  слив stdout/stderr, убийство субпроцесса при отмене, фикс `playlistRegex`;
  регрессионные тесты `pipe_test.go` (helper-процесс), негативные кейсы
  playlistRegex.
- [x] Приложение пересобрано (`wails build`) и запускалось для проверки
  пользователем (общая проверка: 665 видео за сессию без сбоев, отмена
  работает).
- [x] Ревью ветки: план и этот документ сверены с кодом, все проверки
  перезапущены независимо (`go build/vet/gofmt`, `go test -race ./...`,
  `npm run build`, 46 vitest-тестов) — зелёные.
- [x] Reap субпроцесса при отмене — `a37f8a6` (раздел 13), найдено на ревью.
- [x] Судьба `docs/feature-gallery-dl.md`: вынесен в отдельную ветку
  `feature/gallery-dl`, из этой ветки исключён.
- [x] Проверено, что `main` (e085a88) — прямой предок ветки: мёрж
  fast-forward, конфликтов нет.

### В очереди / на решение
- [ ] Живая проверка фикса deadlock пользователем: длительная загрузка
  плейлиста + отмена задачи в процессе загрузки (ранее: замерзание +
  перезапуск приложения). После подтверждения — пометить выполненным.
- [ ] (Опционально) добавить пункт в `docs/improvements-plan.md` о фиксе
  deadlock, если план будет продлеваться.

### Известные ограничения (осознанные, не баги ветки)
- Проверка чексумм (`25d09e1`) берёт манифест с того же GitHub и без
  подписи: защищает от битой/оборванной загрузки и подмены в канале,
  но не от компрометации самого релиза.

### Критерий завершения ветки
- Все пункты плана ✅, живая проверка deadlock-фикса пройдена, тесты и
  сборка зелёные, `docs/improvements-branch-handoff.md` актуален.
