package domain

import (
	"context"
	"time"
)

type Borrow struct {
	ID         string
	UserID     string
	BookID     string
	ExpID      string
	Barcode    string
	DateFrom   time.Time
	DateTo     time.Time
	ReturnedAt *time.Time
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

const (
	StatusActive    = "ACTIVE"
	StatusReturned  = "RETURNED"
	StatusOverdue   = "OVERDUE"
	StatusReserved  = "RESERVED"
	StatusCancelled = "CANCELLED"
)

type BorrowRepository interface {
	Create(ctx context.Context, b *Borrow) error
	GetByID(ctx context.Context, id string) (*Borrow, error)
	GetActiveByExpID(ctx context.Context, expID string) (*Borrow, error)
	Update(ctx context.Context, b *Borrow) error
	UpdateStatus(ctx context.Context, id, newStatus string) error
	List(ctx context.Context, limit, offset int) ([]*Borrow, int, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*Borrow, int, error)
	ListActive(ctx context.Context, limit, offset int) ([]*Borrow, int, error)
	ListOverdue(ctx context.Context, limit, offset int) ([]*Borrow, int, error)
}