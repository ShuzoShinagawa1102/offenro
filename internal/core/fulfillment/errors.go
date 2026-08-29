package fulfillment

import "errors"

var (
	ErrInvalidInput           = errors.New("invalid fulfillment input")
	ErrNotFound               = errors.New("fulfillment resource not found")
	ErrConflict               = errors.New("fulfillment conflict")
	ErrMerchantRecordNotFound = errors.New("merchant fulfillment record not found")
)
