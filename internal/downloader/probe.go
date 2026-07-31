package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"vadlp/internal/core"
	"vadlp/internal/executil"
	"vadlp/internal/jsonutil"
)

const probeTimeout = 30 * time.Second

type Format struct {
	ID         string
	Ext        string
	Resolution string
	FPS        int
	Vcodec     string
	Acodec     string
	Filesize   int64
	TBR        float64
	Note       string
}

type MediaEntry struct {
	Title     string
	ID        string
	URL       string
	Duration  string
	Uploader  string
	Thumbnail string
	Formats   []Format
}

type ProbeResult struct {
	Title    string
	Kind     string
	Entries  []MediaEntry
	Selected int
}

func (p ProbeResult) Active() MediaEntry {
	if len(p.Entries) == 0 {
		return MediaEntry{}
	}
	i := p.Selected
	if i < 0 || i >= len(p.Entries) {
		i = 0
	}
	return p.Entries[i]
}

func Probe(cfg core.Config) (ProbeResult, error) {
	return ProbeCtx(context.Background(), cfg)
}

// ProbeCtx runs yt-dlp format probing with a timeout so a hung subprocess
// cannot block the caller forever.
func ProbeCtx(ctx context.Context, cfg core.Config) (ProbeResult, error) {
	url := strings.TrimSpace(cfg.URL)
	if url == "" {
		return ProbeResult{}, fmt.Errorf("URL is required")
	}

	binary, err := ResolveBinary(cfg.YtDlpPath)
	if err != nil {
		return ProbeResult{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	args := []string{"--no-download", "-J", "--no-warnings"}
	args = append(args, core.ProbeFlags(cfg)...)
	args = append(args, url)

	out, err := executil.OutputContext(ctx, binary, args...)
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			return ProbeResult{}, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(ee.Stderr)))
		}
		return ProbeResult{}, err
	}

	return parseProbeJSON(out)
}

func parseProbeJSON(raw []byte) (ProbeResult, error) {
	var root jsonutil.Object
	if err := json.Unmarshal(raw, &root); err != nil {
		return ProbeResult{}, err
	}

	kind := jsonutil.String(root, "_type")
	title := jsonutil.String(root, "title")

	if kind == "playlist" {
		entriesRaw, ok := root["entries"]
		if !ok {
			return ProbeResult{Title: title, Kind: kind}, nil
		}
		var entries []jsonutil.Object
		if err := json.Unmarshal(entriesRaw, &entries); err != nil {
			return ProbeResult{}, err
		}
		out := ProbeResult{Title: title, Kind: kind}
		for _, e := range entries {
			if e == nil {
				continue
			}
			out.Entries = append(out.Entries, entryFromMap(e))
		}
		return out, nil
	}

	return ProbeResult{
		Title:   title,
		Kind:    kind,
		Entries: []MediaEntry{entryFromMap(root)},
	}, nil
}

func entryFromMap(m jsonutil.Object) MediaEntry {
	e := MediaEntry{
		Title:     jsonutil.String(m, "title"),
		ID:        jsonutil.String(m, "id"),
		URL:       jsonutil.String(m, "webpage_url"),
		Uploader:  jsonutil.String(m, "uploader"),
		Thumbnail: jsonutil.String(m, "thumbnail"),
	}
	if e.URL == "" {
		e.URL = jsonutil.String(m, "url")
	}
	if d := jsonutil.Float(m, "duration"); d > 0 {
		e.Duration = formatDuration(d)
	}
	for _, f := range jsonutil.Objects(m, "formats") {
		e.Formats = append(e.Formats, formatFromMap(f))
	}
	return e
}

func formatFromMap(m jsonutil.Object) Format {
	f := Format{
		ID:     jsonutil.String(m, "format_id"),
		Ext:    jsonutil.String(m, "ext"),
		Vcodec: jsonutil.String(m, "vcodec"),
		Acodec: jsonutil.String(m, "acodec"),
		Note:   jsonutil.String(m, "format_note"),
	}
	f.Resolution = jsonutil.String(m, "resolution")
	if f.Resolution == "" {
		w := jsonutil.Int(m, "width")
		h := jsonutil.Int(m, "height")
		if h > 0 {
			f.Resolution = fmt.Sprintf("%dx%d", w, h)
		}
	}
	f.FPS = jsonutil.Int(m, "fps")
	f.Filesize = jsonutil.Int64(m, "filesize")
	if f.Filesize == 0 {
		f.Filesize = jsonutil.Int64(m, "filesize_approx")
	}
	f.TBR = jsonutil.Float(m, "tbr")
	return f
}

func formatDuration(sec float64) string {
	s := int(sec + 0.5)
	if s < 3600 {
		return fmt.Sprintf("%d:%02d", s/60, s%60)
	}
	return fmt.Sprintf("%d:%02d:%02d", s/3600, (s%3600)/60, s%60)
}

func FormatLabel(f Format) string {
	parts := []string{f.ID}
	if f.Resolution != "" {
		parts = append(parts, f.Resolution)
	}
	if f.Ext != "" {
		parts = append(parts, f.Ext)
	}
	if f.Vcodec != "" && f.Vcodec != "none" {
		parts = append(parts, f.Vcodec)
	}
	if f.Acodec != "" && f.Acodec != "none" {
		parts = append(parts, f.Acodec)
	}
	if f.TBR > 0 {
		parts = append(parts, strconv.Itoa(int(f.TBR))+"k")
	}
	if f.Filesize > 0 {
		parts = append(parts, humanSize(f.Filesize))
	}
	if f.Note != "" {
		parts = append(parts, f.Note)
	}
	return strings.Join(parts, " · ")
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
