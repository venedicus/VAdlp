// Package instance tracks running VAdlp processes via small heartbeat files
// in the config dir, so a newly started instance can see siblings, tell which
// ones are idle vs actively downloading, and ask an idle one to quit or kill
// a specific PID outright. There is no enforced single-instance lock — running
// more than one copy is allowed, this just gives the user visibility/control.
package instance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"vadlp/internal/configdir"
)

// staleAfter is how long without a heartbeat before an instance file is
// considered abandoned (the process crashed without cleaning up after itself).
const staleAfter = 20 * time.Second

type Info struct {
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"startedAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Busy      bool      `json:"busy"`
}

func dir() (string, error) {
	base, err := configdir.Dir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(base, "instances")
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", err
	}
	return d, nil
}

func filePath(d string, pid int) string {
	return filepath.Join(d, strconv.Itoa(pid)+".json")
}

func quitMarkerPath(d string, pid int) string {
	return filepath.Join(d, strconv.Itoa(pid)+".quit")
}

// Register writes this process's heartbeat file. Call once at startup.
func Register() error {
	d, err := dir()
	if err != nil {
		return err
	}
	now := time.Now()
	info := Info{PID: os.Getpid(), StartedAt: now, UpdatedAt: now}
	return writeInfo(d, info)
}

// Heartbeat refreshes this process's busy flag and timestamp. Call periodically.
func Heartbeat(busy bool) error {
	d, err := dir()
	if err != nil {
		return err
	}
	info := Info{PID: os.Getpid(), UpdatedAt: time.Now(), Busy: busy}
	if existing, err := readInfo(filePath(d, os.Getpid())); err == nil {
		info.StartedAt = existing.StartedAt
	} else {
		info.StartedAt = info.UpdatedAt
	}
	return writeInfo(d, info)
}

// Unregister removes this process's heartbeat file. Call on clean shutdown.
func Unregister() {
	d, err := dir()
	if err != nil {
		return
	}
	_ = os.Remove(filePath(d, os.Getpid()))
	_ = os.Remove(quitMarkerPath(d, os.Getpid()))
}

// List returns other live instances (excludes self, drops and cleans up stale files).
func List() ([]Info, error) {
	d, err := dir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(d)
	if err != nil {
		return nil, err
	}
	self := os.Getpid()
	out := []Info{}
	now := time.Now()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		pid, convErr := strconv.Atoi(strings.TrimSuffix(name, ".json"))
		if convErr != nil {
			continue
		}
		full := filepath.Join(d, name)
		info, readErr := readInfo(full)
		if readErr != nil {
			_ = os.Remove(full)
			continue
		}
		if now.Sub(info.UpdatedAt) > staleAfter {
			_ = os.Remove(full)
			_ = os.Remove(quitMarkerPath(d, pid))
			continue
		}
		if pid == self {
			continue
		}
		out = append(out, info)
	}
	return out, nil
}

// RequestQuit asks another instance to quit gracefully (it must be polling ShouldQuit).
func RequestQuit(pid int) error {
	d, err := dir()
	if err != nil {
		return err
	}
	return os.WriteFile(quitMarkerPath(d, pid), []byte("quit"), 0o644)
}

// ShouldQuit reports (and consumes) a pending quit request for this process.
func ShouldQuit() bool {
	d, err := dir()
	if err != nil {
		return false
	}
	p := quitMarkerPath(d, os.Getpid())
	if _, err := os.Stat(p); err != nil {
		return false
	}
	_ = os.Remove(p)
	return true
}

// Kill force-terminates another VAdlp process by PID.
func Kill(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

func writeInfo(d string, info Info) error {
	b, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath(d, info.PID), b, 0o644)
}

func readInfo(path string) (Info, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Info{}, err
	}
	var info Info
	if err := json.Unmarshal(b, &info); err != nil {
		return Info{}, err
	}
	return info, nil
}
