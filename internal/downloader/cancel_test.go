package downloader

import "testing"

func TestCancelJob(t *testing.T) {
	id := "test-cancel-job"
	killed := false
	registerJob(id, func() error {
		killed = true
		return nil
	})
	defer unregisterJob(id)

	if !CancelJob(id) {
		t.Fatal("expected cancel true")
	}
	if !killed || !jobCancelled(id) {
		t.Fatal("job not cancelled")
	}
	if CancelJob("missing") {
		t.Fatal("expected false for unknown job")
	}
}

func TestCancelAll(t *testing.T) {
	registerJob("a", nil)
	registerJob("b", nil)
	defer func() {
		unregisterJob("a")
		unregisterJob("b")
	}()
	CancelAll()
	if !jobCancelled("a") || !jobCancelled("b") {
		t.Fatal("expected all cancelled")
	}
}

func TestIsCancelled(t *testing.T) {
	if !IsCancelled(ErrCancelled) {
		t.Fatal("expected cancelled")
	}
	if IsCancelled(nil) {
		t.Fatal("nil is not cancelled")
	}
}
