package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"elibrary/book-service/internal/domain"
)

type BookUsecase struct {
	books  domain.BookRepository
	copies domain.BookCopyRepository
}

func NewBookUsecase(books domain.BookRepository, copies domain.BookCopyRepository) *BookUsecase {
	return &BookUsecase{books: books, copies: copies}
}

func (uc *BookUsecase) Create(ctx context.Context, name string, authors []string, year int) (*domain.Book, error) {
	name = strings.TrimSpace(name)
	if name == "" || year <= 0 {
		return nil, fmt.Errorf("%w: name and positive year required", domain.ErrInvalidArgument)
	}
	if authors == nil {
		authors = []string{}
	}
	b := &domain.Book{
		ID:      uuid.NewString(),
		Name:    name,
		Authors: authors,
		Year:    year,
		Status:  "ACTIVE",
	}
	if err := uc.books.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (uc *BookUsecase) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	if id == "" {
		return nil, domain.ErrInvalidArgument
	}
	return uc.books.GetByID(ctx, id)
}

func (uc *BookUsecase) Update(ctx context.Context, id, name string, authors []string, year int) (*domain.Book, error) {
	if id == "" {
		return nil, domain.ErrInvalidArgument
	}
	b, err := uc.books.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name = strings.TrimSpace(name); name != "" {
		b.Name = name
	}
	if authors != nil {
		b.Authors = authors
	}
	if year > 0 {
		b.Year = year
	}
	if err := uc.books.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (uc *BookUsecase) Delete(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrInvalidArgument
	}
	return uc.books.Delete(ctx, id)
}

func (uc *BookUsecase) List(ctx context.Context, limit, offset int) ([]*domain.Book, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.books.List(ctx, limit, offset)
}

func (uc *BookUsecase) AddCopy(ctx context.Context, bookID string) (*domain.BookCopy, error) {
	if bookID == "" {
		return nil, domain.ErrInvalidArgument
	}
	// Проверим что книга существует — выдаст красивую 404, если нет
	if _, err := uc.books.GetByID(ctx, bookID); err != nil {
		return nil, err
	}
	c := &domain.BookCopy{
		ExpID:  uuid.NewString(),
		BookID: bookID,
		Status: domain.CopyStatusAvailable,
	}
	if err := uc.copies.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}
