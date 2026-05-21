package core

import "errors"

type ValidationError struct {
	Key string
}

func (e ValidationError) Error() string {
	return e.Key
}

func AsValidation(err error) (ValidationError, bool) {
	var v ValidationError
	if errors.As(err, &v) {
		return v, true
	}
	return ValidationError{}, false
}

var (
	ErrNoURL       = errors.New("no url")
	ErrNoBinary    = errors.New("yt-dlp not found")
	ErrInvalidPath = errors.New("invalid path")
)
