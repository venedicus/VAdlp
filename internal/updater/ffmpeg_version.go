package updater

import (
	"strconv"
	"strings"
)

// FFmpegVersionOlder compares an installed ffmpeg version string against a BtbN release tag.
// Returns comparable=false when either side cannot be parsed (no false "outdated").
func FFmpegVersionOlder(installed, latestTag string) (older bool, comparable bool) {
	im, imin, iok := parseFFmpegInstalledVersion(installed)
	tm, tmin, tok := parseBtbNReleaseTag(latestTag)
	if !iok || !tok {
		return false, false
	}
	if im < tm {
		return true, true
	}
	if im > tm {
		return false, true
	}
	if imin < tmin {
		return true, true
	}
	return false, true
}

func parseFFmpegInstalledVersion(v string) (major, minor int, ok bool) {
	v = normalizeVersion(v)
	if v == "" {
		return 0, 0, false
	}
	base := strings.Split(v, "-")[0]
	parts := strings.Split(base, ".")
	if len(parts) == 0 {
		return 0, 0, false
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}
	if len(parts) > 1 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			minor = 0
		}
	}
	return major, minor, true
}

func parseBtbNReleaseTag(tag string) (major, minor int, ok bool) {
	tag = strings.TrimSpace(tag)
	tag = strings.TrimPrefix(tag, "n")
	tag = strings.TrimPrefix(tag, "N")
	if tag == "" {
		return 0, 0, false
	}
	base := strings.Split(tag, "-")[0]
	parts := strings.Split(base, ".")
	if len(parts) == 0 {
		return 0, 0, false
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}
	if len(parts) > 1 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			minor = 0
		}
	}
	return major, minor, true
}
