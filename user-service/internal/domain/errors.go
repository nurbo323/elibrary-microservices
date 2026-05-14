package domain

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid email or password")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrInvalidToken      = errors.New("invalid or expired verification token")
	ErrSamePassword      = errors.New("new password must differ from old")
)
