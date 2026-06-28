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

	wails := exec.Command(
		"go", "run", "github.com/wailsapp/wails/v2/cmd/wails@latest", "build",
		"-platform", goos+"/"+goarch,
		"-ldflags", ldflags,
		"-nopackage",
		"-o", binary,
	)
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

	built := filepath.Join("build", "bin", binary)
	if err := copyFile(built, binary); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
