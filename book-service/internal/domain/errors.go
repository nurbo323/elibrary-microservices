package domain

import "errors"

var (
	ErrBookNotFound      = errors.New("book not found")
	ErrCopyNotFound      = errors.New("book copy not found")
	ErrBookAlreadyExists = errors.New("book already exists")
	ErrInvalidArgument   = errors.New("invalid argument")
)

// Допустимые статусы экземпляров
const (
	CopyStatusAvailable = "AVAILABLE"
	CopyStatusBorrowed  = "BORROWED"
	CopyStatusReserved  = "RESERVED"
	CopyStatusLost      = "LOST"
	CopyStatusReturned  = "RETURNED"
)
