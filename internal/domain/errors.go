package domain

import "errors"

var (
	ErrValidation           = errors.New("validation error")
	ErrNotFound             = errors.New("not found")
	ErrInvalidTransition    = errors.New("invalid transition")
	ErrConflict             = errors.New("conflict")
	ErrAssistantUnavailable = errors.New("assistant unavailable")
)
