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

var progressRegex = regexp.MustCompile(`(?i)(\d{1,3}(?:\.\d+)?)%`)

type EventType string

const (
	EventLog      EventType = "log"
	EventProgress EventType = "progress"
)

type Event struct {
	Type     EventType
	LogLine  string
	Progress float64
}

func resolveYTDLPBinary() string {
	binName := "yt-dlp"
	if runtime.GOOS == "windows" {
		binName = "yt-dlp.exe"
	}

	localBin := filepath.Join(".", "bin", binName)
	if _, err := os.Stat(localBin); err == nil {
		return localBin
	}

	return "yt-dlp"
}

func Run(cfg core.Config, onEvent func(Event)) (string, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return "", errors.New("URL is required")
	}

	args := core.BuildCommand(cfg)
	binary := resolveYTDLPBinary()
	cmd := exec.Command(binary, args...)

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
			onEvent(Event{Type: EventLog, LogLine: line})
		}

		match := progressRegex.FindStringSubmatch(line)
		if len(match) > 1 {
			percent, parseErr := strconv.ParseFloat(match[1], 64)
			if parseErr == nil && onEvent != nil {
				onEvent(Event{Type: EventProgress, Progress: percent})
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