package model

import "errors"

// Errors _
var (
	ErrNotFound   = errors.New("not found")
	ErrGone       = errors.New("deleted")
	ErrBadRequest = errors.New("bad request")
	ErrNoContent  = errors.New("no content")
)
