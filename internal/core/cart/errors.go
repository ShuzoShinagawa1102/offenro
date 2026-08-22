package cart

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("state conflict")
	ErrRevalidation = errors.New("merchant offer revalidation failed")
)
