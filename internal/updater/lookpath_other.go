//go:build !windows

package updater

import "os/exec"

func lookPath(name string) (string, error) {
	return exec.LookPath(name)
}
