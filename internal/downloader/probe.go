package downloader

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"vadlp/internal/core"
	"vadlp/internal/executil"
)

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
	url := strings.TrimSpace(cfg.URL)
	if url == "" {
		return ProbeResult{}, fmt.Errorf("URL is required")
	}

	binary, err := ResolveBinary(cfg.YtDlpPath)
	if err != nil {
		return ProbeResult{}, err
	}

	args := []string{"--no-download", "-J", "--no-warnings"}
	args = append(args, core.ProbeFlags(cfg)...)
	args = append(args, url)

	out, err := executil.Command(binary, args...).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			return ProbeResult{}, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(ee.Stderr)))
		}
		return ProbeResult{}, err
	}

	return parseProbeJSON(out)
}

func parseProbeJSON(raw []byte) (ProbeResult, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return ProbeResult{}, err
	}

	kind := stringField(root, "_type")
	title := stringField(root, "title")

	if kind == "playlist" {
		entriesRaw, ok := root["entries"]
		if !ok {
			return ProbeResult{Title: title, Kind: kind}, nil
		}
		var entries []map[string]json.RawMessage
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

func entryFromMap(m map[string]json.RawMessage) MediaEntry {
	e := MediaEntry{
		Title:     stringField(m, "title"),
		ID:        stringField(m, "id"),
		URL:       stringField(m, "webpage_url"),
		Uploader:  stringField(m, "uploader"),
		Thumbnail: stringField(m, "thumbnail"),
	}
	if e.URL == "" {
		e.URL = stringField(m, "url")
	}
	if d := floatField(m, "duration"); d > 0 {
		e.Duration = formatDuration(d)
	}
	if raw, ok := m["formats"]; ok {
		var formats []map[string]json.RawMessage
		if err := json.Unmarshal(raw, &formats); err == nil {
			for _, f := range formats {
				e.Formats = append(e.Formats, formatFromMap(f))
			}
		}
	}
	return e
}

func formatFromMap(m map[string]json.RawMessage) Format {
	f := Format{
		ID:     stringField(m, "format_id"),
		Ext:    stringField(m, "ext"),
		Vcodec: stringField(m, "vcodec"),
		Acodec: stringField(m, "acodec"),
		Note:   stringField(m, "format_note"),
	}
	f.Resolution = stringField(m, "resolution")
	if f.Resolution == "" {
		w := intField(m, "width")
		h := intField(m, "height")
		if h > 0 {
			f.Resolution = fmt.Sprintf("%dx%d", w, h)
		}
	}
	f.FPS = intField(m, "fps")
	f.Filesize = int64Field(m, "filesize")
	if f.Filesize == 0 {
		f.Filesize = int64Field(m, "filesize_approx")
	}
	f.TBR = floatField(m, "tbr")
	return f
}

func stringField(m map[string]json.RawMessage, key string) string {
	raw, ok := m[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return ""
	}
	return s
}

func floatField(m map[string]json.RawMessage, key string) float64 {
	raw, ok := m[key]
	if !ok {
		return 0
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err != nil {
		return 0
	}
	return f
}

func intField(m map[string]json.RawMessage, key string) int {
	return int(floatField(m, key))
}

func int64Field(m map[string]json.RawMessage, key string) int64 {
	raw, ok := m[key]
	if !ok {
		return 0
	}
	var n int64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return int64(f)
	}
	return 0
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
