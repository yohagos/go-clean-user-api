package entity

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidEmail = errors.New("invalid email")
	ErrInvalidName = errors.New("invalid name")
	ErrEmailExists = errors.New("email already exists")
)