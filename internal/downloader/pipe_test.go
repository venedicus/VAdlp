package downloader

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"vadlp/internal/executil"
)

// TestHelperProcess is not a real test: it is re-executed by the runCommand
// tests with VADLP_TEST_HELPER set to emulate a yt-dlp binary with
// pathological output. Modes:
//
//	stderr-burst  — writes ~4MiB to stderr (fills the pipe buffer) before any
//	                stdout output, then a single stdout line and exit 0.
//	stderr-forever — writes to stderr until killed.
func TestHelperProcess(t *testing.T) {
	mode := os.Getenv("VADLP_TEST_HELPER")
	if mode == "" {
		return
	}
	chunk := bytes.Repeat([]byte("x"), 64*1024)
	chunk = append(chunk, '\n')
	switch mode {
	case "stderr-burst":
		for i := 0; i < 64; i++ {
			if _, err := os.Stderr.Write(chunk); err != nil {
				os.Exit(1)
			}
		}
		fmt.Fprintln(os.Stdout, "[download] Destination: done.mp4")
		os.Exit(0)
	case "stderr-forever":
		for {
			if _, err := os.Stderr.Write(chunk); err != nil {
				os.Exit(0)
			}
		}
	}
	os.Exit(0)
}

func helperCommand(mode string) *exec.Cmd {
	cmd := executil.Command(os.Args[0], "-test.run=TestHelperProcess", "--", mode)
	cmd.Env = append(os.Environ(), "VADLP_TEST_HELPER="+mode)
	return cmd
}

// TestRunCommandStderrBurst reproduces the freeze: the child fills the stderr
// pipe while stdout stays idle. Serial pipe reading deadlocks here; runCommand
// must drain both pipes concurrently and complete.
func TestRunCommandStderrBurst(t *testing.T) {
	cmd := helperCommand("stderr-burst")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		logs, err := runCommand(ctx, "burst", cmd, nil)
		if err != nil {
			t.Errorf("runCommand error: %v", err)
		}
		if !strings.Contains(logs, "[download] Destination: done.mp4") {
			t.Errorf("stdout line missing from logs: %q", logs)
		}
	}()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		t.Fatal("runCommand deadlocked: stderr burst blocked while stdout was idle")
	}
}

// TestRunCommandCancelKillsStalledChild verifies that a child which never
// stops writing can still be cancelled: cancellation must kill the process
// instead of waiting for pipes that never drain.
func TestRunCommandCancelKillsStalledChild(t *testing.T) {
	cmd := helperCommand("stderr-forever")
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		_, err := runCommand(ctx, "forever", cmd, nil)
		done <- err
	}()

	time.Sleep(300 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != ErrCancelled {
			t.Errorf("expected ErrCancelled, got %v", err)
		}
	case <-time.After(10 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		t.Fatal("cancel did not unblock runCommand")
	}
}
