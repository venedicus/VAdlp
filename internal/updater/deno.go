package updater

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func denoBinName() string {
	if runtime.GOOS == "windows" {
		return "deno.exe"
	}
	return "deno"
}

func DenoDownloadURL() (string, error) {
	arch := runtime.GOARCH
	switch runtime.GOOS {
	case "windows":
		if arch == "arm64" {
			return "https://github.com/denoland/deno/releases/latest/download/deno-aarch64-pc-windows-msvc.zip", nil
		}
		return "https://github.com/denoland/deno/releases/latest/download/deno-x86_64-pc-windows-msvc.zip", nil
	case "darwin":
		if arch == "arm64" {
			return "https://github.com/denoland/deno/releases/latest/download/deno-aarch64-apple-darwin.zip", nil
		}
		return "https://github.com/denoland/deno/releases/latest/download/deno-x86_64-apple-darwin.zip", nil
	default:
		if arch == "arm64" {
			return "https://github.com/denoland/deno/releases/latest/download/deno-aarch64-unknown-linux-gnu.zip", nil
		}
		return "https://github.com/denoland/deno/releases/latest/download/deno-x86_64-unknown-linux-gnu.zip", nil
	}
}

func CheckDeno() ToolStatus {
	return probeToolVersioned(denoBinName(), "--version")
}

func DownloadDeno(destDir string, progress func(pct int)) (string, error) {
	url, err := DenoDownloadURL()
	if err != nil {
		return "", err
	}
	destPath := filepath.Join(destDir, denoBinName())
	if st := probeExact(destPath, "--version"); st.Found {
		if progress != nil {
			progress(100)
		}
		return destPath, nil
	}

	zipPath := filepath.Join(destDir, "deno-dl.zip")
	if err := downloadFile(url, zipPath, progress); err != nil {
		return "", err
	}
	defer os.Remove(zipPath)

	if err := extractDenoBinary(zipPath, destPath); err != nil {
		return "", err
	}
	if progress != nil {
		progress(100)
	}
	return destPath, nil
}

func extractDenoBinary(zipPath, destPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	name := denoBinName()
	for _, f := range r.File {
		base := filepath.Base(f.Name)
		if base != name {
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
	return fmt.Errorf("%s not found in deno archive", name)
}

func downloadFile(url, dest string, progress func(pct int)) error {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	total := resp.ContentLength
	var written int64
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			wn, wErr := f.Write(buf[:n])
			written += int64(wn)
			if wErr != nil {
				return wErr
			}
			if progress != nil && total > 0 {
				progress(int(written * 100 / total))
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	return nil
}
