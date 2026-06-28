package updater

import (
	"io"
	"os"
	"path/filepath"

	"vadlp/internal/configdir"
)

// MigrateToolsToConfigDir copies binaries from the executable directory into the
// permanent tools directory when they exist beside the app but not in config yet.
func MigrateToolsToConfigDir() error {
	toolsDir, err := configdir.ToolsDir()
	if err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return nil
	}
	exeDir := filepath.Dir(exe)
	for _, name := range []string{ytDlpBinName(), ffmpegBinName(), denoBinName()} {
		dst := filepath.Join(toolsDir, name)
		if _, err := os.Stat(dst); err == nil {
			continue
		}
		for _, src := range []string{
			filepath.Join(exeDir, name),
			filepath.Join(exeDir, "bin", name),
		} {
			if err := copyFileIfExists(src, dst); err != nil {
				return err
			}
			if _, err := os.Stat(dst); err == nil {
				break
			}
		}
	}
	return nil
}

func copyFileIfExists(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer in.Close()
	st, err := in.Stat()
	if err != nil {
		return err
	}
	if st.IsDir() {
		return nil
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
