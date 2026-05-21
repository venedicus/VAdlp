package updater

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

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
	return CheckResult{
		YtDlp:  probeToolVersioned(ytDlpBinName(), "--version"),
		FFmpeg: probeToolVersioned(ffmpegBinName(), "-version"),
		Deno:   CheckDeno(),
	}
}

func probeToolVersioned(name, versionFlag string) ToolStatus {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), name)
		if st := probeExact(candidate, versionFlag); st.Found {
			return st
		}
		candidate = filepath.Join(filepath.Dir(exe), "bin", name)
		if st := probeExact(candidate, versionFlag); st.Found {
			return st
		}
	}
	if p, err := exec.LookPath(name); err == nil {
		return probeExact(p, versionFlag)
	}
	return ToolStatus{}
}

func probeExact(path, versionFlag string) ToolStatus {
	if _, err := os.Stat(path); err != nil {
		return ToolStatus{}
	}
	st := ToolStatus{Found: true, Path: path}
	if out, err := exec.Command(path, versionFlag).Output(); err == nil {
		first := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
		if idx := strings.Index(first, "version "); idx != -1 {
			first = strings.TrimSpace(first[idx+len("version "):])
			first = strings.Fields(first)[0]
		}
		st.Version = first
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
	url := YtDlpDownloadURL()
	destPath := filepath.Join(destDir, ytDlpBinName())
	if err := downloadFile(url, destPath, progress); err != nil {
		return "", fmt.Errorf("download yt-dlp: %w", err)
	}
	if progress != nil {
		progress(100)
	}
	return destPath, nil
}

func UpdateYtDlp(binPath string) (string, error) {
	out, err := exec.Command(binPath, "-U").CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func DefaultInstallDir() string {
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
