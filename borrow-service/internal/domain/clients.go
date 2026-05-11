package domain

import "context"

// UserClient — внешний контракт для проверки существования пользователя.
// Реализация в internal/client/user_client.go.
type UserClient interface {
	Exists(ctx context.Context, userID string) error // returns ErrUserNotFound if not found
}

// BookClient — внешний контракт для проверки существования книги.
type BookClient interface {
	Exists(ctx context.Context, bookID string) error // returns ErrBookNotFound if not found
}