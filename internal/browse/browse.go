// Package browse turns yt-dlp's metadata extraction into a browsable
// catalogue: keyword search, and opening a channel or playlist URL. It is an
// optional module — nothing else in the app depends on it, and the UI hides
// it unless the user enables it in settings.
//
// Listings deliberately run with --flat-playlist: resolving full metadata
// (and therefore every format) for each of N results costs one extractor
// round-trip per video and takes minutes. Formats are resolved lazily, for
// the single video the user actually picks, by the existing downloader.Probe.
package browse

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"vadlp/internal/core"
	"vadlp/internal/executil"
	"vadlp/internal/jsonutil"
	"vadlp/internal/updater"
)

// listTimeout bounds a listing subprocess. Searches are usually a couple of
// seconds; large channels are the slow case.
const listTimeout = 60 * time.Second

// DefaultLimit is the number of search results requested when the caller
// does not specify one.
const DefaultLimit = 30

// maxLimit caps the result count. yt-dlp will happily page through thousands
// of entries, which is slow and useless in a grid.
const maxLimit = 100

// Item kinds. A listing can mix them: a channel page yields playlists as
// well as videos.
const (
	KindVideo    = "video"
	KindPlaylist = "playlist"
	KindChannel  = "channel"
)

// Item is one entry of a listing.
type Item struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Uploader    string `json:"uploader"`
	Duration    string `json:"duration"`
	DurationSec int    `json:"durationSec"`
	Thumbnail   string `json:"thumbnail"`
	ViewCount   int64  `json:"viewCount"`
	Live        bool   `json:"live"`
}

// Result is a listing: search results, or the contents of a channel or
// playlist.
type Result struct {
	Title string `json:"title"`
	Query string `json:"query"`
	Items []Item `json:"items"`
}

// ErrEmptyQuery is returned when a search or open request carries no target.
var ErrEmptyQuery = errors.New("query is required")

// Search lists videos matching a free-text query. When the query is itself a
// URL it is opened directly, so pasting a link into the search box does what
// the user means.
func Search(ctx context.Context, cfg core.Config, query string, limit int) (Result, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return Result{}, ErrEmptyQuery
	}
	if isURL(q) {
		return Open(ctx, cfg, q)
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	// ytsearchN: is yt-dlp's own search pseudo-URL. The query is passed as a
	// single argv element, so it needs no quoting or escaping of its own.
	res, err := list(ctx, cfg, "ytsearch"+strconv.Itoa(limit)+":"+q)
	if err != nil {
		return Result{}, err
	}
	res.Query = q
	if res.Title == "" {
		res.Title = q
	}
	return res, nil
}

// Open lists the contents of a channel or playlist URL.
func Open(ctx context.Context, cfg core.Config, target string) (Result, error) {
	t := strings.TrimSpace(target)
	if t == "" {
		return Result{}, ErrEmptyQuery
	}
	res, err := list(ctx, cfg, t)
	if err != nil {
		return Result{}, err
	}
	res.Query = t
	return res, nil
}

func list(ctx context.Context, cfg core.Config, target string) (Result, error) {
	binary, err := updater.ResolveYtDlpPath(cfg.YtDlpPath)
	if err != nil {
		return Result{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, listTimeout)
	defer cancel()

	// --flat-playlist keeps this to one extractor call; --no-warnings keeps
	// stderr clean so a non-zero exit carries a real error message.
	args := []string{"-J", "--flat-playlist", "--no-warnings"}
	// Network settings (cookies, proxy, deno runtime) come from the user's
	// config, so private, age-gated and members-only listings work exactly
	// as they do for downloads.
	args = append(args, core.AppendNetworkArgs(nil, cfg)...)
	args = append(args, target)

	out, err := executil.OutputContext(ctx, binary, args...)
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return Result{}, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(ee.Stderr)))
		}
		return Result{}, err
	}
	return parseListing(out)
}

