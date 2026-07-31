package downloader

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"vadlp/internal/core"
	"vadlp/internal/executil"
	"vadlp/internal/updater"
)

var (
	// progressRegex matches yt-dlp download progress lines only, so a stray
	// percentage in a filename or title cannot produce a spurious event.
	progressRegex = regexp.MustCompile(`(?i)\[download\]\s+(\d{1,3}(?:\.\d+)?)%`)
	// playlistRegex requires the "[download] Downloading" prefix. The old
	// pattern had an empty alternative in the prefix group, which made every
	// "N of M" substring match — including filenames like "Part 1 of 3.mp4"
	// in Destination lines — and produced bogus playlist events.
	playlistRegex = regexp.MustCompile(`(?i)\[download\]\s+Downloading\s+(?:video\s+|item\s+)?(\d+)\s+of\s+(\d+)`)
	// ~ prefix marks approximate speeds (yt-dlp: "at ~1.23MiB/s").
	speedRegex = regexp.MustCompile(`(?i)at\s+~?([\d.]+\s*(?:[KMGT]?i?B|B)(?:/s)?)`)
	etaRegex   = regexp.MustCompile(`(?i)ETA\s+(\d{1,2}:\d{2}(?::\d{2})?)`)
)

// IsProgressLine reports whether line is a yt-dlp download progress line
// (e.g. "[download]  42.5% of 10.00MiB at 1.2MiB/s ETA 00:10"). Progress
// lines are excluded from the UI log because they would spam it.
func IsProgressLine(line string) bool {
	return progressRegex.MatchString(line)
}

type EventType string

const (
	EventLog      EventType = "log"
	EventProgress EventType = "progress"
	EventPlaylist EventType = "playlist"
)

type Stage string

const (
	StageUnknown     Stage = ""
	StageExtracting  Stage = "EXTRACTING"
	StageDownloading Stage = "DOWNLOADING"
	StagePostProcess Stage = "POST-PROCESSING"
)

type Event struct {
	Type            EventType
	LogLine         string
	Progress        float64
	PlaylistCurrent int
	PlaylistTotal   int
	Stage           Stage
	Speed           string
	ETA             string
}

func ResolveBinary(customPath string) (string, error) {
	return updater.ResolveYtDlpPath(customPath)
}

func detectStage(line string) Stage {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "merging formats"),
		strings.Contains(lower, "embedding"),
		strings.Contains(lower, "post-process"),
		strings.Contains(lower, "ffmpeg"),
		strings.Contains(lower, "fixup"),
		strings.Contains(lower, "extractaudio"):
		return StagePostProcess
	case strings.Contains(lower, "[download]") && strings.Contains(lower, "%"):
		return StageDownloading
	case strings.Contains(lower, "[download]") && strings.Contains(lower, "destination"):
		return StageDownloading
	case strings.Contains(lower, "extracting url"),
		strings.Contains(lower, "download webpage"),
		strings.Contains(lower, "downloading api"),
		strings.Contains(lower, "extracting info"):
		return StageExtracting
	case strings.Contains(lower, "[info]") || strings.Contains(lower, "[youtube]") || strings.Contains(lower, "[extractor]"):
		if strings.Contains(lower, "downloading") && !strings.Contains(lower, "[download]") {
			return StageExtracting
		}
	}
	return StageUnknown
}

func Run(cfg core.Config, jobID string, onEvent func(Event)) (string, error) {
	return RunCtx(context.Background(), cfg, jobID, onEvent)
}

func RunCtx(ctx context.Context, cfg core.Config, jobID string, onEvent func(Event)) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if jobID == "" {
		jobID = "main"
	}
	if strings.TrimSpace(cfg.LoadInfoJSON) == "" && len(core.URLsFromConfig(cfg)) == 0 {
		return "", errors.New("URL or load-info-json is required")
	}

	args := core.BuildCommand(cfg)
	binary, binErr := ResolveBinary(cfg.YtDlpPath)
	if binErr != nil {
		return "", binErr
	}
	cmd := executil.Command(binary, args...)

	if cfg.OutputPath != "" {
		cmd.Dir = cfg.OutputPath
	}

	return runCommand(ctx, jobID, cmd, onEvent)
}

