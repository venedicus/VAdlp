package downloader

import "testing"

const probeVideoJSON = `{
  "_type": "video",
  "title": "Sample",
  "id": "abc",
  "webpage_url": "https://example.com/watch?v=abc",
  "duration": 125,
  "formats": [
    {"format_id": "22", "ext": "mp4", "resolution": "1280x720", "vcodec": "avc1", "acodec": "mp4a", "tbr": 500}
  ]
}`

const probePlaylistJSON = `{
  "_type": "playlist",
  "title": "List",
  "entries": [
    {"_type": "video", "title": "One", "id": "1", "webpage_url": "https://example.com/1"},
    {"_type": "video", "title": "Two", "id": "2", "url": "https://example.com/2"}
  ]
}`

func TestParseProbeVideoJSON(t *testing.T) {
	res, err := parseProbeJSON([]byte(probeVideoJSON))
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != "video" || res.Title != "Sample" {
		t.Fatalf("meta: %+v", res)
	}
	if len(res.Entries) != 1 {
		t.Fatalf("entries: %d", len(res.Entries))
	}
	e := res.Entries[0]
	if e.ID != "abc" || e.URL != "https://example.com/watch?v=abc" {
		t.Fatalf("entry: %+v", e)
	}
	if e.Duration != "2:05" {
		t.Fatalf("duration: %q", e.Duration)
	}
	if len(e.Formats) != 1 || e.Formats[0].ID != "22" {
		t.Fatalf("formats: %+v", e.Formats)
	}
}

func TestParseProbePlaylistJSON(t *testing.T) {
	res, err := parseProbeJSON([]byte(probePlaylistJSON))
	if err != nil {
		t.Fatal(err)
	}
	if res.Kind != "playlist" || len(res.Entries) != 2 {
		t.Fatalf("playlist: %+v", res)
	}
	if res.Entries[1].URL != "https://example.com/2" {
		t.Fatalf("entry url: %+v", res.Entries[1])
	}
}

func FuzzParseProbeJSON(f *testing.F) {
	f.Add([]byte(probeVideoJSON))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseProbeJSON(data)
	})
}

func TestFormatLabel(t *testing.T) {
	lbl := FormatLabel(Format{ID: "22", Resolution: "1280x720", Ext: "mp4", TBR: 500})
	if lbl == "" || lbl[0] != '2' {
		t.Fatalf("label: %q", lbl)
	}
}