func parseListing(raw []byte) (Result, error) {
	root, err := jsonutil.Decode(raw)
	if err != nil {
		return Result{}, err
	}

	res := Result{Title: jsonutil.String(root, "title")}
	entries := jsonutil.Objects(root, "entries")
	if len(entries) == 0 {
		// A bare video URL resolves to a single object rather than a listing.
		if jsonutil.String(root, "id") != "" {
			res.Items = append(res.Items, itemFrom(root))
		}
		return res, nil
	}
	for _, e := range entries {
		if e == nil {
			continue
		}
		item := itemFrom(e)
		if item.ID == "" && item.URL == "" {
			continue
		}
		res.Items = append(res.Items, item)
	}
	return res, nil
}

func itemFrom(m jsonutil.Object) Item {
	it := Item{
		ID:        jsonutil.String(m, "id"),
		Title:     jsonutil.String(m, "title"),
		URL:       jsonutil.String(m, "url"),
		ViewCount: jsonutil.Int64(m, "view_count"),
	}
	if it.URL == "" {
		it.URL = jsonutil.String(m, "webpage_url")
	}
	// Channel name lives under different keys depending on the extractor and
	// on whether the entry came from a flat listing.
	for _, key := range []string{"channel", "uploader", "playlist_uploader", "creator"} {
		if v := jsonutil.String(m, key); v != "" {
			it.Uploader = v
			break
		}
	}
	if d := jsonutil.Float(m, "duration"); d > 0 {
		it.DurationSec = int(d + 0.5)
		it.Duration = formatDuration(it.DurationSec)
	}
	it.Kind = kindOf(m)
	it.Live = isLive(m)
	it.Thumbnail = thumbnailOf(m, it.ID)
	if it.URL == "" && it.Kind == KindVideo && it.ID != "" {
		it.URL = "https://www.youtube.com/watch?v=" + it.ID
	}
	return it
}

// kindOf maps yt-dlp's _type plus URL shape onto our three kinds. Flat
// entries report _type "url", which is ambiguous, so the URL decides.
func kindOf(m jsonutil.Object) string {
	switch jsonutil.String(m, "_type") {
	case "playlist", "multi_video":
		return KindPlaylist
	}
	if jsonutil.String(m, "ie_key") == "YoutubeTab" {
		u := jsonutil.String(m, "url")
		if strings.Contains(u, "/playlist") || strings.Contains(u, "list=") {
			return KindPlaylist
		}
		return KindChannel
	}
	return KindVideo
}

func isLive(m jsonutil.Object) bool {
	if jsonutil.String(m, "live_status") == "is_live" {
		return true
	}
	return jsonutil.Bool(m, "is_live")
}

// thumbnailOf prefers the widest thumbnail that is still small enough to be
// a grid tile, then falls back to YouTube's predictable still URL, which
// flat listings often omit.
func thumbnailOf(m jsonutil.Object, id string) string {
	if t := jsonutil.String(m, "thumbnail"); t != "" {
		return t
	}
	best := ""
	bestW := 0
	for _, t := range jsonutil.Objects(m, "thumbnails") {
		u := jsonutil.String(t, "url")
		if u == "" {
			continue
		}
		w := jsonutil.Int(t, "width")
		// Prefer <= 640px wide; anything bigger is wasted bytes in a tile.
		if best == "" || (w > bestW && w <= 640) {
			best, bestW = u, w
		}
	}
	if best != "" {
		return best
	}
	if id != "" && isYouTubeID(id) {
		return "https://i.ytimg.com/vi/" + id + "/hqdefault.jpg"
	}
	return ""
}

// isYouTubeID reports whether id has the shape of a YouTube video id, so the
// constructed still URL is not pointed at an unrelated site's id.
func isYouTubeID(id string) bool {
	if len(id) != 11 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}

func isURL(s string) bool {
	if !strings.Contains(s, "://") {
		return false
	}
	u, err := url.Parse(s)
	return err == nil && u.Host != ""
}

func formatDuration(sec int) string {
	if sec < 3600 {
		return fmt.Sprintf("%d:%02d", sec/60, sec%60)
	}
	return fmt.Sprintf("%d:%02d:%02d", sec/3600, (sec%3600)/60, sec%60)
}
