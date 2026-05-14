package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"elibrary/book-service/internal/cache"
	"elibrary/book-service/internal/domain"
	"elibrary/pkg/eventbus"
)

type BookUsecase struct {
	repo     domain.BookRepository
	copyRepo domain.CopyRepository
	cache    *cache.BookCache
	eventBus *eventbus.Publisher
}

func New(r domain.BookRepository, cr domain.CopyRepository, c *cache.BookCache, eb *eventbus.Publisher) *BookUsecase {
	return &BookUsecase{
		repo:     r,
		copyRepo: cr,
		cache:    c,
		eventBus: eb,
	}
}

func (uc *BookUsecase) Create(ctx context.Context, name string, authors []string, year int) (*domain.Book, error) {
	name = strings.TrimSpace(name)
	if name == "" || year <= 0 {
		return nil, domain.ErrInvalidArgument
	}
	b := &domain.Book{
		ID:        uuid.NewString(),
		Name:      name,
		Authors:   authors,
		Year:      year,
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
	}

	if err := uc.repo.Create(ctx, b); err != nil {
		return nil, err
	}

	// ПРОВЕРКА: Инвалидация кэша
	if uc.cache != nil {
		_ = uc.cache.InvalidateAllLists(ctx)
	}

	// ПРОВЕРКА: Публикация события
	if uc.eventBus != nil {
		_ = uc.eventBus.Publish(ctx, "book.created", eventbus.BookCreatedEvent{
			BookID: b.ID, Name: b.Name, Authors: strings.Join(b.Authors, ", "), Year: b.Year, CreatedAt: b.CreatedAt,
		})
	}

	return b, nil
}

func (uc *BookUsecase) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	// ПРОВЕРКА: Получение из кэша
	if uc.cache != nil {
		if b, hit, _ := uc.cache.GetBook(ctx, id); hit {
			return b, nil
		}
	}

	b, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// ПРОВЕРКА: Запись в кэш
	if uc.cache != nil {
		_ = uc.cache.SetBook(ctx, b)
	}

	return b, nil
}

func (uc *BookUsecase) Update(ctx context.Context, id, name string, authors []string, year int) (*domain.Book, error) {
	b, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		b.Name = name
	}
	if authors != nil {
		b.Authors = authors
	}
	if year > 0 {
		b.Year = year
	}

	if err := uc.repo.Update(ctx, b); err != nil {
		return nil, err
	}

	// ПРОВЕРКА: Инвалидация кэша
	if uc.cache != nil {
		_ = uc.cache.InvalidateBook(ctx, id)
		_ = uc.cache.InvalidateAllLists(ctx)
	}

	// ПРОВЕРКА: Событие
	if uc.eventBus != nil {
		_ = uc.eventBus.Publish(ctx, "book.updated", eventbus.BookUpdatedEvent{
			BookID: b.ID, Name: b.Name, Authors: strings.Join(b.Authors, ", "), Year: b.Year,
		})
	}

	return b, nil
}

func (uc *BookUsecase) Delete(ctx context.Context, id string) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	// ПРОВЕРКА: Инвалидация кэша
	if uc.cache != nil {
		_ = uc.cache.InvalidateBook(ctx, id)
		_ = uc.cache.InvalidateAllLists(ctx)
	}

	// ПРОВЕРКА: Событие
	if uc.eventBus != nil {
		_ = uc.eventBus.Publish(ctx, "book.deleted", eventbus.BookDeletedEvent{BookID: id})
	}
	return nil
}

func (uc *BookUsecase) List(ctx context.Context, limit, offset int) ([]*domain.Book, int, error) {
	if uc.cache != nil {
		if b, total, hit, _ := uc.cache.GetList(ctx, "list", limit, offset); hit {
			return b, total, nil
		}
	}
	books, total, err := uc.repo.List(ctx, limit, offset)
	if err == nil && uc.cache != nil {
		_ = uc.cache.SetList(ctx, "list", limit, offset, total, books)
	}
	return books, total, err
}

func (uc *BookUsecase) Search(ctx context.Context, query string, limit, offset int) ([]*domain.Book, int, error) {
	prefix := "search:" + query
	if uc.cache != nil {
		if b, total, hit, _ := uc.cache.GetList(ctx, prefix, limit, offset); hit {
			return b, total, nil
		}
	}
	books, total, err := uc.repo.Search(ctx, query, limit, offset)
	if err == nil && uc.cache != nil {
		_ = uc.cache.SetList(ctx, prefix, limit, offset, total, books)
	}
	return books, total, err
}

func (uc *BookUsecase) ListByAuthor(ctx context.Context, author string, limit, offset int) ([]*domain.Book, int, error) {
	prefix := "author:" + author
	if uc.cache != nil {
		if b, total, hit, _ := uc.cache.GetList(ctx, prefix, limit, offset); hit {
			return b, total, nil
		}
	}
	books, total, err := uc.repo.ListByAuthor(ctx, author, limit, offset)
	if err == nil && uc.cache != nil {
		_ = uc.cache.SetList(ctx, prefix, limit, offset, total, books)
	}
	return books, total, err
}

func (uc *BookUsecase) ListByYear(ctx context.Context, year, limit, offset int) ([]*domain.Book, int, error) {
	prefix := fmt.Sprintf("year:%d", year)
	if uc.cache != nil {
		if b, total, hit, _ := uc.cache.GetList(ctx, prefix, limit, offset); hit {
			return b, total, nil
		}
	}
	books, total, err := uc.repo.ListByYear(ctx, year, limit, offset)
	if err == nil && uc.cache != nil {
		_ = uc.cache.SetList(ctx, prefix, limit, offset, total, books)
	}
	return books, total, err
}

func (uc *BookUsecase) AddCopy(ctx context.Context, bookID string) (*domain.BookCopy, error) {
	if _, err := uc.repo.GetByID(ctx, bookID); err != nil {
		return nil, err
	}
	c := &domain.BookCopy{
		ExpID:  uuid.NewString(),
		BookID: bookID,
		Status: domain.CopyStatusAvailable,
	}
	if err := uc.copyRepo.Add(ctx, c); err != nil {
		return nil, err
	}
	if uc.cache != nil {
		_ = uc.cache.InvalidateAllLists(ctx)
	}
	return c, nil
}

func (uc *BookUsecase) UpdateCopyStatus(ctx context.Context, expID, status string) (*domain.BookCopy, error) {
	if !domain.IsValidCopyStatus(status) {
		return nil, domain.ErrInvalidStatus
	}
	oldCopy, err := uc.copyRepo.GetByID(ctx, expID)
	if err != nil {
		return nil, err
	}
	if err := uc.copyRepo.UpdateStatus(ctx, expID, status); err != nil {
		return nil, err
	}

	if uc.eventBus != nil {
		_ = uc.eventBus.Publish(ctx, "book.copy.status_changed", eventbus.CopyStatusChangedEvent{
			ExpID:     expID,
			BookID:    oldCopy.BookID,
			OldStatus: oldCopy.Status,
			NewStatus: status,
			ChangedAt: time.Now(),
		})
	}

	if uc.cache != nil {
		_ = uc.cache.InvalidateAllLists(ctx)
	}

	return uc.copyRepo.GetByID(ctx, expID)
}

func (uc *BookUsecase) GetBookCopies(ctx context.Context, bookID string) ([]*domain.BookCopy, error) {
	return uc.copyRepo.ListByBook(ctx, bookID)
}

func (uc *BookUsecase) GetAvailableCopies(ctx context.Context, bookID string) ([]*domain.BookCopy, error) {
	return uc.copyRepo.ListAvailable(ctx, bookID)
}
