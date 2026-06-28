package updater

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ulikunitz/xz"
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
	return installFFmpeg(destDir, progress, false)
}

func UpdateFFmpeg(destDir string, progress func(pct int)) (string, error) {
	return installFFmpeg(destDir, progress, true)
}

func installFFmpeg(destDir string, progress func(pct int), force bool) (string, error) {
	url, err := FFmpegDownloadURL()
	if err != nil {
		return "", err
	}

	binName := ffmpegBinName()
	destPath := filepath.Join(destDir, binName)

	if !force {
		if st := probeExact(destPath, "-version"); st.Found {
			if progress != nil {
				progress(100)
			}
			return destPath, nil
		}
	}

	if runtime.GOOS == "linux" {
		return installFFmpegLinux(url, destDir, destPath, binName, progress, force)
	}

	archivePath := filepath.Join(destDir, "ffmpeg-dl.zip")
	if err := downloadFileForce(url, archivePath, progress, force); err != nil {
		return "", err
	}
	defer os.Remove(archivePath)

	if err := extractFFmpegFromZip(archivePath, destPath, binName); err != nil {
		return "", err
	}
	if progress != nil {
		progress(100)
	}
	return destPath, nil
}

func installFFmpegLinux(url, destDir, destPath, binName string, progress func(pct int), force bool) (string, error) {
	archivePath := filepath.Join(destDir, "ffmpeg-dl.tar.xz")
	if err := downloadFileForce(url, archivePath, progress, force); err != nil {
		return "", err
	}
	defer os.Remove(archivePath)

	if err := extractFFmpegFromTarXz(archivePath, destPath, binName); err != nil {
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
		return extractToPath(f.Open, destPath)
	}
	return fmt.Errorf("%s not found in ffmpeg archive", binName)
}

func extractFFmpegFromTarXz(archivePath, destPath, binName string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	xr, err := xz.NewReader(f)
	if err != nil {
		return err
	}
	tr := tar.NewReader(xr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg || filepath.Base(hdr.Name) != binName {
			continue
		}
		return writeReaderToPath(tr, destPath)
	}
	return fmt.Errorf("%s not found in ffmpeg archive", binName)
}

func extractToPath(open func() (io.ReadCloser, error), destPath string) error {
	rc, err := open()
	if err != nil {
		return err
	}
	defer rc.Close()
	return writeReaderToPath(rc, destPath)
}

func writeReaderToPath(r io.Reader, destPath string) error {
	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, r)
	return err
}
