package auth

import "errors"

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrNotFound     = errors.New("not found")

	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInactiveUser       = errors.New("inactive user")
)
