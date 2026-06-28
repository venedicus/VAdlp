// Release packaging for CI (replaces package-release.sh).
package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: releasepack <format> <binary-or-bundle> <archive>")
		fmt.Fprintln(os.Stderr, "format: tar | zip | dmg | appimage")
		fmt.Fprintln(os.Stderr, "for dmg, <binary-or-bundle> is a .app bundle directory")
		os.Exit(2)
	}
	format, source, archive := os.Args[1], os.Args[2], os.Args[3]

	if _, err := os.Stat(source); err != nil {
		fmt.Fprintf(os.Stderr, "source not found: %s\n", source)
		os.Exit(1)
	}

	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	staging, err := os.MkdirTemp("", "vadlp-pack-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(staging)

	archivePath := archive
	if !filepath.IsAbs(archivePath) && !isWindowsAbs(archivePath) {
		archivePath = filepath.Join(root, archivePath)
	}

	switch format {
	case "tar":
		if err = stageBinaryWithDocs(source, staging, root); err == nil {
			err = packTar(staging, archivePath)
		}
	case "zip":
		if err = stageBinaryWithDocs(source, staging, root); err == nil {
			err = packZip(staging, archivePath)
		}
	case "dmg":
		err = packDmg(source, staging, archivePath)
	case "appimage":
		err = packAppImage(staging, source, archivePath)
	default:
		fmt.Fprintf(os.Stderr, "unknown format: %s\n", format)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("created", archivePath)
}

// stageBinaryWithDocs copies a flat executable plus README/LICENSE into
// staging, for the CLI-style tar/zip archives.
func stageBinaryWithDocs(binary, staging, root string) error {
	binName := filepath.Base(binary)
	if err := copyFile(binary, filepath.Join(staging, binName)); err != nil {
		return err
	}
	for _, name := range []string{"README.md", "LICENSE"} {
		src := filepath.Join(root, name)
		if _, err := os.Stat(src); err == nil {
			if err := copyFile(src, filepath.Join(staging, name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func isWindowsAbs(path string) bool {
	if len(path) < 3 {
		return false
	}
	return ((path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z')) && path[1] == ':'
}

func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	// Preserve the source's mode (notably the executable bit) — os.Create
	// would otherwise always write 0666, silently stripping +x from binaries.
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// copyDir recursively copies src into dst, preserving file modes and
// symlinks (used to copy a .app bundle into a dmg staging directory).
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.Mode()&os.ModeSymlink != 0 {
			linkDest, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(linkDest, target)
		}
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
}

func packTar(staging, archivePath string) error {
	f, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	return filepath.Walk(staging, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(staging, path)
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		data, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tw, data)
		data.Close()
		return copyErr
	})
}

func packZip(staging, archivePath string) error {
	f, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()

	return filepath.Walk(staging, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(staging, path)
		if err != nil {
			return err
		}
		w, err := zw.Create(filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		data, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(w, data)
		data.Close()
		return copyErr
	})
}

// packDmg stages a copy of the .app bundle at appBundle (plus an
// /Applications symlink for drag-to-install) and wraps it in a .dmg.
func packDmg(appBundle, staging, archivePath string) error {
	dest := filepath.Join(staging, filepath.Base(appBundle))
	if err := copyDir(appBundle, dest); err != nil {
		return err
	}
	if err := os.Symlink("/Applications", filepath.Join(staging, "Applications")); err != nil {
		return err
	}
	cmd := exec.Command("hdiutil", "create", "-volname", "VAdlp", "-srcfolder", staging, "-ov", "-format", "UDZO", archivePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func packAppImage(staging, binary, archivePath string) error {
	appDir := filepath.Join(staging, "VAdlp.AppDir")
	binDir := filepath.Join(appDir, "usr", "bin")
	desktopDir := filepath.Join(appDir, "usr", "share", "applications")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(desktopDir, 0o755); err != nil {
		return err
	}
	if err := copyFile(binary, filepath.Join(binDir, "vadlp")); err != nil {
		return err
	}
	_ = os.Chmod(filepath.Join(binDir, "vadlp"), 0o755)

	desktop := `[Desktop Entry]
Name=VAdlp
Comment=Desktop GUI for yt-dlp
Exec=vadlp
Icon=vadlp
Type=Application
Categories=AudioVideo;Network;
`
	if err := os.WriteFile(filepath.Join(desktopDir, "vadlp.desktop"), []byte(desktop), 0o644); err != nil {
		return err
	}

	appRun := `#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
exec "$HERE/usr/bin/vadlp" "$@"
`
	if err := os.WriteFile(filepath.Join(appDir, "AppRun"), []byte(appRun), 0o755); err != nil {
		return err
	}

	linuxdeploy := filepath.Join(staging, "linuxdeploy-x86_64.AppImage")
	if err := downloadURL(
		"https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage",
		linuxdeploy,
	); err != nil {
		return err
	}
	_ = os.Chmod(linuxdeploy, 0o755)

	gtkPlugin := filepath.Join(staging, "linuxdeploy-plugin-gtk-x86_64.AppImage")
	if err := downloadURL(
		"https://github.com/linuxdeploy/linuxdeploy-plugin-gtk/releases/download/continuous/linuxdeploy-plugin-gtk-x86_64.AppImage",
		gtkPlugin,
	); err != nil {
		return err
	}
	_ = os.Chmod(gtkPlugin, 0o755)

	cmd := exec.Command(linuxdeploy, "--appdir", appDir, "--plugin", "gtk", "--output", "appimage")
	cmd.Env = append(os.Environ(), "ARCH=x86_64", "APPIMAGE_EXTRACT_AND_RUN=1")
	cmd.Dir = staging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	matches, err := filepath.Glob(filepath.Join(staging, "VAdlp*.AppImage"))
	if err != nil || len(matches) == 0 {
		return fmt.Errorf("AppImage not produced in %s", staging)
	}
	return os.Rename(matches[0], archivePath)
}

func downloadURL(url, dest string) error {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
