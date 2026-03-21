package domain

import "errors"

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrInvalidChannel       = errors.New("invalid notification channel")
	ErrDeliveryFailed       = errors.New("notification delivery failed")
)
