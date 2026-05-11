package domain

import "errors"

var (
	ErrBorrowNotFound  = errors.New("borrow not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrBookNotFound    = errors.New("book not found")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotActive       = errors.New("borrow is not active")
)