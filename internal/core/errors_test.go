package core

import (
	"errors"
	"testing"
)

func TestValidationError(t *testing.T) {
	err := ValidationError{Key: "err.test"}
	if err.Error() != "err.test" {
		t.Fatalf("got %q", err.Error())
	}
	wrapped := errors.New("wrap")
	if _, ok := AsValidation(wrapped); ok {
		t.Fatal("expected false")
	}
	if v, ok := AsValidation(err); !ok || v.Key != "err.test" {
		t.Fatalf("got %v ok=%v", v, ok)
	}
}
