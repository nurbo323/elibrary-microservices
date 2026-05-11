package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/google/uuid"

	"elibrary/borrow-service/internal/domain"
)

const defaultBorrowDays = 14

type BorrowUsecase struct {
	repo  domain.BorrowRepository
	users domain.UserClient
	books domain.BookClient
}

func NewBorrowUsecase(repo domain.BorrowRepository, users domain.UserClient, books domain.BookClient) *BorrowUsecase {
	return &BorrowUsecase{repo: repo, users: users, books: books}
}

func (uc *BorrowUsecase) Borrow(ctx context.Context, userID, bookID string) (*domain.Borrow, error) {
	if userID == "" || bookID == "" {
		return nil, fmt.Errorf("%w: user_id and book_id required", domain.ErrInvalidArgument)
	}
	if err := uc.users.Exists(ctx, userID); err != nil {
		return nil, err
	}
	if err := uc.books.Exists(ctx, bookID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	b := &domain.Borrow{
		ID:       uuid.NewString(),
		UserID:   userID,
		BookID:   bookID,
		Barcode:  randomBarcode(),
		DateFrom: now,
		DateTo:   now.Add(defaultBorrowDays * 24 * time.Hour),
		Status:   domain.StatusActive,
	}
	if err := uc.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (uc *BorrowUsecase) Return(ctx context.Context, borrowID string) (*domain.Borrow, error) {
	if borrowID == "" {
		return nil, domain.ErrInvalidArgument
	}
	b, err := uc.repo.GetByID(ctx, borrowID)
	if err != nil {
		return nil, err
	}
	if b.Status != domain.StatusActive {
		return nil, fmt.Errorf("%w: current status is %s", domain.ErrNotActive, b.Status)
	}
	if err := uc.repo.UpdateStatus(ctx, borrowID, domain.StatusReturned); err != nil {
		return nil, err
	}
	b.Status = domain.StatusReturned
	return b, nil
}

func (uc *BorrowUsecase) GetByID(ctx context.Context, id string) (*domain.Borrow, error) {
	if id == "" {
		return nil, domain.ErrInvalidArgument
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *BorrowUsecase) List(ctx context.Context, limit, offset int) ([]*domain.Borrow, int, error) {
	limit, offset = normalizePaging(limit, offset)
	return uc.repo.List(ctx, limit, offset)
}

func (uc *BorrowUsecase) UserHistory(ctx context.Context, userID string, limit, offset int) ([]*domain.Borrow, int, error) {
	if userID == "" {
		return nil, 0, domain.ErrInvalidArgument
	}
	limit, offset = normalizePaging(limit, offset)
	return uc.repo.ListByUser(ctx, userID, limit, offset)
}

func (uc *BorrowUsecase) Active(ctx context.Context, limit, offset int) ([]*domain.Borrow, int, error) {
	limit, offset = normalizePaging(limit, offset)
	return uc.repo.ListActive(ctx, limit, offset)
}

func normalizePaging(limit, offset int) (int, int) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func randomBarcode() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%X", b)
}