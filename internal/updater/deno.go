package updater

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"vadlp/internal/executil"
)

func DownloadDeno(destDir string, progress func(pct int)) (string, error) {
	return installDeno(destDir, progress, false)
}

func UpdateDeno(destDir string, progress func(pct int)) (string, error) {
	destPath := filepath.Join(destDir, denoBinName())
	if st := probeExact(destPath, "--version"); st.Found {
		out, err := executil.CombinedOutput(destPath, "upgrade", "-n")
		if err == nil {
			if progress != nil {
				progress(100)
			}
			return destPath, nil
		}
		_ = out
	}
	return installDeno(destDir, progress, true)
}

func installDeno(destDir string, progress func(pct int), force bool) (string, error) {
	url, err := DenoDownloadURL()
	if err != nil {
		return "", err
	}
	destPath := filepath.Join(destDir, denoBinName())
	if !force {
		if st := probeExact(destPath, "--version"); st.Found {
			if progress != nil {
				progress(100)
			}
			return destPath, nil
		}
	}

	zipPath := filepath.Join(destDir, "deno-dl.zip")
	if err := downloadFileForce(url, zipPath, progress, force); err != nil {
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
		if filepath.Base(f.Name) != name {
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
