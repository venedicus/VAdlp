package downloader

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"vadlp/internal/core"
	"vadlp/internal/executil"
	"vadlp/internal/updater"
)

var (
	progressRegex = regexp.MustCompile(`(?i)(\d{1,3}(?:\.\d+)?)%`)
	playlistRegex = regexp.MustCompile(`(?i)(?:\[download\][^\d]*)?(?:Downloading\s+(?:video\s+|item\s+)?|)(\d+)\s+of\s+(\d+)`)
	speedRegex    = regexp.MustCompile(`(?i)at\s+([\d.]+\s*(?:[KMGT]?i?B|B)(?:/s)?)`)
	etaRegex      = regexp.MustCompile(`(?i)ETA\s+(\d{1,2}:\d{2}(?::\d{2})?)`)
)

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

	reader := io.MultiReader(stdout, stderr)
	scanner := bufio.NewScanner(reader)

	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
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
	})
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	var logs strings.Builder

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			return logs.String(), context.Cause(ctx)
		default:
		}
		if jobCancelled(jobID) {
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			return logs.String(), ErrCancelled
		}
		line := scanner.Text()
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

	if scanErr := scanner.Err(); scanErr != nil {
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
