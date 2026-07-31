package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	goruntime "runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"vadlp/internal/applog"
	"vadlp/internal/browse"
	"vadlp/internal/configdir"
	"vadlp/internal/core"
	"vadlp/internal/downloader"
	"vadlp/internal/health"
	"vadlp/internal/i18n"
	"vadlp/internal/instance"
	"vadlp/internal/service"
	"vadlp/internal/settings"
	"vadlp/internal/updater"
	"vadlp/internal/version"
)

type App struct {
	ctx context.Context

	mu          sync.RWMutex
	appSettings settings.App
	queue       []QueueTaskDTO
	journal     []string
	running     atomic.Bool
	runCtxMu    sync.Mutex
	runCancels  map[string]context.CancelFunc
	svc         *service.Service

	// emitEvent and notifyFn are swappable for tests; the Wails runtime
	// (runtime.EventsEmit / desktop notifications) is not available there.
	emitEvent func(name string, data ...interface{})
	notifyFn  func(title, message string)

	depsCacheMu sync.RWMutex
	cachedDeps  []updater.DependencyInfo

	sessionSnapMu   sync.Mutex
	lastSessionSnap core.Session
	hasSessionSnap  bool

	scheduleMu    sync.Mutex
	scheduledAt   time.Time
	scheduleTimer *time.Timer

	allowQuit atomic.Bool
}

type QueueTaskDTO struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Config ConfigDTO `json:"config"`
	Status string    `json:"status"`
}

type AppStateDTO struct {
	Settings         AppSettingsDTO `json:"settings"`
	Queue            []QueueTaskDTO `json:"queue"`
	Journal          []string       `json:"journal"`
	Running          bool           `json:"running"`
	Version          string         `json:"version"`
	ToolsDir         string         `json:"toolsDir"`
	ScheduledQueueAt int64          `json:"scheduledQueueAt"`
}

type AppSettingsDTO struct {
	Config              ConfigDTO `json:"config"`
	FFmpegPath          string    `json:"ffmpegPath"`
	SessionPath         string    `json:"sessionPath"`
	QueueParallel       int       `json:"queueParallel"`
	Language            string    `json:"language"`
	YtDlpPath           string    `json:"ytDlpPath"`
	DenoPath            string    `json:"denoPath"`
	LastProfile         string    `json:"lastProfile"`
	DebugLog            bool      `json:"debugLog"`
	ActivityPanelOpen   bool      `json:"activityPanelOpen"`
	UIScale             float32   `json:"uiScale"`
	Theme               string    `json:"theme"`
	WindowWidth         float32   `json:"windowWidth"`
	WindowHeight        float32   `json:"windowHeight"`
	ActivityPanelOffset float64   `json:"activityPanelOffset"`
	// BrowserEnabled turns on the optional Browse tab (yt-dlp-backed search).
	// Off by default: the module is opt-in and costs an extractor call per
	// search.
	BrowserEnabled bool `json:"browserEnabled"`
}

