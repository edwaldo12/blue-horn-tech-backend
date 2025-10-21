package domain

import "errors"

var (
	// ErrNotFound indicates requested entity does not exist.
	ErrNotFound = errors.New("not found")

	// ErrInvalidStatusTransition signals a failed state change.
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	// ErrUnauthorized signals the requester is not permitted.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates the caller lacks permission.
	ErrForbidden = errors.New("forbidden")

	// ErrValidationFailure indicates invalid input payload.
	ErrValidationFailure = errors.New("validation failure")
)
