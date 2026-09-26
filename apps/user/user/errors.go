package user

import "errors"

var (
	ErrInvalidInput = errors.New("invalid user input")
	ErrNotFound     = errors.New("user not found")
	ErrEmailExists  = errors.New("email already exists")
)
