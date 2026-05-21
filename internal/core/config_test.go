package core

import "testing"

func TestConfigNormalizeVerboseQuiet(t *testing.T) {
	c := Config{Verbose: true, Quiet: true}
	c.Normalize()
	if !c.Verbose || c.Quiet {
		t.Fatalf("expected verbose on, quiet off: %+v", c)
	}
}
