package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const versionCacheTTL = 24 * time.Hour

type ghRelease struct {
	TagName string `json:"tag_name"`
}

var (
	versionCacheMu sync.RWMutex
	versionCache   = map[DepID]cachedTag{}
	httpClient     = &http.Client{Timeout: 8 * time.Second}
)

type cachedTag struct {
	tag     string
	fetched time.Time
}

// FetchAllLatestTags fetches latest release tags from GitHub. Partial failures are returned per dependency.
func FetchAllLatestTags(ctx context.Context) map[DepID]error {
	type job struct {
		id           DepID
		owner, repo  string
	}
	jobs := []job{
		{DepYtDlp, "yt-dlp", "yt-dlp"},
		{DepDeno, "denoland", "deno"},
		{DepFFmpeg, "BtbN", "FFmpeg-Builds"},
	}
	errs := make(map[DepID]error, len(jobs))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, j := range jobs {
		j := j
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := fetchLatestTag(ctx, j.id, j.owner, j.repo); err != nil {
				mu.Lock()
				errs[j.id] = err
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return errs
}

func fetchLatestTag(ctx context.Context, id DepID, owner, repo string) (string, error) {
	if tag, ok := cachedLatestTag(id); ok {
		return tag, nil
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "VAdlp")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("github api %s/%s: %s", owner, repo, strings.TrimSpace(string(body)))
	}
	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	tag := normalizeVersion(rel.TagName)
	versionCacheMu.Lock()
	versionCache[id] = cachedTag{tag: tag, fetched: time.Now()}
	versionCacheMu.Unlock()
	return tag, nil
}

type AppReleaseInfo struct {
	Version string
	URL     string
}

type cachedAppRelease struct {
	info    AppReleaseInfo
	fetched time.Time
}

var (
	appReleaseCacheMu sync.RWMutex
	appReleaseCache   *cachedAppRelease
)

// FetchLatestAppRelease fetches VAdlp's own latest GitHub release (cached for versionCacheTTL).
func FetchLatestAppRelease(ctx context.Context) (AppReleaseInfo, error) {
	appReleaseCacheMu.RLock()
	if appReleaseCache != nil && time.Since(appReleaseCache.fetched) < versionCacheTTL {
		info := appReleaseCache.info
		appReleaseCacheMu.RUnlock()
		return info, nil
	}
	appReleaseCacheMu.RUnlock()

	url := "https://api.github.com/repos/venedicus/VAdlp/releases/latest"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return AppReleaseInfo{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "VAdlp")

	resp, err := httpClient.Do(req)
	if err != nil {
		return AppReleaseInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return AppReleaseInfo{}, fmt.Errorf("github api venedicus/VAdlp: %s", strings.TrimSpace(string(body)))
	}
	var rel struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return AppReleaseInfo{}, err
	}
	info := AppReleaseInfo{Version: normalizeVersion(rel.TagName), URL: rel.HTMLURL}
	appReleaseCacheMu.Lock()
	appReleaseCache = &cachedAppRelease{info: info, fetched: time.Now()}
	appReleaseCacheMu.Unlock()
	return info, nil
}

func cachedLatestTag(id DepID) (string, bool) {
	versionCacheMu.RLock()
	defer versionCacheMu.RUnlock()
	c, ok := versionCache[id]
	if !ok || time.Since(c.fetched) > versionCacheTTL {
		return "", false
	}
	return c.tag, c.tag != ""
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	if idx := strings.Index(v, " "); idx >= 0 {
		v = v[:idx]
	}
	return v
}

// VersionOlder reports whether current is strictly older than latest by numeric segments.
func VersionOlder(current, latest string) bool {
	current = normalizeVersion(current)
	latest = normalizeVersion(latest)
	if current == "" || latest == "" {
		return false
	}
	if current == latest {
		return false
	}
	cParts := versionParts(current)
	lParts := versionParts(latest)
	max := len(cParts)
	if len(lParts) > max {
		max = len(lParts)
	}
	for i := 0; i < max; i++ {
		cn, cl := segmentValue(cParts, i)
		ln, ll := segmentValue(lParts, i)
		if !cl && !ll {
			return strings.Join(cParts, ".") < strings.Join(lParts, ".")
		}
		if !cl {
			return true
		}
		if !ll {
			return false
		}
		if cn != ln {
			return cn < ln
		}
	}
	return false
}

func versionParts(v string) []string {
	v = strings.ReplaceAll(v, "-", ".")
	return strings.Split(v, ".")
}

func segmentValue(parts []string, i int) (num int, ok bool) {
	if i >= len(parts) {
		return 0, false
	}
	n, err := strconv.Atoi(parts[i])
	if err != nil {
		return 0, false
	}
	return n, true
}
