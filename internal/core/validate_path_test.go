package core

import "testing"

func TestValidatePathWindowsDrive(t *testing.T) {
	paths := []string{
		`C:\Users\veno\Videos`,
		`C:/Users/test/Downloads`,
		`\\server\share\folder`,
	}
	for _, p := range paths {
		if err := ValidatePath(p); err != nil {
			t.Fatalf("%q: %v", p, err)
		}
	}
}

func TestValidatePathRejectsWildcards(t *testing.T) {
	if err := ValidatePath(`C:\bad|dir`); err == nil {
		t.Fatal("expected error")
	}
}
