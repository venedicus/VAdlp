package updater

import (
	"path/filepath"
	"strings"
)

func localToolPath(name string) string {
	return filepath.Join(DefaultInstallDir(), name)
}

func ProbeLocalFFmpeg(customPath string) ToolStatus {
	if p := strings.TrimSpace(customPath); p != "" {
		return toolStatusFromResolved(resolveTool(ffmpegBinName(), "-version", p))
	}
	return probeExact(localToolPath(ffmpegBinName()), "-version")
}

func ProbeLocalDeno(customPath string) ToolStatus {
	if p := strings.TrimSpace(customPath); p != "" {
		return toolStatusFromResolved(resolveTool(denoBinName(), "--version", p))
	}
	return probeExact(localToolPath(denoBinName()), "--version")
}
