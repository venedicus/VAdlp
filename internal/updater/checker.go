package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ToolStatus holds resolved info about an external binary.
type ToolStatus struct {
	Found   bool
	Path    string
	Version string
}

// CheckResult is returned by CheckTools.
type CheckResult struct {
	YtDlp  ToolStatus
	FFmpeg ToolStatus
}

// CheckTools probes for yt-dlp and ffmpeg in PATH and next to the executable.
func CheckTools() CheckResult {
	return CheckResult{
		YtDlp:  probeToolVersioned(ytDlpBinName(), "--version"),
		FFmpeg: probeToolVersioned(ffmpegBinName(), "-version"),
	}
}

func probeToolVersioned(name, versionFlag string) ToolStatus {
	// Try next to our own executable first (bundled layout).
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
	// Then fall back to PATH.
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
		// yt-dlp: first line is the version; ffmpeg: "ffmpeg version X …"
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

	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return "", fmt.Errorf("download request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	total := resp.ContentLength
	var written int64
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := f.Write(buf[:n]); wErr != nil {
				return "", fmt.Errorf("write: %w", wErr)
			}
			written += int64(n)
			if progress != nil && total > 0 {
				progress(int(written * 100 / total))
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return "", fmt.Errorf("read: %w", readErr)
		}
	}
	if progress != nil {
		progress(100)
	}
	return destPath, nil
}

// UpdateYtDlp runs `yt-dlp -U` to self-update an existing installation.
func UpdateYtDlp(binPath string) (string, error) {
	out, err := exec.Command(binPath, "-U").CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// DefaultInstallDir returns the directory next to our own executable (preferred
// for bundled installs) or falls back to the current working directory.
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
