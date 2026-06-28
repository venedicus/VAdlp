package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"vadlp/internal/configdir"
	"vadlp/internal/executil"
)

// versionTokenRe matches a dotted version number anywhere in a line, e.g.
// "2025.06.25" (yt-dlp), "6.1.1-essentials_build" (ffmpeg), "2.9.0" (deno).
var versionTokenRe = regexp.MustCompile(`\d+(?:\.\d+){1,3}[\w.-]*`)

type ToolStatus struct {
	Found   bool
	Path    string
	Version string
}

type CheckResult struct {
	YtDlp  ToolStatus
	FFmpeg ToolStatus
	Deno   ToolStatus
}

func CheckTools() CheckResult {
	return CheckToolsWithPaths(DependencyPaths{})
}

func CheckToolsWithPaths(paths DependencyPaths) CheckResult {
	ytdlp := resolveTool(ytDlpBinName(), "--version", paths.YtDlp)
	ffmpeg := resolveTool(ffmpegBinName(), "-version", paths.FFmpeg)
	deno := resolveTool(denoBinName(), "--version", paths.Deno)
	return CheckResult{
		YtDlp:  toolStatusFromResolved(ytdlp),
		FFmpeg: toolStatusFromResolved(ffmpeg),
		Deno:   toolStatusFromResolved(deno),
	}
}

func toolStatusFromResolved(st resolvedTool) ToolStatus {
	if !st.Found {
		return ToolStatus{}
	}
	return ToolStatus{Found: true, Path: st.Path, Version: st.Version}
}

func probeToolVersioned(name, versionFlag string) ToolStatus {
	st := resolveTool(name, versionFlag, "")
	return toolStatusFromResolved(st)
}

func probeExact(path, versionFlag string) ToolStatus {
	if _, err := os.Stat(path); err != nil {
		return ToolStatus{}
	}
	st := ToolStatus{Found: true, Path: path}
	if out, err := executil.Output(path, versionFlag); err == nil {
		first := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
		if tok := versionTokenRe.FindString(first); tok != "" {
			st.Version = tok
		} else {
			st.Version = first
		}
	}
	return st
}

func YtDlpDownloadURL() string {
	switch runtime.GOOS {
	case "windows":
		return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe"
	case "darwin":
		return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos"
	default:
		return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp"
	}
}

func DownloadYtDlp(destDir string, progress func(pct int)) (string, error) {
	return installYtDlp(destDir, progress, false)
}

func installYtDlp(destDir string, progress func(pct int), force bool) (string, error) {
	url := YtDlpDownloadURL()
	destPath := filepath.Join(destDir, ytDlpBinName())
	if !force {
		if st := probeExact(destPath, "--version"); st.Found {
			if progress != nil {
				progress(100)
			}
			return destPath, nil
		}
	}
	if err := downloadFileForce(url, destPath, progress, true); err != nil {
		return "", fmt.Errorf("download yt-dlp: %w", err)
	}
	if progress != nil {
		progress(100)
	}
	return destPath, nil
}

func UpdateYtDlp(binPath string) (string, error) {
	out, err := executil.CombinedOutput(binPath, "-U")
	return strings.TrimSpace(string(out)), err
}

func DefaultInstallDir() string {
	if dir, err := configdir.ToolsDir(); err == nil {
		return dir
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	wd, _ := os.Getwd()
	return wd
}

func ytDlpBinName() string {
	if runtime.GOOS == "windows" {
		return "yt-dlp.exe"
	}
	return "yt-dlp"
}

func ffmpegBinName() string {
	if runtime.GOOS == "windows" {
		return "ffmpeg.exe"
	}
	return "ffmpeg"
}
