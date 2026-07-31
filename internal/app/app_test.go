package app

import (
	"context"
	"reflect"
	"testing"
	"time"

	"vadlp/internal/core"
	"vadlp/internal/settings"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	app := New()
	app.emitEvent = func(name string, data ...interface{}) {}
	app.notifyFn = func(title, message string) {}
	return app
}

func TestQueueLifecycle(t *testing.T) {
	a := newTestApp(t)
	cfg := ConfigDTO{URL: "https://example.com/a"}

	task, err := a.AddToQueue(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != "queued" || task.Name != "https://example.com/a" {
		t.Fatalf("task: %+v", task)
	}
	if _, err := a.AddToQueue(ConfigDTO{}); err == nil {
		t.Fatal("expected error for URL-less task")
	}

	if err := a.UpdateQueueTask(task.ID, ConfigDTO{URL: "https://example.com/b"}); err != nil {
		t.Fatal(err)
	}
	if err := a.UpdateQueueTask("missing", cfg); err == nil {
		t.Fatal("expected error for unknown task")
	}

	a.PauseQueueTask(task.ID)
	if st := queueStatus(a, task.ID); st != "paused" {
		t.Fatalf("status %q", st)
	}
	a.ResumeQueueTask(task.ID)
	if st := queueStatus(a, task.ID); st != "queued" {
		t.Fatalf("status %q", st)
	}

	// Reordering with an unknown id in the list must not drop tasks.
	a.ReorderQueue([]string{"does-not-exist"})
	state := a.GetState()
	if len(state.Queue) != 1 || state.Queue[0].ID != task.ID {
		t.Fatalf("queue after reorder: %+v", state.Queue)
	}

	a.RemoveFromQueue(task.ID)
	if len(a.GetState().Queue) != 0 {
		t.Fatal("expected empty queue")
	}
}

func TestQueueRetryAndClear(t *testing.T) {
	a := newTestApp(t)
	t1, _ := a.AddToQueue(ConfigDTO{URL: "u1"})
	t2, _ := a.AddToQueue(ConfigDTO{URL: "u2"})
	a.mu.Lock()
	for i := range a.queue {
		switch a.queue[i].ID {
		case t1.ID:
			a.queue[i].Status = "error"
		case t2.ID:
			a.queue[i].Status = "cancelled"
		}
	}
	a.mu.Unlock()

	a.RetryFailedQueue()
	if queueStatus(a, t1.ID) != "queued" || queueStatus(a, t2.ID) != "queued" {
		t.Fatal("retry should reset error/cancelled tasks")
	}

	a.ClearQueue()
	if len(a.GetState().Queue) != 0 {
		t.Fatal("expected empty queue")
	}
}

func TestRunDownloadRejectsConcurrent(t *testing.T) {
	a := newTestApp(t)
	a.running.Store(true)
	if err := a.RunDownload(ConfigDTO{URL: "https://example.com/a"}); err == nil {
		t.Fatal("expected error while busy")
	}
	a.running.Store(false)
}

func TestRunQueueRejectsConcurrent(t *testing.T) {
	a := newTestApp(t)
	a.running.Store(true)
	if err := a.RunQueue(); err == nil {
		t.Fatal("expected error while busy")
	}
	a.running.Store(false)
}

func TestRunJobValidationErrorMarksTaskAndHistory(t *testing.T) {
	a := newTestApp(t)
	dir := t.TempDir()
	core.SetHistoryDir(dir)
	defer core.SetHistoryDir("")
	if err := core.ClearHistory(); err != nil {
		t.Fatal(err)
	}

	task := QueueTaskDTO{ID: "t1", Name: "x", Status: "queued", Config: ConfigDTO{}}
	a.mu.Lock()
	a.queue = append(a.queue, task)
	a.mu.Unlock()

	cfg := core.DefaultConfig()
	cfg.URL = "https://example.com/a"
	cfg.Retries = 200 // fails validation before any process starts

	a.runJob(cfg, task.ID, 1, 1, false)
	if st := queueStatus(a, task.ID); st != "error" {
		t.Fatalf("task status %q, want error", st)
	}

	h, err := core.LoadHistory()
	if err != nil {
		t.Fatal(err)
	}
	if len(h.Items) != 1 || h.Items[0].Status != "error" || h.Items[0].URL != cfg.URL {
		t.Fatalf("history: %+v", h.Items)
	}
}

func TestStopDownloadCancelsAllRuns(t *testing.T) {
	a := newTestApp(t)
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	a.runCtxMu.Lock()
	a.runCancels["t1"] = cancel1
	a.runCancels["t2"] = cancel2
	a.runCtxMu.Unlock()

	a.StopDownload()
	if ctx1.Err() == nil || ctx2.Err() == nil {
		t.Fatal("StopDownload should cancel every running job")
	}
}

func TestCancelQueueTaskOnlyCancelsItsOwnRun(t *testing.T) {
	a := newTestApp(t)
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	a.runCtxMu.Lock()
	a.runCancels["t1"] = cancel1
	a.runCancels["t2"] = cancel2
	a.runCtxMu.Unlock()

	if !a.CancelQueueTask("t1") {
		t.Fatal("CancelQueueTask should report the task was handled")
	}
	if ctx1.Err() == nil {
		t.Fatal("target task context should be cancelled")
	}
	if ctx2.Err() != nil {
		t.Fatal("other task context must not be cancelled")
	}
}

func TestScheduleQueueRun(t *testing.T) {
	a := newTestApp(t)
	if err := a.ScheduleQueueRun(0); err == nil {
		t.Fatal("expected error for invalid time")
	}
	if err := a.ScheduleQueueRun(time.Now().Add(-time.Second).UnixMilli()); err == nil {
		t.Fatal("expected error for past time")
	}

	at := time.Now().Add(100 * time.Millisecond).UnixMilli()
	if err := a.ScheduleQueueRun(at); err != nil {
		t.Fatal(err)
	}
	if a.GetScheduledQueueRun() == 0 {
		t.Fatal("scheduled run should be pending")
	}
	deadline := time.Now().Add(5 * time.Second)
	for a.GetScheduledQueueRun() != 0 {
		if time.Now().After(deadline) {
			t.Fatal("scheduled run never fired")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if a.running.Load() {
		t.Fatal("empty queue run should not leave running=true")
	}
}

func TestCancelScheduledQueueRun(t *testing.T) {
	a := newTestApp(t)
	if err := a.ScheduleQueueRun(time.Now().Add(time.Minute).UnixMilli()); err != nil {
		t.Fatal(err)
	}
	a.CancelScheduledQueueRun()
	if a.GetScheduledQueueRun() != 0 {
		t.Fatal("scheduled run should be cancelled")
	}
}

func TestConfigDTOConversionRoundTrip(t *testing.T) {
	d := ConfigDTO{
		URL: "https://example.com/v", Quality: "best", Format: "mp4",
		AudioOnly: true, AudioFormat: "mp3", OutputPath: `C:\Videos`,
		UseCookiesBrowser: true, CookiesBrowser: "chrome", Proxy: "socks5://localhost:1080",
		RateLimit: "1M", PlaylistStart: 2, PlaylistEnd: 10, MaxDownloads: 5,
		Retries: 7, ConcurrentFragments: 4, SocketTimeout: 30,
		Verbose: true, ExtraArgs: "--cookies-from-browser edge",
		FFmpegLocation: "C:\\tools\\ffmpeg", Username: "u", Password: "p",
		SponsorBlockRemove: true, BatchURLs: "a\nb\n", YtDlpPath: "C:\\tools\\yt-dlp.exe",
	}
	got := configToDTO(dtoToConfig(d))
	if !reflect.DeepEqual(d, got) {
		t.Fatalf("round trip mismatch:\n got %+v\nwant %+v", got, d)
	}
}

func TestSettingsDTOConversionRoundTrip(t *testing.T) {
	s := settings.Default()
	s.Config.URL = "https://example.com/s"
	s.YtDlpPath = "C:\\tools\\yt-dlp.exe"
	s.Config.YtDlpPath = ""
	s.Language = "ru"
	s.Theme = "dark"
	s.QueueParallel = 3
	s.ActivityPanelOffset = 0.7
	s.WindowWidth = 1400
	s.WindowHeight = 900

	d := settingsToDTO(s)
	got := dtoToSettings(d)
	if got.Language != "ru" || got.Theme != "dark" || got.QueueParallel != 3 {
		t.Fatalf("settings mismatch: %+v", got)
	}
	if got.ActivityPanelOffset != 0.7 {
		t.Fatalf("offset %v", got.ActivityPanelOffset)
	}
	if got.WindowWidth != 1400 || got.WindowHeight != 900 {
		t.Fatalf("window size %v x %v", got.WindowWidth, got.WindowHeight)
	}
	if got.YtDlpPath != s.YtDlpPath {
		t.Fatalf("yt-dlp path %q, want %q", got.YtDlpPath, s.YtDlpPath)
	}
}

func TestDTOToSettingsSanitizesInvalid(t *testing.T) {
	d := AppSettingsDTO{QueueParallel: 0, ActivityPanelOffset: 0.999, UIScale: 3}
	got := dtoToSettings(d)
	if got.QueueParallel != 1 {
		t.Fatalf("queueParallel %d, want 1", got.QueueParallel)
	}
	if got.ActivityPanelOffset == 0.999 {
		t.Fatal("out-of-range offset should not be applied")
	}
}

func TestGetStateSnapshot(t *testing.T) {
	a := newTestApp(t)
	task, _ := a.AddToQueue(ConfigDTO{URL: "https://example.com/x"})
	state := a.GetState()
	if len(state.Queue) != 1 || state.Queue[0].ID != task.ID {
		t.Fatalf("state queue: %+v", state.Queue)
	}
	if len(state.Journal) != 0 {
		t.Fatalf("journal: %v", state.Journal)
	}
}

func queueStatus(a *App, id string) string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, t := range a.queue {
		if t.ID == id {
			return t.Status
		}
	}
	return ""
}
