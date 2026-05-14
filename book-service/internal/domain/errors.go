package domain

import "errors"

var (
	ErrBookNotFound      = errors.New("book not found")
	ErrCopyNotFound      = errors.New("book copy not found")
	ErrBookAlreadyExists = errors.New("book already exists")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrInvalidStatus     = errors.New("invalid copy status") // Новая ошибка Day 2
)
