package browse

import (
	"context"
	"testing"

	"vadlp/internal/core"
)

// searchListing is the shape `yt-dlp -J --flat-playlist ytsearch2:...`
// produces: a playlist wrapper whose entries are flat "url" stubs with a
// thumbnails array rather than a single thumbnail field.
const searchListing = `{
  "_type": "playlist",
  "id": "ytsearch2:go concurrency",
  "title": "go concurrency",
  "entries": [
    {
      "_type": "url",
      "ie_key": "Youtube",
      "id": "oV9rvDllKEg",
      "url": "https://www.youtube.com/watch?v=oV9rvDllKEg",
      "title": "Concurrency is not Parallelism",
      "duration": 1830.0,
      "channel": "Gopher Academy",
      "view_count": 412345,
      "live_status": "not_live",
      "thumbnails": [
        {"url": "https://i.ytimg.com/vi/x/default.jpg", "width": 120},
        {"url": "https://i.ytimg.com/vi/x/hqdefault.jpg", "width": 480},
        {"url": "https://i.ytimg.com/vi/x/maxres.jpg", "width": 1280}
      ]
    },
    {
      "_type": "url",
      "ie_key": "Youtube",
      "id": "B9lP-E4J_lc",
      "url": "https://www.youtube.com/watch?v=B9lP-E4J_lc",
      "title": "Live coding stream",
      "duration": 45.0,
      "uploader": "Someone",
      "live_status": "is_live"
    }
  ]
}`

func TestParseListingSearch(t *testing.T) {
	res, err := parseListing([]byte(searchListing))
	if err != nil {
		t.Fatalf("parseListing: %v", err)
	}
	if res.Title != "go concurrency" {
		t.Errorf("Title = %q, want %q", res.Title, "go concurrency")
	}
	if len(res.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(res.Items))
	}

	first := res.Items[0]
	if first.ID != "oV9rvDllKEg" {
		t.Errorf("ID = %q", first.ID)
	}
	if first.Kind != KindVideo {
		t.Errorf("Kind = %q, want %q", first.Kind, KindVideo)
	}
	if first.Uploader != "Gopher Academy" {
		t.Errorf("Uploader = %q", first.Uploader)
	}
	if first.Duration != "30:30" {
		t.Errorf("Duration = %q, want 30:30", first.Duration)
	}
	if first.DurationSec != 1830 {
		t.Errorf("DurationSec = %d, want 1830", first.DurationSec)
	}
	if first.ViewCount != 412345 {
		t.Errorf("ViewCount = %d", first.ViewCount)
	}
	if first.Live {
		t.Error("Live = true, want false")
	}
	// 480px is the widest thumbnail at or below the 640px tile budget.
	if want := "https://i.ytimg.com/vi/x/hqdefault.jpg"; first.Thumbnail != want {
		t.Errorf("Thumbnail = %q, want %q", first.Thumbnail, want)
	}

	second := res.Items[1]
	if !second.Live {
		t.Error("Live = false, want true for live_status is_live")
	}
	// No thumbnails array: fall back to the derived YouTube still.
	if want := "https://i.ytimg.com/vi/B9lP-E4J_lc/hqdefault.jpg"; second.Thumbnail != want {
		t.Errorf("fallback Thumbnail = %q, want %q", second.Thumbnail, want)
	}
	if second.Uploader != "Someone" {
		t.Errorf("Uploader = %q, want the uploader key when channel is absent", second.Uploader)
	}
}

