package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"elibrary/borrow-service/internal/domain"
	"elibrary/borrow-service/internal/mailer"
	"elibrary/pkg/eventbus"
)

const defaultBorrowDays = 14

type BorrowUsecase struct {
	repo      domain.BorrowRepository
	users     domain.UserClient
	books     domain.BookClient
	mailer    *mailer.Mailer
	publisher *eventbus.Publisher
}

func NewBorrowUsecase(
	repo domain.BorrowRepository,
	users domain.UserClient,
	books domain.BookClient,
	m *mailer.Mailer,
	p *eventbus.Publisher,
) *BorrowUsecase {
	return &BorrowUsecase{
		repo:      repo,
		users:     users,
		books:     books,
		mailer:    m,
		publisher: p,
	}
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

	now := time.Now().UTC()
	b.Status = domain.StatusReturned
	b.ReturnedAt = &now

	if err := uc.repo.Update(ctx, b); err != nil {
		return nil, err
	}
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

func (uc *BorrowUsecase) BorrowSpecificCopy(ctx context.Context, userID, expID string, days int) (*domain.Borrow, error) {
	if userID == "" || expID == "" {
		return nil, domain.ErrInvalidArgument
	}
	if days <= 0 {
		days = defaultBorrowDays
	}

	if uc.users != nil {
		if err := uc.users.Exists(ctx, userID); err != nil {
			return nil, err
		}
	}

	if uc.books == nil {
		return nil, fmt.Errorf("book-service unavailable")
	}

	copy, err := uc.books.UpdateCopyStatus(ctx, expID, "BORROWED")
	if err != nil {
		return nil, fmt.Errorf("update copy status to BORROWED: %w", err)
	}

	now := time.Now().UTC()
	b := &domain.Borrow{
		ID:       uuid.NewString(),
		UserID:   userID,
		BookID:   copy.GetBookId(),
		ExpID:    expID,
		Barcode:  randomBarcode(),
		DateFrom: now,
		DateTo:   now.Add(time.Duration(days) * 24 * time.Hour),
		Status:   domain.StatusActive,
	}

	if err := uc.repo.Create(ctx, b); err != nil {
		if _, rerr := uc.books.UpdateCopyStatus(context.Background(), expID, "AVAILABLE"); rerr != nil {
			log.Printf("compensation failed for exp_id=%s: %v", expID, rerr)
		}
		return nil, err
	}

	uc.publishAsync("book.borrowed", eventbus.BookBorrowedEvent{
		BorrowID: b.ID,
		UserID:   b.UserID,
		BookID:   b.BookID,
		ExpID:    b.ExpID,
		DateFrom: b.DateFrom,
		DateTo:   b.DateTo,
	})

	uc.sendBorrowEmail(b)

	return b, nil
}

func (uc *BorrowUsecase) ReturnSpecificCopy(ctx context.Context, expID string) (*domain.Borrow, error) {
	if expID == "" {
		return nil, domain.ErrInvalidArgument
	}

	b, err := uc.repo.GetActiveByExpID(ctx, expID)
	if err != nil {
		return nil, err
	}

	if uc.books != nil {
		if _, err := uc.books.UpdateCopyStatus(ctx, expID, "AVAILABLE"); err != nil {
			return nil, fmt.Errorf("update copy status to AVAILABLE: %w", err)
		}
	}

	now := time.Now().UTC()
	b.Status = domain.StatusReturned
	b.ReturnedAt = &now

	if err := uc.repo.Update(ctx, b); err != nil {
		return nil, err
	}

	uc.publishAsync("book.returned", eventbus.BookReturnedEvent{
		BorrowID:   b.ID,
		UserID:     b.UserID,
		BookID:     b.BookID,
		ExpID:      b.ExpID,
		ReturnedAt: now,
	})

	uc.sendReturnEmail(b)

	return b, nil
}

func (uc *BorrowUsecase) ExtendBorrowPeriod(ctx context.Context, borrowID string, days int) (*domain.Borrow, error) {
	if borrowID == "" || days <= 0 {
		return nil, domain.ErrInvalidArgument
	}

	b, err := uc.repo.GetByID(ctx, borrowID)
	if err != nil {
		return nil, err
	}

	if b.Status != domain.StatusActive {
		return nil, fmt.Errorf("%w: only active borrows can be extended", domain.ErrInvalidArgument)
	}

	b.DateTo = b.DateTo.Add(time.Duration(days) * 24 * time.Hour)

	if err := uc.repo.Update(ctx, b); err != nil {
		return nil, err
	}

	return b, nil
}

func (uc *BorrowUsecase) GetOverdueBorrows(ctx context.Context, limit, offset int) ([]*domain.Borrow, int, error) {
	limit, offset = normalizePaging(limit, offset)
	return uc.repo.ListOverdue(ctx, limit, offset)
}

func (uc *BorrowUsecase) ReserveBookCopy(ctx context.Context, userID, expID string) (*domain.Borrow, error) {
	if userID == "" || expID == "" {
		return nil, domain.ErrInvalidArgument
	}

	if uc.users != nil {
		if err := uc.users.Exists(ctx, userID); err != nil {
			return nil, err
		}
	}

	if uc.books == nil {
		return nil, fmt.Errorf("book-service unavailable")
	}

	copy, err := uc.books.UpdateCopyStatus(ctx, expID, "RESERVED")
	if err != nil {
		return nil, fmt.Errorf("update copy status to RESERVED: %w", err)
	}

	now := time.Now().UTC()
	b := &domain.Borrow{
		ID:       uuid.NewString(),
		UserID:   userID,
		BookID:   copy.GetBookId(),
		ExpID:    expID,
		Barcode:  randomBarcode(),
		DateFrom: now,
		DateTo:   now.Add(7 * 24 * time.Hour),
		Status:   domain.StatusReserved,
	}

	if err := uc.repo.Create(ctx, b); err != nil {
		if _, rerr := uc.books.UpdateCopyStatus(context.Background(), expID, "AVAILABLE"); rerr != nil {
			log.Printf("compensation failed for exp_id=%s: %v", expID, rerr)
		}
		return nil, err
	}

	uc.publishAsync("book.reserved", eventbus.BookReservedEvent{
		BorrowID: b.ID,
		UserID:   b.UserID,
		BookID:   b.BookID,
		ExpID:    b.ExpID,
		At:       now,
	})

	uc.sendReservationEmail(b)

	return b, nil
}

func (uc *BorrowUsecase) CancelReservation(ctx context.Context, borrowID string) error {
	if borrowID == "" {
		return domain.ErrInvalidArgument
	}

	b, err := uc.repo.GetByID(ctx, borrowID)
	if err != nil {
		return err
	}

	if b.Status != domain.StatusReserved {
		return fmt.Errorf("%w: only reservations can be cancelled", domain.ErrInvalidArgument)
	}

	if uc.books != nil && b.ExpID != "" {
		if _, err := uc.books.UpdateCopyStatus(ctx, b.ExpID, "AVAILABLE"); err != nil {
			return fmt.Errorf("update copy status to AVAILABLE: %w", err)
		}
	}

	b.Status = domain.StatusCancelled

	if err := uc.repo.Update(ctx, b); err != nil {
		return err
	}

	uc.publishAsync("book.reservation_cancelled", eventbus.ReservationCancelledEvent{
		BorrowID: b.ID,
		UserID:   b.UserID,
		ExpID:    b.ExpID,
	})

	return nil
}

func (uc *BorrowUsecase) publishAsync(subject string, payload any) {
	if uc.publisher == nil {
		return
	}

	go func() {
		if err := uc.publisher.Publish(context.Background(), subject, payload); err != nil {
			log.Printf("publish %s: %v", subject, err)
		}
	}()
}

func (uc *BorrowUsecase) sendBorrowEmail(b *domain.Borrow) {
	if uc.mailer == nil || uc.users == nil || uc.books == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		u, err := uc.users.GetUser(ctx, b.UserID)
		if err != nil {
			log.Printf("get user for borrow email: %v", err)
			return
		}

		bk, err := uc.books.GetBook(ctx, b.BookID)
		if err != nil {
			log.Printf("get book for borrow email: %v", err)
			return
		}

		if err := uc.mailer.SendBorrowConfirmation(u.GetEmail(), u.GetName(), bk.GetName(), b.DateTo); err != nil {
			log.Printf("send borrow email: %v", err)
		}
	}()
}

