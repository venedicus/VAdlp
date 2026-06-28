package instance

import (
	"os"
	"testing"
)

func TestRegisterListUnregister(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())

	if err := Register(); err != nil {
		t.Fatalf("Register: %v", err)
	}
	defer Unregister()

	if err := Heartbeat(true); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}

	// A real second process isn't spawned here, so List() should not include self.
	others, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, o := range others {
		if o.PID == os.Getpid() {
			t.Fatalf("List() should not include self, got %+v", o)
		}
	}
}

func TestQuitMarkerRoundtrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())

	if err := Register(); err != nil {
		t.Fatalf("Register: %v", err)
	}
	defer Unregister()

	if ShouldQuit() {
		t.Fatal("ShouldQuit() should be false before any request")
	}
	if err := RequestQuit(os.Getpid()); err != nil {
		t.Fatalf("RequestQuit: %v", err)
	}
	if !ShouldQuit() {
		t.Fatal("ShouldQuit() should be true after RequestQuit")
	}
	if ShouldQuit() {
		t.Fatal("ShouldQuit() should consume the marker (false on second call)")
	}
}
