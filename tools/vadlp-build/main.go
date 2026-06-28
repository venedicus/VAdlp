// Cross-platform CI/local build helper for Wails + React.
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	goos := os.Getenv("GOOS")
	goarch := os.Getenv("GOARCH")
	binary := os.Getenv("BINARY")
	version := os.Getenv("VERSION")
	commit := os.Getenv("COMMIT")
	date := os.Getenv("DATE")

	for name, val := range map[string]string{
		"GOOS": goos, "GOARCH": goarch, "BINARY": binary,
		"VERSION": version, "COMMIT": commit, "DATE": date,
	} {
		if val == "" {
			fmt.Fprintf(os.Stderr, "%s required\n", name)
			os.Exit(1)
		}
	}

	ldflags := strings.Join([]string{
		"-X", "vadlp/internal/version.Version=" + version,
		"-X", "vadlp/internal/version.Commit=" + commit,
		"-X", "vadlp/internal/version.BuildDate=" + date,
	}, " ")

	args := []string{
		"run", "github.com/wailsapp/wails/v2/cmd/wails@latest", "build",
		"-platform", goos + "/" + goarch,
		"-ldflags", ldflags,
		"-o", binary,
	}
	if goos != "darwin" {
		// On darwin we want Wails to produce a real .app bundle (Info.plist,
		// icon, Contents/MacOS/<binary>) so it can be packaged into a proper
		// .dmg instead of shipping a bare, unsigned Mach-O executable.
		args = append(args, "-nopackage")
	}
	if goos == "linux" {
		// Modern distros (e.g. Ubuntu 24.04+) only ship webkit2gtk-4.1, not
		// the 4.0 pkg-config name Wails defaults to; this tag switches it.
		args = append(args, "-tags", "webkit2_41")
	}
	wails := exec.Command("go", args...)
	wails.Env = append(os.Environ(),
		"CGO_ENABLED=1",
		"GOOS="+goos,
		"GOARCH="+goarch,
	)
	if os.Getenv("RUNNER_OS") == "Windows" {
		wails.Env = append(wails.Env, "PATH=/mingw64/bin:/usr/bin:"+os.Getenv("PATH"))
	}
	wails.Stdout = os.Stdout
	wails.Stderr = os.Stderr
	if err := wails.Run(); err != nil {
		os.Exit(1)
	}

	if goos == "darwin" {
		bundle, err := findAppBundle(filepath.Join("build", "bin"))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// Extract the flat executable from inside the bundle too, for
		// callers that just want a CLI-style binary (e.g. the tar.gz).
		if err := copyFile(filepath.Join(bundle, "Contents", "MacOS", binary), binary); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	built := filepath.Join("build", "bin", binary)
	if err := copyFile(built, binary); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// findAppBundle locates the single .app bundle Wails produced in dir.
func findAppBundle(dir string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.app"))
	if err != nil {
		return "", err
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("expected exactly one .app bundle in %s, found %d", dir, len(matches))
	}
	return matches[0], nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("copy from %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("copy to %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