// runCommand streams stdout and stderr of cmd concurrently and blocks until
// the process exits, the context is cancelled or a pipe fails. stdout and
// stderr are drained by separate pump goroutines: reading them serially
// (e.g. via io.MultiReader) deadlocks whenever the child fills the stderr
// pipe buffer while stdout is idle — the child blocks writing stderr, the
// parent blocks reading stdout, and neither ctx cancellation nor CancelJob
// can reach the stuck process (yt-dlp hits this on stderr-heavy output,
// e.g. verbose dumps or tracebacks). A dedicated goroutine additionally
// kills the subprocess as soon as ctx is done, because the main loop may be
// blocked inside a pipe Read or cmd.Wait where ctx.Done() is invisible.
func runCommand(ctx context.Context, jobID string, cmd *exec.Cmd, onEvent func(Event)) (string, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start yt-dlp: %w", err)
	}

	registerJob(jobID, func() error {
		if cmd.Process != nil {
			return cmd.Process.Kill()
		}
		return nil
	})
	defer unregisterJob(jobID)

	stopKiller := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
		case <-stopKiller:
		}
	}()
	defer close(stopKiller)

	lines := make(chan string, 256)
	var readWG sync.WaitGroup
	readWG.Add(2)
	var readErrMu sync.Mutex
	var readErr error

	pump := func(r io.Reader) {
		defer readWG.Done()
		scanner := bufio.NewScanner(r)
		scanner.Split(splitLines)
		scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
		for scanner.Scan() {
			select {
			case lines <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
		if err := scanner.Err(); err != nil {
			readErrMu.Lock()
			if readErr == nil {
				readErr = err
			}
			readErrMu.Unlock()
		}
	}
	go pump(stdout)
	go pump(stderr)
	go func() {
		readWG.Wait()
		close(lines)
	}()

	// terminate kills the subprocess and reaps it. Killing alone leaves the
	// child unreaped (a zombie on Unix, a leaked handle on Windows) for the
	// lifetime of the app, and cancellations accumulate over a long session.
	// The pumps must finish before cmd.Wait, which closes the pipes they read
	// from; draining lines both unblocks a pump stuck on send and ends once
	// both pumps have hit EOF on the dead process and closed the channel.
	terminate := func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		for range lines { //nolint:revive // drain to release the pumps
		}
		_ = cmd.Wait()
	}

	var logs strings.Builder
loop:
	for {
		select {
		case <-ctx.Done():
			terminate()
			return logs.String(), ErrCancelled
		case line, ok := <-lines:
			if !ok {
				break loop
			}
			if jobCancelled(jobID) {
				terminate()
				return logs.String(), ErrCancelled
			}
			logs.WriteString(line)
			logs.WriteString("\n")

			if onEvent != nil {
				onEvent(Event{Type: EventLog, LogLine: line, Stage: detectStage(line)})
			}

			match := progressRegex.FindStringSubmatch(line)
			if len(match) > 1 {
				percent, parseErr := strconv.ParseFloat(match[1], 64)
				if parseErr == nil && onEvent != nil {
					ev := Event{Type: EventProgress, Progress: percent}
					if sm := speedRegex.FindStringSubmatch(line); len(sm) > 1 {
						ev.Speed = strings.TrimSpace(sm[1])
					}
					if em := etaRegex.FindStringSubmatch(line); len(em) > 1 {
						ev.ETA = em[1]
					}
					onEvent(ev)
				}
			}

			if pm := playlistRegex.FindStringSubmatch(line); len(pm) == 3 {
				cur, e1 := strconv.Atoi(pm[1])
				tot, e2 := strconv.Atoi(pm[2])
				if e1 == nil && e2 == nil && tot > 0 && onEvent != nil {
					onEvent(Event{
						Type:            EventPlaylist,
						PlaylistCurrent: cur,
						PlaylistTotal:   tot,
						LogLine:         line,
					})
				}
			}
		}
	}

	readErrMu.Lock()
	scanErr := readErr
	readErrMu.Unlock()
	if scanErr != nil {
		return logs.String(), scanErr
	}
	if waitErr := cmd.Wait(); waitErr != nil {
		if jobCancelled(jobID) {
			return logs.String(), ErrCancelled
		}
		return logs.String(), waitErr
	}

	if onEvent != nil {
		onEvent(Event{Type: EventProgress, Progress: 100})
	}
	return logs.String(), nil
}

// splitLines splits on \r, \n or \r\n, and emits a final unterminated token
// when the stream ends (same semantics the old scanner split used).
func splitLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
