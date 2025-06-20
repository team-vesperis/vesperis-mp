package mp

import "errors"

// error returns
var (
	ErrPlayerNotFound = errors.New("player not found")
	ErrSuccessful     = errors.New("successful")
)
