package domain

import "errors"

var (
	ErrEstimationNotFound = errors.New("estimation not found")
	ErrAlreadySubmitted   = errors.New("estimation already submitted")
	ErrEmptyItems         = errors.New("estimation must have at least one item")
	ErrNotDraft           = errors.New("only draft estimations can be modified")
	ErrInvalidHours       = errors.New("hours must be non-negative and min <= likely <= max")
)
