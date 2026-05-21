package updater

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func FFmpegDownloadURL() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip", nil
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-macosarm64-gpl.zip", nil
		}
		return "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-macos64-gpl.zip", nil
	default:
		return "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz", nil
	}
}

func DownloadFFmpeg(destDir string, progress func(pct int)) (string, error) {
	url, err := FFmpegDownloadURL()
	if err != nil {
		return "", err
	}

	binName := ffmpegBinName()
	destPath := filepath.Join(destDir, binName)

	if st := probeExact(destPath, "-version"); st.Found {
		if progress != nil {
			progress(100)
		}
		return destPath, nil
	}

	if runtime.GOOS == "linux" {
		return "", fmt.Errorf("automatic ffmpeg install on Linux is not supported yet; install via your package manager and set the path in Tools")
	}

	zipPath := filepath.Join(destDir, "ffmpeg-dl.zip")
	if err := downloadFile(url, zipPath, progress); err != nil {
		return "", err
	}
	defer os.Remove(zipPath)

	if err := extractFFmpegFromZip(zipPath, destPath, binName); err != nil {
		return "", err
	}
	if progress != nil {
		progress(100)
	}
	return destPath, nil
}

func extractFFmpegFromZip(zipPath, destPath, binName string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if filepath.Base(f.Name) != binName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("%s not found in ffmpeg archive", binName)
}
