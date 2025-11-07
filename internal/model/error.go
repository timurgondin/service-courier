package model

import "errors"

var (
	ErrCourierNotFound      = errors.New("courier not found")
	ErrCourierStatusInvalid = errors.New("courier status must be one of: paused, available, busy")
	ErrCourierNameEmpty     = errors.New("courier name is required")
	ErrCourierPhoneEmpty    = errors.New("courier phone is required")
	ErrCourierPhoneInvalid  = errors.New("courier phone must start with '+', have 12 characters, and contain only digits after '+'")
	ErrCourierStatusEmpty   = errors.New("courier status is required")
	ErrPhoneExists          = errors.New("profile with this phone already exists")
)
