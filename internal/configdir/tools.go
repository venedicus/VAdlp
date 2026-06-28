package configdir

import (
	"os"
	"path/filepath"
)

// ToolsDir returns the permanent directory for yt-dlp, ffmpeg, and deno binaries.
// Typically ~/.config/vadlp/tools (Linux), %AppData%\vadlp\tools (Windows).
func ToolsDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	tools := filepath.Join(dir, "tools")
	if err := os.MkdirAll(tools, 0o755); err != nil {
		return "", err
	}
	return tools, nil
}
