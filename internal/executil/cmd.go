package executil

import (
	"context"
	"os/exec"
)

// Command returns an exec.Cmd with platform-specific flags to avoid visible console windows.
func Command(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	hideWindow(cmd)
	return cmd
}

// CommandContext returns an exec.Cmd bound to ctx, with platform-specific flags
// to avoid visible console windows.
func CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, arg...)
	hideWindow(cmd)
	return cmd
}

// Output runs a hidden subprocess and returns its stdout.
func Output(name string, arg ...string) ([]byte, error) {
	return Command(name, arg...).Output()
}

// OutputContext runs a hidden subprocess bound to ctx and returns its stdout.
func OutputContext(ctx context.Context, name string, arg ...string) ([]byte, error) {
	return CommandContext(ctx, name, arg...).Output()
}

// CombinedOutput runs a hidden subprocess and returns combined stdout/stderr.
func CombinedOutput(name string, arg ...string) ([]byte, error) {
	return Command(name, arg...).CombinedOutput()
}
