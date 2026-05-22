package updater

import "path/filepath"

func localToolPath(name string) string {
	return filepath.Join(DefaultInstallDir(), name)
}

func ProbeLocalFFmpeg() ToolStatus {
	return probeExact(localToolPath(ffmpegBinName()), "-version")
}

func ProbeLocalDeno() ToolStatus {
	return probeExact(localToolPath(denoBinName()), "-version")
}
