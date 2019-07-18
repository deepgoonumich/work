package models

import "errors"

// Some common errors.
var (
	ErrDuplicate  = errors.New("duplicate event")
	ErrNotFound   = errors.New("not found")
	ErrNoRevision = errors.New("no such revision")
)
