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

// Обновленный интерфейс репозитория книг с поиском и фильтрами
type BookRepository interface {
	Create(ctx context.Context, b *Book) error
	GetByID(ctx context.Context, id string) (*Book, error)
	Update(ctx context.Context, b *Book) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Book, int, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*Book, int, error)
	ListByAuthor(ctx context.Context, author string, limit, offset int) ([]*Book, int, error)
	ListByYear(ctx context.Context, year, limit, offset int) ([]*Book, int, error)
}

// Обновленный интерфейс репозитория копий
type CopyRepository interface {
	Add(ctx context.Context, c *BookCopy) error
	ListByBook(ctx context.Context, bookID string) ([]*BookCopy, error)
	GetByID(ctx context.Context, expID string) (*BookCopy, error)
	UpdateStatus(ctx context.Context, expID, status string) error
	ListAvailable(ctx context.Context, bookID string) ([]*BookCopy, error)
}

const (
	CopyStatusAvailable = "AVAILABLE"
	CopyStatusBorrowed  = "BORROWED"
	CopyStatusReserved  = "RESERVED"
	CopyStatusLost      = "LOST"
	CopyStatusReturned  = "RETURNED"
)

// Проверка валидности статуса
func IsValidCopyStatus(s string) bool {
	switch s {
	case CopyStatusAvailable, CopyStatusBorrowed, CopyStatusReserved, CopyStatusLost, CopyStatusReturned:
		return true
	}
	return false
}