func (uc *BorrowUsecase) sendReturnEmail(b *domain.Borrow) {
	if uc.mailer == nil || uc.users == nil || uc.books == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		u, err := uc.users.GetUser(ctx, b.UserID)
		if err != nil {
			log.Printf("get user for return email: %v", err)
			return
		}

		bk, err := uc.books.GetBook(ctx, b.BookID)
		if err != nil {
			log.Printf("get book for return email: %v", err)
			return
		}

		if err := uc.mailer.SendReturnConfirmation(u.GetEmail(), u.GetName(), bk.GetName()); err != nil {
			log.Printf("send return email: %v", err)
		}
	}()
}

func (uc *BorrowUsecase) sendReservationEmail(b *domain.Borrow) {
	if uc.mailer == nil || uc.users == nil || uc.books == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		u, err := uc.users.GetUser(ctx, b.UserID)
		if err != nil {
			log.Printf("get user for reservation email: %v", err)
			return
		}

		bk, err := uc.books.GetBook(ctx, b.BookID)
		if err != nil {
			log.Printf("get book for reservation email: %v", err)
			return
		}

		if err := uc.mailer.SendReservationConfirmation(u.GetEmail(), u.GetName(), bk.GetName()); err != nil {
			log.Printf("send reservation email: %v", err)
		}
	}()
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