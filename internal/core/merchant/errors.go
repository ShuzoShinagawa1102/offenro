package merchant

import "errors"

var (
	ErrInvalidInput = errors.New("invalid merchant management input")
	ErrNotFound     = errors.New("merchant management resource not found")
	ErrConflict     = errors.New("merchant management conflict")
	ErrVerification = errors.New("merchant capability verification failed")
)