type ConfigDTO struct {
	URL                 string `json:"url"`
	Quality             string `json:"quality"`
	Format              string `json:"format"`
	AudioOnly           bool   `json:"audioOnly"`
	AudioFormat         string `json:"audioFormat"`
	OutputPath          string `json:"outputPath"`
	OutputTemplate      string `json:"outputTemplate"`
	UseCookiesFile      bool   `json:"useCookiesFile"`
	CookiesFile         string `json:"cookiesFile"`
	UseCookiesBrowser   bool   `json:"useCookiesBrowser"`
	CookiesBrowser      string `json:"cookiesBrowser"`
	Proxy               string `json:"proxy"`
	RateLimit           string `json:"rateLimit"`
	PlaylistReverse     bool   `json:"playlistReverse"`
	Continue            bool   `json:"continue"`
	NoPart              bool   `json:"noPart"`
	PlaylistStart       int    `json:"playlistStart"`
	PlaylistEnd         int    `json:"playlistEnd"`
	MaxDownloads        int    `json:"maxDownloads"`
	DownloadArchive     string `json:"downloadArchive"`
	NoPlaylist          bool   `json:"noPlaylist"`
	FlatPlaylist        bool   `json:"flatPlaylist"`
	WriteSubs           bool   `json:"writeSubs"`
	WriteAutoSub        bool   `json:"writeAutoSub"`
	EmbedSubs           bool   `json:"embedSubs"`
	SubLangs            string `json:"subLangs"`
	WriteThumbnail      bool   `json:"writeThumbnail"`
	EmbedThumbnail      bool   `json:"embedThumbnail"`
	EmbedMetadata       bool   `json:"embedMetadata"`
	EmbedChapters       bool   `json:"embedChapters"`
	Retries             int    `json:"retries"`
	FragmentRetries     int    `json:"fragmentRetries"`
	ConcurrentFragments int    `json:"concurrentFragments"`
	SocketTimeout       int    `json:"socketTimeout"`
	NoWarnings          bool   `json:"noWarnings"`
	Verbose             bool   `json:"verbose"`
	Quiet               bool   `json:"quiet"`
	WriteInfoJSON       bool   `json:"writeInfoJSON"`
	LoadInfoJSON        string `json:"loadInfoJson"`
	WindowsFilenames    bool   `json:"windowsFilenames"`
	NoMtime             bool   `json:"noMtime"`
	AbortOnError        bool   `json:"abortOnError"`
	IgnoreErrors        bool   `json:"ignoreErrors"`
	ExtraArgs           string `json:"extraArgs"`
	FFmpegLocation      string `json:"ffmpegLocation"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	SponsorBlockRemove  bool   `json:"sponsorBlockRemove"`
	BatchURLs           string `json:"batchUrls"`
	YtDlpPath           string `json:"ytDlpPath"`
	DenoPath            string `json:"denoPath"`
}

type DependencyDTO struct {
	ID          string `json:"id"`
	Level       string `json:"level"`
	Status      string `json:"status"`
	Path        string `json:"path"`
	Version     string `json:"version"`
	LatestVer   string `json:"latestVer"`
	UpdateAvail bool   `json:"updateAvail"`
	Source      string `json:"source"`
	Error       string `json:"error"`
}

type InstallGuardDTO struct {
	AlreadyInstalled bool   `json:"alreadyInstalled"`
	FoundElsewhere   bool   `json:"foundElsewhere"`
	Path             string `json:"path"`
}

type HistoryItemDTO struct {
	At          string `json:"at"`
	URL         string `json:"url"`
	Status      string `json:"status"`
	Output      string `json:"output"`
	Error       string `json:"error"`
	DurationSec int    `json:"durationSec"`
}

type HealthIssueDTO struct {
	ID       string                 `json:"id"`
	Severity string                 `json:"severity"`
	Key      string                 `json:"key"`
	Params   map[string]interface{} `json:"params"`
}

type ProfileDTO struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Config      ConfigDTO `json:"config"`
}

type DownloadProgressDTO struct {
	FilePct    float64 `json:"filePct"`
	OverallPct float64 `json:"overallPct"`
	Speed      string  `json:"speed"`
	ETA        string  `json:"eta"`
	Status     string  `json:"status"`
	Phase      string  `json:"phase"`
	PlCurrent  int     `json:"plCurrent"`
	PlTotal    int     `json:"plTotal"`
	QueueIdx   int     `json:"queueIdx"`
	QueueTotal int     `json:"queueTotal"`
	TaskID     string  `json:"taskId"`
}

func New() *App {
	a := &App{
		svc:        service.New(),
		runCancels: map[string]context.CancelFunc{},
		notifyFn:   defaultNotify,
	}
	a.emitEvent = func(name string, data ...interface{}) {
		runtime.EventsEmit(a.ctx, name, data...)
	}
	return a
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	if err := updater.MigrateToolsToConfigDir(); err != nil {
		applog.Info("migrate tools failed", "err", err.Error())
	}
	appSettings, err := settings.Load()
	if err != nil {
		applog.Info("load settings failed", "err", err.Error())
	}
	a.mu.Lock()
	a.appSettings = appSettings
	if a.appSettings.YtDlpPath != "" {
		a.appSettings.Config.YtDlpPath = a.appSettings.YtDlpPath
	}
	if a.appSettings.DenoPath != "" {
		a.appSettings.Config.DenoPath = a.appSettings.DenoPath
	}
	if a.appSettings.FFmpegPath != "" {
		a.appSettings.Config.FFmpegLocation = a.appSettings.FFmpegPath
	} else if a.appSettings.Config.FFmpegLocation != "" {
		a.appSettings.FFmpegPath = a.appSettings.Config.FFmpegLocation
	}
	a.mu.Unlock()
	lang := appSettings.Language
	if lang == "" {
		lang = "en"
	}
	if err := i18n.Init(lang); err != nil {
		applog.Info("i18n init failed", "err", err.Error())
	}
	if err := applog.Init(appSettings.DebugLog); err != nil {
		applog.Info("applog init failed", "err", err.Error())
	}
	go a.checkStartupDeps()
	a.startTray()
	a.startInstanceMonitor()
}

// startInstanceMonitor registers this process in the shared instance registry,
// notifies the UI if sibling instances are already running, and keeps a
// heartbeat (with busy status) going so other instances can see this one.
func (a *App) startInstanceMonitor() {
	if err := instance.Register(); err != nil {
		applog.Info("instance register failed", "err", err.Error())
	}
	if others, err := instance.List(); err == nil && len(others) > 0 {
		a.emitEvent("startup:other-instances", instanceDTOs(others))
	}
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := instance.Heartbeat(a.running.Load()); err != nil {
				applog.Info("instance heartbeat failed", "err", err.Error())
			}
			if instance.ShouldQuit() {
				runtime.Quit(a.ctx)
				return
			}
		}
	}()
}

// ShouldQuit reports whether the user explicitly chose Quit from the tray menu,
// as opposed to closing the window (which hides to tray instead).
func (a *App) ShouldQuit() bool {
	return a.allowQuit.Load()
}

func (a *App) checkStartupDeps() {
	paths := a.dependencyPaths()
	result := updater.CheckToolsWithPaths(paths)
	if !result.YtDlp.Found {
		a.emitEvent("startup:ytdlp-missing")
		return
	}
	a.addJournal(i18n.T("log.ytdlp_ok", map[string]interface{}{
		"Path":    result.YtDlp.Path,
		"Version": result.YtDlp.Version,
	}), nil)
	go func() {
		infos, err := updater.RefreshLatestVersions(a.ctx, paths)
		if err != nil {
			a.addJournal(i18n.T("dep.check_failed", nil), err)
			return
		}
		a.setCachedDeps(infos)
	}()
}

func (a *App) GetState() AppStateDTO {
	a.mu.RLock()
	defer a.mu.RUnlock()
	toolsDir, _ := configdir.ToolsDir()
	return AppStateDTO{
		Settings:         settingsToDTO(a.appSettings),
		Queue:            append([]QueueTaskDTO(nil), a.queue...),
		Journal:          append([]string(nil), a.journal...),
		Running:          a.running.Load(),
		Version:          version.Label(),
		ToolsDir:         toolsDir,
		ScheduledQueueAt: a.GetScheduledQueueRun(),
	}
}

func (a *App) SaveSettings(dto AppSettingsDTO) error {
	a.mu.Lock()
	a.appSettings = dtoToSettings(dto)
	a.mu.Unlock()
	if err := settings.Save(a.appSettings); err != nil {
		return err
	}
	if lang := strings.TrimSpace(dto.Language); lang != "" {
		if err := i18n.Init(lang); err != nil {
			applog.Info("i18n init failed", "err", err.Error())
		}
	}
	if err := applog.Init(dto.DebugLog); err != nil {
		applog.Info("applog init failed", "err", err.Error())
	}
	return nil
}

func (a *App) PreviewCommand(cfg ConfigDTO) string {
	c := dtoToConfig(cfg)
	prog := "yt-dlp"
	if p, err := downloader.ResolveBinary(c.YtDlpPath); err == nil {
		prog = p
	}
	return core.PreviewCommand(c, prog)
}

func (a *App) ApplyPreset(presetID string) ConfigDTO {
	a.mu.Lock()
	defer a.mu.Unlock()
	cfg := a.appSettings.Config
	if core.ApplyPreset(&cfg, presetID) {
		a.appSettings.Config = cfg
	}
	return configToDTO(cfg)
}

func (a *App) RunDownload(cfg ConfigDTO) error {
	if !a.running.CompareAndSwap(false, true) {
		return fmt.Errorf("download already running")
	}
	go func() {
		defer a.running.Store(false)
		a.runJob(dtoToConfig(cfg), "", 1, 1, true)
		a.saveSettingsLocked()
	}()
	return nil
}

// StopDownload cancels every currently running download job.
func (a *App) StopDownload() {
	a.runCtxMu.Lock()
	cancels := make([]context.CancelFunc, 0, len(a.runCancels))
	for _, c := range a.runCancels {
		cancels = append(cancels, c)
	}
	a.runCtxMu.Unlock()
	for _, c := range cancels {
		c()
	}
}

func (a *App) AddToQueue(cfg ConfigDTO) (QueueTaskDTO, error) {
	c := dtoToConfig(cfg)
	if strings.TrimSpace(c.URL) == "" && strings.TrimSpace(c.LoadInfoJSON) == "" && strings.TrimSpace(c.BatchURLs) == "" {
		return QueueTaskDTO{}, core.ValidationError{Key: "err.queue_no_url"}
	}
	task := QueueTaskDTO{
		ID:     fmt.Sprintf("q-%d", time.Now().UnixNano()),
		Name:   firstNonEmpty(c.URL, c.LoadInfoJSON, c.BatchURLs),
		Config: cfg,
		Status: "queued",
	}
	a.mu.Lock()
	a.queue = append(a.queue, task)
	a.mu.Unlock()
	a.emitQueue()
	return task, nil
}

func (a *App) RemoveFromQueue(id string) {
	a.mu.Lock()
	for i, t := range a.queue {
		if t.ID == id {
			a.queue = append(a.queue[:i], a.queue[i+1:]...)
			break
		}
	}
	a.mu.Unlock()
	a.emitQueue()
}

func (a *App) MoveQueueItem(id string, direction int) {
	a.mu.Lock()
	idx := -1
	for i, t := range a.queue {
		if t.ID == id {
			idx = i
			break
		}
	}
	if idx >= 0 {
		j := idx + direction
		if j >= 0 && j < len(a.queue) {
			a.queue[idx], a.queue[j] = a.queue[j], a.queue[idx]
		}
	}
	a.mu.Unlock()
	a.emitQueue()
}

func (a *App) UpdateQueueTask(id string, cfg ConfigDTO) error {
	c := dtoToConfig(cfg)
	if strings.TrimSpace(c.URL) == "" && strings.TrimSpace(c.LoadInfoJSON) == "" && strings.TrimSpace(c.BatchURLs) == "" {
		return core.ValidationError{Key: "err.queue_no_url"}
	}
	a.mu.Lock()
	found := false
	for i := range a.queue {
		if a.queue[i].ID == id && (a.queue[i].Status == "queued" || a.queue[i].Status == "paused") {
			a.queue[i].Config = cfg
			a.queue[i].Name = firstNonEmpty(c.URL, c.LoadInfoJSON, c.BatchURLs)
			found = true
			break
		}
	}
	a.mu.Unlock()
	if !found {
		return fmt.Errorf("task not editable")
	}
	a.emitQueue()
	return nil
}

func (a *App) ReorderQueue(ids []string) {
	a.mu.Lock()
	index := make(map[string]int, len(a.queue))
	for i, t := range a.queue {
		index[t.ID] = i
	}
	reordered := make([]QueueTaskDTO, 0, len(a.queue))
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if i, ok := index[id]; ok && !seen[id] {
			reordered = append(reordered, a.queue[i])
			seen[id] = true
		}
	}
	for _, t := range a.queue {
		if !seen[t.ID] {
			reordered = append(reordered, t)
		}
	}
	a.queue = reordered
	a.mu.Unlock()
	a.emitQueue()
}

func (a *App) ClearQueue() {
	a.mu.Lock()
	a.queue = nil
	a.mu.Unlock()
	a.emitQueue()
}

func (a *App) PauseQueueTask(id string) {
	a.mu.Lock()
	for i := range a.queue {
		if a.queue[i].ID == id && a.queue[i].Status == "queued" {
			a.queue[i].Status = "paused"
			break
		}
	}
	a.mu.Unlock()
	a.emitQueue()
}

func (a *App) ResumeQueueTask(id string) {
	a.mu.Lock()
	for i := range a.queue {
		if a.queue[i].ID == id && a.queue[i].Status == "paused" {
			a.queue[i].Status = "queued"
			break
		}
	}
	a.mu.Unlock()
	a.emitQueue()
}

func (a *App) RetryFailedQueue() {
	a.mu.Lock()
	for i := range a.queue {
		if a.queue[i].Status == "error" || a.queue[i].Status == "cancelled" {
			a.queue[i].Status = "queued"
		}
	}
	a.mu.Unlock()
	a.emitQueue()
}

func (a *App) RunQueue() error {
	if !a.running.CompareAndSwap(false, true) {
		return fmt.Errorf("download already running")
	}
	go func() {
		defer a.running.Store(false)
		a.mu.RLock()
		var tasks []QueueTaskDTO
		for _, t := range a.queue {
			if t.Status == "queued" {
				tasks = append(tasks, t)
			}
		}
		workers := a.appSettings.QueueParallel
		a.mu.RUnlock()
		if workers < 1 {
			workers = 1
		}
		qTot := len(tasks)
		if qTot == 0 {
			return
		}
		if workers == 1 {
			for i, task := range tasks {
				a.setTaskStatus(task.ID, "running")
				a.runJob(dtoToConfig(task.Config), task.ID, i+1, qTot, true)
			}
		} else {
			var wg sync.WaitGroup
			sem := make(chan struct{}, workers)
			for i, task := range tasks {
				wg.Add(1)
				go func(idx int, t QueueTaskDTO) {
					defer wg.Done()
					sem <- struct{}{}
					defer func() { <-sem }()
					a.setTaskStatus(t.ID, "running")
					a.runJob(dtoToConfig(t.Config), t.ID, idx+1, qTot, false)
				}(i, task)
			}
			wg.Wait()
		}
		a.saveSettingsLocked()
		a.emitQueue()
		a.notifyQueueDone(tasks)
	}()
	return nil
}

func (a *App) notifyQueueDone(tasks []QueueTaskDTO) {
	a.mu.RLock()
	statusByID := make(map[string]string, len(a.queue))
	for _, t := range a.queue {
		statusByID[t.ID] = t.Status
	}
	a.mu.RUnlock()
	done, failed := 0, 0
	for _, t := range tasks {
		switch statusByID[t.ID] {
		case "completed":
			done++
		case "error", "cancelled":
			failed++
		}
	}
	a.notifyFn(i18n.T("tray.notify_title", nil), i18n.T("tray.notify_queue_done", map[string]interface{}{
		"Done": done, "Failed": failed,
	}))
}

func (a *App) ScheduleQueueRun(atUnixMillis int64) error {
	if atUnixMillis <= 0 {
		return fmt.Errorf("invalid time")
	}
	at := time.UnixMilli(atUnixMillis)
	if !at.After(time.Now()) {
		return fmt.Errorf("scheduled time must be in the future")
	}
	a.scheduleMu.Lock()
	if a.scheduleTimer != nil {
		a.scheduleTimer.Stop()
	}
	a.scheduledAt = at
	a.scheduleTimer = time.AfterFunc(time.Until(at), func() {
		a.scheduleMu.Lock()
		a.scheduledAt = time.Time{}
		a.scheduleTimer = nil
		a.scheduleMu.Unlock()
		a.emitEvent("queue:scheduled", int64(0))
		if err := a.RunQueue(); err != nil {
			a.addJournal(i18n.T("err.schedule_failed", map[string]interface{}{"Error": err.Error()}), nil)
			a.notifyFn(i18n.T("tray.notify_title", nil), i18n.T("err.schedule_failed", map[string]interface{}{"Error": err.Error()}))
		}
	})
	a.scheduleMu.Unlock()
	a.emitEvent("queue:scheduled", at.UnixMilli())
	return nil
}

func (a *App) CancelScheduledQueueRun() {
	a.scheduleMu.Lock()
	if a.scheduleTimer != nil {
		a.scheduleTimer.Stop()
		a.scheduleTimer = nil
	}
	a.scheduledAt = time.Time{}
	a.scheduleMu.Unlock()
	a.emitEvent("queue:scheduled", int64(0))
}

func (a *App) GetScheduledQueueRun() int64 {
	a.scheduleMu.Lock()
	defer a.scheduleMu.Unlock()
	if a.scheduledAt.IsZero() {
		return 0
	}
	return a.scheduledAt.UnixMilli()
}

func (a *App) GetHistory() ([]HistoryItemDTO, error) {
	h, err := core.LoadHistory()
	if err != nil {
		return nil, err
	}
	out := make([]HistoryItemDTO, 0, len(h.Items))
	for _, item := range h.Items {
		out = append(out, HistoryItemDTO{
			At:          item.At.Format(time.RFC3339),
			URL:         item.URL,
			Status:      item.Status,
			Output:      item.Output,
			Error:       item.Error,
			DurationSec: item.DurationSec,
		})
	}
	return out, nil
}

func (a *App) ClearHistory() error {
	return core.ClearHistory()
}

func (a *App) ListProfiles() ([]string, error) {
	return core.ListProfiles()
}

func (a *App) LoadProfile(name string) (ProfileDTO, error) {
	p, err := core.LoadProfile(name)
	if err != nil {
		return ProfileDTO{}, err
	}
	return ProfileDTO{
		Name:        p.Name,
		Description: p.Description,
		Config:      configToDTO(p.Config),
	}, nil
}

func (a *App) SaveProfile(p ProfileDTO) error {
	return core.SaveProfile(core.Profile{
		Name:        p.Name,
		Description: p.Description,
		Config:      dtoToConfig(p.Config),
	})
}

func (a *App) DeleteProfile(name string) error {
	return core.DeleteProfile(name)
}

func (a *App) RenameProfile(oldName, newName string) error {
	return core.RenameProfile(oldName, newName)
}

func (a *App) CheckDependencies() ([]DependencyDTO, error) {
	paths := a.dependencyPaths()
	infos, err := updater.RefreshLatestVersions(a.ctx, paths)
	a.setCachedDeps(infos)
	if err != nil {
		a.addJournal(i18n.T("dep.check_failed", nil), err)
	}
	return depsToDTO(infos), err
}

func (a *App) ResolveDependenciesLocal() []DependencyDTO {
	paths := a.dependencyPaths()
	infos := updater.ResolveAll(paths)
	a.setCachedDeps(infos)
	return depsToDTO(infos)
}

func (a *App) CheckInstallGuard(id string) InstallGuardDTO {
	paths := a.dependencyPaths()
	switch updater.DepID(id) {
	case updater.DepYtDlp:
		st := updater.CheckToolsWithPaths(paths).YtDlp
		if st.Found {
			return InstallGuardDTO{AlreadyInstalled: true, Path: st.Path}
		}
	case updater.DepFFmpeg:
		global := updater.CheckToolsWithPaths(paths).FFmpeg
		if global.Found {
			if updater.ProbeLocalFFmpeg(paths.FFmpeg).Found {
				return InstallGuardDTO{AlreadyInstalled: true, Path: global.Path}
			}
			return InstallGuardDTO{FoundElsewhere: true, Path: global.Path}
		}
	case updater.DepDeno:
		global := updater.CheckToolsWithPaths(paths).Deno
		if global.Found {
			if updater.ProbeLocalDeno(paths.Deno).Found {
				return InstallGuardDTO{AlreadyInstalled: true, Path: global.Path}
			}
			return InstallGuardDTO{FoundElsewhere: true, Path: global.Path}
		}
	}
	return InstallGuardDTO{}
}

func (a *App) InstallDependency(id string) (string, error) {
	destDir := updater.DefaultInstallDir()
	progress := func(pct int) {
		a.emitEvent("install:progress", map[string]interface{}{
			"id":  id,
			"pct": pct,
		})
	}
	var path string
	var err error
	switch updater.DepID(id) {
	case updater.DepYtDlp:
		path, err = updater.DownloadYtDlp(a.ctx, destDir, progress)
	case updater.DepFFmpeg:
		path, err = updater.DownloadFFmpeg(a.ctx, destDir, progress)
	case updater.DepDeno:
		path, err = updater.DownloadDeno(a.ctx, destDir, progress)
	default:
		return "", fmt.Errorf("unknown dependency: %s", id)
	}
	if err != nil {
		return "", err
	}
	a.saveDependencyPath(updater.DepID(id), path)
	a.addJournal(fmt.Sprintf("Installed %s at %s", id, path), nil)
	return path, nil
}

func (a *App) UpdateDependency(id string) (string, error) {
	destDir := updater.DefaultInstallDir()
	progress := func(pct int) {
		a.emitEvent("install:progress", map[string]interface{}{"id": id, "pct": pct})
	}
	paths := a.dependencyPaths()
	var path string
	var err error
	switch updater.DepID(id) {
	case updater.DepYtDlp:
		bin, resolveErr := updater.ResolveYtDlpPath(paths.YtDlp)
		if resolveErr != nil {
			return "", resolveErr
		}
		_, err = updater.UpdateYtDlp(bin)
		path = bin
	case updater.DepFFmpeg:
		path, err = updater.UpdateFFmpeg(a.ctx, destDir, progress)
	case updater.DepDeno:
		path, err = updater.UpdateDeno(a.ctx, destDir, progress)
	default:
		return "", fmt.Errorf("unknown dependency: %s", id)
	}
	if err != nil {
		return "", err
	}
	a.saveDependencyPath(updater.DepID(id), path)
	a.addJournal(fmt.Sprintf("Updated %s at %s", id, path), nil)
	return path, nil
}

func (a *App) ProbeFormats(cfg ConfigDTO) (downloader.ProbeResult, error) {
	return a.svc.Probe(a.ctx, dtoToConfig(cfg))
}

// BrowseSearch lists videos for a keyword query, or opens the listing when
// the query is a URL. cfg supplies cookies/proxy so restricted content
// resolves the same way it does for downloads.
func (a *App) BrowseSearch(cfg ConfigDTO, query string, limit int) (browse.Result, error) {
	return a.svc.BrowseSearch(a.ctx, dtoToConfig(cfg), query, limit)
}

// BrowseOpen lists the contents of a channel or playlist URL.
func (a *App) BrowseOpen(cfg ConfigDTO, target string) (browse.Result, error) {
	return a.svc.BrowseOpen(a.ctx, dtoToConfig(cfg), target)
}

func (a *App) HealthCheck() []HealthIssueDTO {
	a.depsCacheMu.RLock()
	deps := append([]updater.DependencyInfo(nil), a.cachedDeps...)
	a.depsCacheMu.RUnlock()
	if len(deps) == 0 {
		var err error
		deps, err = updater.RefreshLatestVersions(a.ctx, a.dependencyPaths())
		if err != nil {
			applog.Info("health check refresh failed", "err", err.Error())
		}
		a.setCachedDeps(deps)
	}
	mon := health.NewMonitor(
		&health.DependencyChecker{Deps: func() []updater.DependencyInfo { return deps }},
		&health.NetworkChecker{},
	)
	out := []HealthIssueDTO{}
	for _, iss := range mon.CheckAll() {
		sev := "ok"
		switch iss.Severity {
		case health.SeverityInfo:
			sev = "info"
		case health.SeverityWarning:
			sev = "warning"
		case health.SeverityCritical:
			sev = "critical"
		}
		out = append(out, HealthIssueDTO{
			ID: iss.ID, Severity: sev, Key: iss.Key, Params: iss.Params,
		})
	}
	return out
}

func (a *App) OpenFolder(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		a.mu.RLock()
		path = a.appSettings.Config.OutputPath
		a.mu.RUnlock()
	}
	if path == "" {
		return fmt.Errorf("no output path")
	}
	switch goruntime.GOOS {
	case "windows":
		return exec.Command("explorer", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}

func (a *App) PickFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: i18n.T("btn.browse", nil),
	})
}

func (a *App) PickFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: i18n.T("btn.browse", nil),
	})
}

func (a *App) PickSaveFile(defaultName string) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           i18n.T("btn.save_session", nil),
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON", Pattern: "*.json"},
		},
	})
}

func (a *App) CancelQueueTask(id string) bool {
	id = strings.TrimSpace(id)
	if id == "" {
		return false
	}
	a.runCtxMu.Lock()
	cancel := a.runCancels[id]
	a.runCtxMu.Unlock()
	if cancel != nil {
		cancel()
		return true
	}
	return downloader.CancelJob(id)
}

func (a *App) Shutdown(ctx context.Context) {
	instance.Unregister()
	a.stopTray()
	a.CancelScheduledQueueRun()
	a.StopDownload()
	downloader.CancelAll()
	a.mu.RLock()
	s := a.appSettings
	running := a.running.Load()
	a.mu.RUnlock()
	if running {
		a.sessionSnapMu.Lock()
		snap := a.lastSessionSnap
		ok := a.hasSessionSnap
		a.sessionSnapMu.Unlock()
		if ok {
			if path := strings.TrimSpace(s.SessionPath); path != "" {
				if err := core.SaveSession(path, snap); err != nil {
					applog.Info("save session failed", "err", err.Error())
				}
			}
		}
	}
	if err := settings.Save(s); err != nil {
		applog.Info("save settings failed", "err", err.Error())
	}
}

func (a *App) SaveSession(path string) error {
	a.mu.RLock()
	cfg := a.appSettings.Config
	a.mu.RUnlock()
	return core.SaveSession(path, core.Session{Config: cfg})
}

func (a *App) LoadSession(path string) (ConfigDTO, error) {
	s, err := core.LoadSession(path)
	if err != nil {
		return ConfigDTO{}, err
	}
	return configToDTO(s.Config), nil
}

func (a *App) ResumeSession(path string) (ConfigDTO, error) {
	s, err := core.LoadSession(path)
	if err != nil {
		return ConfigDTO{}, err
	}
	return configToDTO(s.ApplyResumeHints()), nil
}

type SettingsBackup struct {
	Version  int            `json:"version"`
	Settings AppSettingsDTO `json:"settings"`
	Profiles []ProfileDTO   `json:"profiles"`
}

func (a *App) ExportSettings(path string) error {
	a.mu.RLock()
	dto := settingsToDTO(a.appSettings)
	a.mu.RUnlock()
	names, err := core.ListProfiles()
	if err != nil {
		return err
	}
	profiles := make([]ProfileDTO, 0, len(names))
	for _, n := range names {
		p, err := core.LoadProfile(n)
		if err != nil {
			continue
		}
		profiles = append(profiles, ProfileDTO{Name: p.Name, Description: p.Description, Config: configToDTO(p.Config)})
	}
	backup := SettingsBackup{Version: 1, Settings: dto, Profiles: profiles}
	b, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func (a *App) ImportSettings(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var backup SettingsBackup
	if err := json.Unmarshal(b, &backup); err != nil {
		return err
	}
	next := dtoToSettings(backup.Settings)
	if err := settings.Save(next); err != nil {
		return err
	}
	a.mu.Lock()
	a.appSettings = next
	a.mu.Unlock()
	for _, p := range backup.Profiles {
		if err := core.SaveProfile(core.Profile{Name: p.Name, Description: p.Description, Config: dtoToConfig(p.Config)}); err != nil {
			applog.Info("import profile failed", "err", err.Error())
		}
	}
	return nil
}

type AppUpdateDTO struct {
	Current     string `json:"current"`
	Latest      string `json:"latest"`
	UpdateAvail bool   `json:"updateAvail"`
	URL         string `json:"url"`
}

func (a *App) CheckAppUpdate() (AppUpdateDTO, error) {
	info, err := updater.FetchLatestAppRelease(a.ctx)
	out := AppUpdateDTO{Current: version.Version, Latest: info.Version, URL: info.URL}
	if err != nil {
		return out, err
	}
	out.UpdateAvail = updater.VersionOlder(version.Version, info.Version)
	return out, nil
}

type InstanceDTO struct {
	PID       int    `json:"pid"`
	StartedAt string `json:"startedAt"`
	Busy      bool   `json:"busy"`
}

func instanceDTOs(infos []instance.Info) []InstanceDTO {
	out := make([]InstanceDTO, 0, len(infos))
	for _, info := range infos {
		out = append(out, InstanceDTO{
			PID:       info.PID,
			StartedAt: info.StartedAt.Format(time.RFC3339),
			Busy:      info.Busy,
		})
	}
	return out
}

// ListOtherInstances returns sibling VAdlp processes currently running, if any.
func (a *App) ListOtherInstances() ([]InstanceDTO, error) {
	others, err := instance.List()
	if err != nil {
		return nil, err
	}
	return instanceDTOs(others), nil
}

// CloseIdleInstances asks every sibling instance with no active download to quit
// gracefully, and reports how many were asked.
func (a *App) CloseIdleInstances() (int, error) {
	others, err := instance.List()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, info := range others {
		if info.Busy {
			continue
		}
		if err := instance.RequestQuit(info.PID); err == nil {
			n++
		}
	}
	return n, nil
}

// KillInstance force-terminates a sibling VAdlp process by PID.
func (a *App) KillInstance(pid int) error {
	return instance.Kill(pid)
}

func (a *App) GetLocales(lang string) (map[string]string, error) {
	if lang == "" {
		a.mu.RLock()
		lang = a.appSettings.Language
		a.mu.RUnlock()
	}
	if lang == "" {
		lang = "en"
	}
	return i18n.LocaleMap(lang)
}

func (a *App) GetPresets() []string {
	return append([]string(nil), core.PresetIDs...)
}

func (a *App) GetQualityPresets() []map[string]string {
	out := make([]map[string]string, 0, len(core.QualityPresets))
	for _, q := range core.QualityPresets {
		out = append(out, map[string]string{"key": q.Key, "value": q.Value})
	}
	return out
}

func (a *App) GetMergeFormats() []string {
	return append([]string(nil), core.MergeFormats...)
}

// --- helpers ---

func (a *App) dependencyPaths() updater.DependencyPaths {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return updater.DependencyPaths{
		YtDlp:  a.appSettings.YtDlpPath,
		FFmpeg: a.appSettings.FFmpegPath,
		Deno:   a.appSettings.DenoPath,
	}
}

func (a *App) setCachedDeps(infos []updater.DependencyInfo) {
	a.depsCacheMu.Lock()
	a.cachedDeps = append([]updater.DependencyInfo(nil), infos...)
	a.depsCacheMu.Unlock()
}

func (a *App) saveDependencyPath(id updater.DepID, path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}
	a.mu.Lock()
	switch id {
	case updater.DepYtDlp:
		a.appSettings.YtDlpPath = path
		a.appSettings.Config.YtDlpPath = path
	case updater.DepFFmpeg:
		a.appSettings.FFmpegPath = path
		a.appSettings.Config.FFmpegLocation = path
	case updater.DepDeno:
		a.appSettings.DenoPath = path
		a.appSettings.Config.DenoPath = path
	}
	a.mu.Unlock()
	if err := settings.Save(a.appSettings); err != nil {
		applog.Info("save settings failed", "err", err.Error())
	}
}

func (a *App) emitQueue() {
	a.mu.RLock()
	q := append([]QueueTaskDTO(nil), a.queue...)
	a.mu.RUnlock()
	a.emitEvent("queue:update", q)
}

func (a *App) setTaskStatus(id, status string) {
	a.mu.Lock()
	for i := range a.queue {
		if a.queue[i].ID == id {
			a.queue[i].Status = status
			break
		}
	}
	a.mu.Unlock()
	a.emitQueue()
}

func (a *App) addJournal(msg string, err error) {
	entry := msg
	if err != nil {
		entry = msg + ": " + err.Error()
	}
	a.mu.Lock()
	a.journal = append(a.journal, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), entry))
	a.mu.Unlock()
	a.emitEvent("journal:add", entry)
}

func (a *App) saveSettingsLocked() {
	a.mu.RLock()
	s := a.appSettings
	a.mu.RUnlock()
	if err := settings.Save(s); err != nil {
		applog.Info("save settings failed", "err", err.Error())
	}
}

func (a *App) emitProgress(p DownloadProgressDTO) {
	a.emitEvent("download:progress", p)
}

func (a *App) emitLog(line string) {
	a.emitEvent("download:log", line)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func configToDTO(c core.Config) ConfigDTO {
	return ConfigDTO{
		URL: c.URL, Quality: c.Quality, Format: c.Format, AudioOnly: c.AudioOnly,
		AudioFormat: c.AudioFormat, OutputPath: c.OutputPath, OutputTemplate: c.OutputTemplate,
		UseCookiesFile: c.UseCookiesFile, CookiesFile: c.CookiesFile,
		UseCookiesBrowser: c.UseCookiesBrowser, CookiesBrowser: c.CookiesBrowser,
		Proxy: c.Proxy, RateLimit: c.RateLimit, PlaylistReverse: c.PlaylistReverse,
		Continue: c.Continue, NoPart: c.NoPart, PlaylistStart: c.PlaylistStart,
		PlaylistEnd: c.PlaylistEnd, MaxDownloads: c.MaxDownloads, DownloadArchive: c.DownloadArchive,
		NoPlaylist: c.NoPlaylist, FlatPlaylist: c.FlatPlaylist, WriteSubs: c.WriteSubs,
		WriteAutoSub: c.WriteAutoSub, EmbedSubs: c.EmbedSubs, SubLangs: c.SubLangs,
		WriteThumbnail: c.WriteThumbnail, EmbedThumbnail: c.EmbedThumbnail,
		EmbedMetadata: c.EmbedMetadata, EmbedChapters: c.EmbedChapters,
		Retries: c.Retries, FragmentRetries: c.FragmentRetries,
		ConcurrentFragments: c.ConcurrentFragments, SocketTimeout: c.SocketTimeout,
		NoWarnings: c.NoWarnings, Verbose: c.Verbose, Quiet: c.Quiet,
		WriteInfoJSON: c.WriteInfoJSON, LoadInfoJSON: c.LoadInfoJSON,
		WindowsFilenames: c.WindowsFilenames, NoMtime: c.NoMtime,
		AbortOnError: c.AbortOnError, IgnoreErrors: c.IgnoreErrors,
		ExtraArgs: c.ExtraArgs, FFmpegLocation: c.FFmpegLocation,
		Username: c.Username, Password: c.Password,
		SponsorBlockRemove: c.SponsorBlockRemove, BatchURLs: c.BatchURLs,
		YtDlpPath: c.YtDlpPath, DenoPath: c.DenoPath,
	}
}

func dtoToConfig(d ConfigDTO) core.Config {
	return core.Config{
		URL: d.URL, Quality: d.Quality, Format: d.Format, AudioOnly: d.AudioOnly,
		AudioFormat: d.AudioFormat, OutputPath: d.OutputPath, OutputTemplate: d.OutputTemplate,
		UseCookiesFile: d.UseCookiesFile, CookiesFile: d.CookiesFile,
		UseCookiesBrowser: d.UseCookiesBrowser, CookiesBrowser: d.CookiesBrowser,
		Proxy: d.Proxy, RateLimit: d.RateLimit, PlaylistReverse: d.PlaylistReverse,
		Continue: d.Continue, NoPart: d.NoPart, PlaylistStart: d.PlaylistStart,
		PlaylistEnd: d.PlaylistEnd, MaxDownloads: d.MaxDownloads, DownloadArchive: d.DownloadArchive,
		NoPlaylist: d.NoPlaylist, FlatPlaylist: d.FlatPlaylist, WriteSubs: d.WriteSubs,
		WriteAutoSub: d.WriteAutoSub, EmbedSubs: d.EmbedSubs, SubLangs: d.SubLangs,
		WriteThumbnail: d.WriteThumbnail, EmbedThumbnail: d.EmbedThumbnail,
		EmbedMetadata: d.EmbedMetadata, EmbedChapters: d.EmbedChapters,
		Retries: d.Retries, FragmentRetries: d.FragmentRetries,
		ConcurrentFragments: d.ConcurrentFragments, SocketTimeout: d.SocketTimeout,
		NoWarnings: d.NoWarnings, Verbose: d.Verbose, Quiet: d.Quiet,
		WriteInfoJSON: d.WriteInfoJSON, LoadInfoJSON: d.LoadInfoJSON,
		WindowsFilenames: d.WindowsFilenames, NoMtime: d.NoMtime,
		AbortOnError: d.AbortOnError, IgnoreErrors: d.IgnoreErrors,
		ExtraArgs: d.ExtraArgs, FFmpegLocation: d.FFmpegLocation,
		Username: d.Username, Password: d.Password,
		SponsorBlockRemove: d.SponsorBlockRemove, BatchURLs: d.BatchURLs,
		YtDlpPath: d.YtDlpPath, DenoPath: d.DenoPath,
	}
}

func settingsToDTO(s settings.App) AppSettingsDTO {
	cfg := s.Config
	cfg.YtDlpPath = firstNonEmpty(s.YtDlpPath, cfg.YtDlpPath)
	cfg.DenoPath = firstNonEmpty(s.DenoPath, cfg.DenoPath)
	return AppSettingsDTO{
		Config:              configToDTO(cfg),
		FFmpegPath:          s.FFmpegPath,
		SessionPath:         s.SessionPath,
		QueueParallel:       s.QueueParallel,
		Language:            s.Language,
		YtDlpPath:           s.YtDlpPath,
		DenoPath:            s.DenoPath,
		LastProfile:         s.LastProfile,
		DebugLog:            s.DebugLog,
		ActivityPanelOpen:   s.ActivityPanelOpen,
		UIScale:             s.UIScale,
		Theme:               firstNonEmpty(s.Theme, "auto"),
		WindowWidth:         s.WindowWidth,
		WindowHeight:        s.WindowHeight,
		ActivityPanelOffset: s.ActivityPanelOffset,
		BrowserEnabled:      s.BrowserEnabled,
	}
}

func dtoToSettings(d AppSettingsDTO) settings.App {
	s := settings.Default()
	s.Config = dtoToConfig(d.Config)
	s.FFmpegPath = d.FFmpegPath
	s.SessionPath = d.SessionPath
	s.QueueParallel = d.QueueParallel
	s.Language = d.Language
	s.YtDlpPath = d.YtDlpPath
	s.DenoPath = d.DenoPath
	s.LastProfile = d.LastProfile
	s.DebugLog = d.DebugLog
	s.ActivityPanelOpen = d.ActivityPanelOpen
	s.UIScale = d.UIScale
	s.Theme = d.Theme
	s.WindowWidth = d.WindowWidth
	s.WindowHeight = d.WindowHeight
	s.BrowserEnabled = d.BrowserEnabled
	if d.ActivityPanelOffset > 0.05 && d.ActivityPanelOffset < 0.95 {
		s.ActivityPanelOffset = d.ActivityPanelOffset
	}
	if s.QueueParallel < 1 {
		s.QueueParallel = 1
	}
	return s
}

func depLevelString(l updater.DepLevel) string {
	switch l {
	case updater.DepRequired:
		return "required"
	case updater.DepRecommended:
		return "recommended"
	default:
		return "optional"
	}
}

func depsToDTO(infos []updater.DependencyInfo) []DependencyDTO {
	out := make([]DependencyDTO, 0, len(infos))
	for _, d := range infos {
		out = append(out, DependencyDTO{
			ID: string(d.ID), Level: depLevelString(d.Level), Status: string(d.Status),
			Path: d.Path, Version: d.Version, LatestVer: d.LatestVer,
			UpdateAvail: d.UpdateAvail, Source: string(d.Source), Error: d.Error,
		})
	}
	return out
}

func (a *App) runJob(current core.Config, taskID string, qIdx, qTot int, focusUI bool) bool {
	plCur, plTot := 0, 0
	filePct := 0.0
	var localLogs []string
	emitUI := focusUI || taskID != ""

	var persistMu sync.Mutex
	lastPersistWrite := time.Time{}
	updateSessionSnap := func() {
		snap := core.Session{
			Config:          current,
			PlaylistCurrent: plCur,
			PlaylistTotal:   plTot,
			QueueDone:       maxInt(0, qIdx-1),
			QueueTotal:      qTot,
		}
		a.sessionSnapMu.Lock()
		a.lastSessionSnap = snap
		a.hasSessionSnap = true
		a.sessionSnapMu.Unlock()

		a.mu.RLock()
		path := strings.TrimSpace(a.appSettings.SessionPath)
		a.mu.RUnlock()
		if path == "" {
			return
		}
		persistMu.Lock()
		defer persistMu.Unlock()
		if time.Since(lastPersistWrite) < 2*time.Second {
			return
		}
		lastPersistWrite = time.Now()
		if err := core.SaveSession(path, snap); err != nil {
			a.addJournal(i18n.T("err.save_session", nil), err)
		}
	}

	progressDTO := func(status, phase string, file, overall float64, speed, eta string) DownloadProgressDTO {
		return DownloadProgressDTO{
			Status: status, Phase: phase,
			FilePct: file, OverallPct: overall,
			Speed: speed, ETA: eta,
			PlCurrent: plCur, PlTotal: plTot,
			QueueIdx: qIdx, QueueTotal: qTot,
			TaskID: taskID,
		}
	}

	if emitUI {
		a.emitProgress(progressDTO("running", "idle", 0, 0, "", ""))
	}

	// computeOverall blends this task's own progress (file%, or its internal
	// playlist position if it has one) into its slot within the whole queue run,
	// so the bar always spans the full run instead of jumping scales whenever the
	// active task happens to be a playlist.
	computeOverall := func() float64 {
		taskFrac := filePct / 100.0
		if plTot > 0 && plCur > 0 {
			taskFrac = (float64(plCur-1) + filePct/100.0) / float64(plTot)
		}
		if qTot > 0 && qIdx > 0 {
			return (float64(qIdx-1) + taskFrac) / float64(qTot) * 100.0
		}
		return taskFrac * 100.0
	}

	jobID := taskID
	if jobID == "" {
		jobID = "main"
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.runCtxMu.Lock()
	a.runCancels[jobID] = cancel
	a.runCtxMu.Unlock()
	defer func() {
		cancel()
		a.runCtxMu.Lock()
		delete(a.runCancels, jobID)
		a.runCtxMu.Unlock()
	}()

	result, err := a.svc.Download(ctx, current, jobID, func(ev downloader.Event) {
		switch ev.Type {
		case downloader.EventLog:
			line := ev.LogLine
			if emitUI && ev.Stage != downloader.StageUnknown {
				a.emitProgress(progressDTO("running", string(ev.Stage), filePct, computeOverall(), "", ""))
			}
			if !downloader.IsProgressLine(line) {
				localLogs = append(localLogs, line)
				if len(localLogs) > 450 {
					localLogs = localLogs[len(localLogs)-450:]
				}
				if focusUI {
					a.emitLog(strings.Join(localLogs, "\n"))
				}
			}
		case downloader.EventProgress:
			filePct = ev.Progress
			if emitUI {
				a.emitProgress(progressDTO("running", "downloading", ev.Progress, computeOverall(), ev.Speed, ev.ETA))
			}
			updateSessionSnap()
		case downloader.EventPlaylist:
			plCur = ev.PlaylistCurrent
			plTot = ev.PlaylistTotal
			if emitUI {
				a.emitProgress(progressDTO("running", "downloading", filePct, computeOverall(), "", ""))
			}
			updateSessionSnap()
		}
	})

	if err != nil {
		st := "error"
		msg := err.Error()
		if downloader.IsCancelled(err) {
			st = "cancelled"
			msg = "cancelled"
		}
		if emitUI {
			a.emitProgress(progressDTO(st, "idle", filePct, computeOverall(), "", ""))
		}
		if focusUI {
			localLogs = append(localLogs, strings.ToUpper(msg))
			a.emitLog(strings.Join(localLogs, "\n"))
		}
		if err := core.AppendHistory(core.HistoryItem{
			URL: firstNonEmpty(current.URL, current.BatchURLs), Status: st,
			Output: current.OutputPath, Error: msg, DurationSec: result.DurationSec,
		}); err != nil {
			applog.Info("append history failed", "err", err.Error())
		}
		if taskID != "" {
			a.setTaskStatus(taskID, st)
		} else if st == "error" {
			a.notifyFn(i18n.T("tray.notify_title", nil), i18n.T("tray.notify_error", map[string]interface{}{"URL": current.URL}))
		}
		return !downloader.IsCancelled(err)
	}

	if emitUI {
		a.emitProgress(progressDTO("completed", "idle", 100, 100, "", ""))
	}
	if err := core.AppendHistory(core.HistoryItem{
		URL: firstNonEmpty(current.URL, current.BatchURLs), Status: "completed",
		Output: current.OutputPath, DurationSec: result.DurationSec,
	}); err != nil {
		applog.Info("append history failed", "err", err.Error())
	}
	if taskID != "" {
		a.setTaskStatus(taskID, "completed")
	} else {
		a.notifyFn(i18n.T("tray.notify_title", nil), i18n.T("tray.notify_done", map[string]interface{}{"URL": current.URL}))
	}
	return true
}
