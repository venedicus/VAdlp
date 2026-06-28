package executil

import "os/exec"

// Command returns an exec.Cmd with platform-specific flags to avoid visible console windows.
func Command(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	hideWindow(cmd)
	return cmd
}

// Output runs a hidden subprocess and returns its stdout.
func Output(name string, arg ...string) ([]byte, error) {
	return Command(name, arg...).Output()
}

// CombinedOutput runs a hidden subprocess and returns combined stdout/stderr.
func CombinedOutput(name string, arg ...string) ([]byte, error) {
	return Command(name, arg...).CombinedOutput()
}
