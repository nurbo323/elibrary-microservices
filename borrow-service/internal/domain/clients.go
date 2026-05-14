package domain

import (
	"context"

	"elibrary/gen/bookpb"
	"elibrary/gen/userpb"
)

type UserClient interface {
	Exists(ctx context.Context, userID string) error
	GetUser(ctx context.Context, userID string) (*userpb.User, error)
}

type BookClient interface {
	Exists(ctx context.Context, bookID string) error
	GetBook(ctx context.Context, bookID string) (*bookpb.Book, error)
	UpdateCopyStatus(ctx context.Context, expID, status string) (*bookpb.BookCopy, error)
}