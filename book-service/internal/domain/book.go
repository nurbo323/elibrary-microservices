package domain

import (
	"context"
	"time"
)

type Book struct {
	ID        string
	Name      string
	Authors   []string
	Year      int
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BookCopy struct {
	ExpID     string
	BookID    string
	Status    string // AVAILABLE / BORROWED / RESERVED / LOST / RETURNED
	CreatedAt time.Time
}

type BookRepository interface {
	Create(ctx context.Context, b *Book) error
	GetByID(ctx context.Context, id string) (*Book, error)
	Update(ctx context.Context, b *Book) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Book, int, error)
}

type BookCopyRepository interface {
	Create(ctx context.Context, c *BookCopy) error
	GetByID(ctx context.Context, id string) (*BookCopy, error)
	ListByBook(ctx context.Context, bookID string) ([]*BookCopy, error)
}
