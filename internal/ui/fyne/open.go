package fyneui

import (
	"os/exec"
	"runtime"
)

func openFolder(path string) error {
	if path == "" {
		return nil
	}
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}
