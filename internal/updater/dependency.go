package updater

import (
	"context"
	"fmt"
	"strings"
)

type DepID string

const (
	DepYtDlp  DepID = "ytdlp"
	DepFFmpeg DepID = "ffmpeg"
	DepDeno   DepID = "deno"
)

type DepLevel int

const (
	DepRequired DepLevel = iota
	DepRecommended
	DepOptional
)

type DepStatus string

const (
	DepMissing  DepStatus = "missing"
	DepFound    DepStatus = "found"
	DepOutdated DepStatus = "outdated"
	DepChecking DepStatus = "checking"
	DepError    DepStatus = "error"
	DepUnknown  DepStatus = "unknown"
)

type DepSource string

const (
	SourceCustom  DepSource = "custom"
	SourceManaged DepSource = "managed"
	SourceLocal   DepSource = "local"
	SourceSystem  DepSource = "system"
)

type DependencyPaths struct {
	YtDlp  string
	FFmpeg string
	Deno   string
}

type DependencyInfo struct {
	ID          DepID
	Level       DepLevel
	Status      DepStatus
	Path        string
	Version     string
	LatestVer   string
	UpdateAvail bool
	Source      DepSource
	Error       string
}

func LevelFor(id DepID) DepLevel {
	switch id {
	case DepYtDlp:
		return DepRequired
	case DepFFmpeg:
		return DepRecommended
	case DepDeno:
		return DepOptional
	default:
		return DepOptional
	}
}

func ResolveDependency(id DepID, paths DependencyPaths) DependencyInfo {
	info := DependencyInfo{
		ID:    id,
		Level: LevelFor(id),
	}
	custom := ""
	switch id {
	case DepYtDlp:
		custom = paths.YtDlp
		if st := resolveTool(ytDlpBinName(), "--version", custom); st.Found {
			info.applyTool(st)
		} else {
			info.Status = DepMissing
		}
	case DepFFmpeg:
		custom = paths.FFmpeg
		if st := resolveTool(ffmpegBinName(), "-version", custom); st.Found {
			info.applyTool(st)
		} else {
			info.Status = DepMissing
		}
	case DepDeno:
		custom = paths.Deno
		if st := resolveTool(denoBinName(), "--version", custom); st.Found {
			info.applyTool(st)
		} else {
			info.Status = DepMissing
		}
	default:
		info.Status = DepError
		info.Error = "unknown dependency"
	}
	return info
}

func (info *DependencyInfo) applyTool(st resolvedTool) {
	if st.Version == "" {
		info.Status = DepUnknown
	} else {
		info.Status = DepFound
	}
	info.Path = st.Path
	info.Version = st.Version
	info.Source = st.Source
}

func ResolveAll(paths DependencyPaths) []DependencyInfo {
	return []DependencyInfo{
		ResolveDependency(DepYtDlp, paths),
		ResolveDependency(DepFFmpeg, paths),
		ResolveDependency(DepDeno, paths),
	}
}

func ApplyLatestVersions(infos []DependencyInfo) {
	for i := range infos {
		latest, ok := cachedLatestTag(infos[i].ID)
		if !ok || latest == "" {
			continue
		}
		infos[i].LatestVer = latest
		if infos[i].Status != DepFound || infos[i].Version == "" {
			continue
		}
		switch infos[i].ID {
		case DepFFmpeg:
			older, comparable := FFmpegVersionOlder(infos[i].Version, latest)
			if comparable && older {
				infos[i].UpdateAvail = true
				infos[i].Status = DepOutdated
			}
		default:
			if VersionOlder(infos[i].Version, latest) {
				infos[i].UpdateAvail = true
				infos[i].Status = DepOutdated
			}
		}
	}
}

// RefreshLatestVersions fetches remote latest tags (cached) and returns infos with update info applied.
// Per-dependency fetch errors are stored in DependencyInfo.Error; a non-nil return value means at least one fetch failed.
func RefreshLatestVersions(ctx context.Context, paths DependencyPaths) ([]DependencyInfo, error) {
	infos := ResolveAll(paths)
	fetchErrs := FetchAllLatestTags(ctx)
	for id, err := range fetchErrs {
		for i := range infos {
			if infos[i].ID != id {
				continue
			}
			infos[i].Error = err.Error()
		}
	}
	ApplyLatestVersions(infos)
	if len(fetchErrs) > 0 {
		return infos, fetchErrors(fetchErrs)
	}
	return infos, nil
}

func fetchErrors(errs map[DepID]error) error {
	if len(errs) == 0 {
		return nil
	}
	var parts []string
	for _, id := range []DepID{DepYtDlp, DepFFmpeg, DepDeno} {
		if err, ok := errs[id]; ok {
			parts = append(parts, string(id)+": "+err.Error())
		}
	}
	return fmt.Errorf("%s", strings.Join(parts, "; "))
}

func NeedsAttention(infos []DependencyInfo) int {
	n := 0
	for _, d := range infos {
		switch d.Status {
		case DepMissing:
			if d.Level == DepRequired {
				n++
			} else if d.Level == DepRecommended {
				n++
			}
		case DepOutdated:
			n++
		case DepError:
			n++
		case DepUnknown:
			n++
		}
	}
	return n
}

func RequiredMissing(infos []DependencyInfo) bool {
	for _, d := range infos {
		if d.ID == DepYtDlp && d.Status == DepMissing {
			return true
		}
	}
	return false
}