func TestParseListingChannelMixesPlaylistsAndVideos(t *testing.T) {
	const raw = `{
      "_type": "playlist",
      "title": "Some Channel - Videos",
      "entries": [
        {"_type": "url", "ie_key": "YoutubeTab", "id": "PL123",
         "url": "https://www.youtube.com/playlist?list=PL123", "title": "A playlist"},
        {"_type": "url", "ie_key": "YoutubeTab", "id": "UC456",
         "url": "https://www.youtube.com/@someone", "title": "A channel"},
        {"_type": "url", "ie_key": "Youtube", "id": "abcdefghijk",
         "url": "https://www.youtube.com/watch?v=abcdefghijk", "title": "A video"}
      ]
    }`
	res, err := parseListing([]byte(raw))
	if err != nil {
		t.Fatalf("parseListing: %v", err)
	}
	want := []string{KindPlaylist, KindChannel, KindVideo}
	if len(res.Items) != len(want) {
		t.Fatalf("len(Items) = %d, want %d", len(res.Items), len(want))
	}
	for i, kind := range want {
		if res.Items[i].Kind != kind {
			t.Errorf("Items[%d].Kind = %q, want %q", i, res.Items[i].Kind, kind)
		}
	}
}

func TestParseListingSingleVideo(t *testing.T) {
	// A bare video URL resolves to an object with no entries array.
	const raw = `{"id": "oV9rvDllKEg", "title": "Just one",
	              "webpage_url": "https://www.youtube.com/watch?v=oV9rvDllKEg",
	              "duration": 61.0, "channel": "Chan"}`
	res, err := parseListing([]byte(raw))
	if err != nil {
		t.Fatalf("parseListing: %v", err)
	}
	if len(res.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(res.Items))
	}
	if res.Items[0].URL != "https://www.youtube.com/watch?v=oV9rvDllKEg" {
		t.Errorf("URL = %q, want webpage_url to be used", res.Items[0].URL)
	}
	if res.Items[0].Duration != "1:01" {
		t.Errorf("Duration = %q, want 1:01", res.Items[0].Duration)
	}
}

func TestParseListingSkipsEmptyEntries(t *testing.T) {
	// yt-dlp emits null entries for unavailable/private videos.
	const raw = `{"_type": "playlist", "title": "x",
	              "entries": [null, {"title": "no id or url"},
	                          {"id": "abcdefghijk", "title": "ok"}]}`
	res, err := parseListing([]byte(raw))
	if err != nil {
		t.Fatalf("parseListing: %v", err)
	}
	if len(res.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1 (null and id-less entries dropped)", len(res.Items))
	}
	if res.Items[0].ID != "abcdefghijk" {
		t.Errorf("ID = %q", res.Items[0].ID)
	}
}

func TestParseListingInvalidJSON(t *testing.T) {
	if _, err := parseListing([]byte("not json")); err == nil {
		t.Fatal("parseListing(invalid) = nil error, want error")
	}
}

func TestSearchRejectsEmptyQuery(t *testing.T) {
	if _, err := Search(context.Background(), core.Config{}, "   ", 10); err != ErrEmptyQuery {
		t.Fatalf("Search(blank) error = %v, want ErrEmptyQuery", err)
	}
	if _, err := Open(context.Background(), core.Config{}, ""); err != ErrEmptyQuery {
		t.Fatalf("Open(blank) error = %v, want ErrEmptyQuery", err)
	}
}

func TestIsURL(t *testing.T) {
	cases := map[string]bool{
		"https://www.youtube.com/watch?v=x": true,
		"http://example.com":                true,
		"go concurrency":                    false,
		"how to use https":                  false,
		"https://":                          false,
		"":                                  false,
	}
	for in, want := range cases {
		if got := isURL(in); got != want {
			t.Errorf("isURL(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIsYouTubeID(t *testing.T) {
	cases := map[string]bool{
		"oV9rvDllKEg":  true,
		"a-b_c1234_5":  true,
		"tooshort":     false,
		"waytoolongid": false,
		"has spaces!":  false,
	}
	for in, want := range cases {
		if got := isYouTubeID(in); got != want {
			t.Errorf("isYouTubeID(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	cases := map[int]string{0: "0:00", 61: "1:01", 599: "9:59", 3600: "1:00:00", 3661: "1:01:01"}
	for sec, want := range cases {
		if got := formatDuration(sec); got != want {
			t.Errorf("formatDuration(%d) = %q, want %q", sec, got, want)
		}
	}
}
