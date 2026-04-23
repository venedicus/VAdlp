package downloader

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"ytgui/internal/core"
)

var (
	progressRegex = regexp.MustCompile(`(?i)(\d{1,3}(?:\.\d+)?)%`)
	playlistRegex = regexp.MustCompile(`(?i)(?:\[download\][^\d]*)?(?:Downloading\s+(?:video\s+|item\s+)?|)(\d+)\s+of\s+(\d+)`)
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
}

// ResolveBinary finds yt-dlp without relying on the shell: next to the executable, then CWD ./bin, then PATH.
// GUI launches (double-click on Windows) often have a different working directory and a shorter PATH than your terminal,
// which is why "it works in the terminal only" unless we search explicitly.
func ResolveBinary() (string, error) {
	binName := "yt-dlp"
	if runtime.GOOS == "windows" {
		binName = "yt-dlp.exe"
	}

	try := func(path string) (string, bool) {
		if path == "" {
			return "", false
		}
		path = filepath.Clean(path)
		st, err := os.Stat(path)
		if err != nil || st.IsDir() {
			return "", false
		}
		return path, true
	}

	seen := map[string]struct{}{}
	add := func(list *[]string, p string) {
		if p == "" {
			return
		}
		p = filepath.Clean(p)
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		*list = append(*list, p)
	}

	var candidates []string

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		add(&candidates, filepath.Join(exeDir, "bin", binName))
		add(&candidates, filepath.Join(exeDir, binName))
		// go install / layout: .../bin/ytgui.exe with yt-dlp.exe alongside
		add(&candidates, filepath.Join(exeDir, "..", "bin", binName))
	}
	if wd, err := os.Getwd(); err == nil {
		add(&candidates, filepath.Join(wd, "bin", binName))
	}

	for _, p := range candidates {
		if abs, ok := try(p); ok {
			return abs, nil
		}
	}

	if p, err := exec.LookPath(binName); err == nil {
		return p, nil
	}
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("yt-dlp"); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("yt-dlp not found (install from https://github.com/yt-dlp/yt-dlp or place %s next to this app, or under bin\\ beside it)", binName)
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

func Run(cfg core.Config, onEvent func(Event)) (string, error) {
	if strings.TrimSpace(cfg.LoadInfoJSON) == "" && strings.TrimSpace(cfg.URL) == "" {
		return "", errors.New("URL or load-info-json is required")
	}

	args := core.BuildCommand(cfg)
	binary, binErr := ResolveBinary()
	if binErr != nil {
		return "", binErr
	}
	cmd := exec.Command(binary, args...)

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
				onEvent(Event{Type: EventProgress, Progress: percent})
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
		return logs.String(), waitErr
	}

	if onEvent != nil {
		onEvent(Event{Type: EventProgress, Progress: 100})
	}
	return logs.String(), nil
}
