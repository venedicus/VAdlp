package downloader

import (
	"errors"
	"sync"
)

var ErrCancelled = errors.New("download cancelled")

type job struct {
	cancelled bool
	kill      func() error
}

var (
	jobsMu sync.Mutex
	jobs   = map[string]*job{}
)

func CancelJob(id string) bool {
	jobsMu.Lock()
	j, ok := jobs[id]
	jobsMu.Unlock()
	if !ok {
		return false
	}
	j.cancelled = true
	if j.kill != nil {
		_ = j.kill()
	}
	return true
}

func CancelAll() {
	jobsMu.Lock()
	ids := make([]string, 0, len(jobs))
	for id := range jobs {
		ids = append(ids, id)
	}
	jobsMu.Unlock()
	for _, id := range ids {
		CancelJob(id)
	}
}

func IsCancelled(err error) bool {
	return errors.Is(err, ErrCancelled)
}

func registerJob(id string, kill func() error) {
	jobsMu.Lock()
	jobs[id] = &job{kill: kill}
	jobsMu.Unlock()
}

func unregisterJob(id string) {
	jobsMu.Lock()
	delete(jobs, id)
	jobsMu.Unlock()
}

func jobCancelled(id string) bool {
	jobsMu.Lock()
	j, ok := jobs[id]
	cancelled := ok && j.cancelled
	jobsMu.Unlock()
	return cancelled
}
