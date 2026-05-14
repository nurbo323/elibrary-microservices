package domain

import (
	"context"
	"time"
)

type User struct {
	ID                string
	Name              string
	Email             string
	PasswordHash      string
	EmailVerified     bool
	VerificationToken string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByVerificationToken(ctx context.Context, token string) (*User, error)
	Update(ctx context.Context, u *User) error
	UpdatePassword(ctx context.Context, id, newHash string) error
	MarkEmailVerified(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*User, int, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*User, int, error)
}
