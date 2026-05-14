package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"elibrary/borrow-service/internal/domain"
)

type mockRepo struct {
	borrows map[string]*domain.Borrow
}

func newMockRepo() *mockRepo {
	return &mockRepo{borrows: map[string]*domain.Borrow{}}
}

func (m *mockRepo) Create(_ context.Context, b *domain.Borrow) error {
	m.borrows[b.ID] = b
	return nil
}

func (m *mockRepo) GetByID(_ context.Context, id string) (*domain.Borrow, error) {
	b, ok := m.borrows[id]
	if !ok {
		return nil, domain.ErrBorrowNotFound
	}
	return b, nil
}

func (m *mockRepo) GetActiveByExpID(_ context.Context, expID string) (*domain.Borrow, error) {
	for _, b := range m.borrows {
		if b.ExpID == expID && (b.Status == domain.StatusActive || b.Status == domain.StatusReserved) {
			return b, nil
		}
	}
	return nil, domain.ErrBorrowNotFound
}

func (m *mockRepo) Update(_ context.Context, b *domain.Borrow) error {
	if _, ok := m.borrows[b.ID]; !ok {
		return domain.ErrBorrowNotFound
	}
	m.borrows[b.ID] = b
	return nil
}

func (m *mockRepo) UpdateStatus(_ context.Context, id, newStatus string) error {
	b, ok := m.borrows[id]
	if !ok {
		return domain.ErrBorrowNotFound
	}
	b.Status = newStatus
	return nil
}

func (m *mockRepo) List(_ context.Context, limit, offset int) ([]*domain.Borrow, int, error) {
	out := make([]*domain.Borrow, 0, len(m.borrows))
	for _, b := range m.borrows {
		out = append(out, b)
	}
	return out, len(out), nil
}

func (m *mockRepo) ListByUser(_ context.Context, userID string, limit, offset int) ([]*domain.Borrow, int, error) {
	out := make([]*domain.Borrow, 0)
	for _, b := range m.borrows {
		if b.UserID == userID {
			out = append(out, b)
		}
	}
	return out, len(out), nil
}

func (m *mockRepo) ListActive(_ context.Context, limit, offset int) ([]*domain.Borrow, int, error) {
	out := make([]*domain.Borrow, 0)
	for _, b := range m.borrows {
		if b.Status == domain.StatusActive {
			out = append(out, b)
		}
	}
	return out, len(out), nil
}

func (m *mockRepo) ListOverdue(_ context.Context, limit, offset int) ([]*domain.Borrow, int, error) {
	out := make([]*domain.Borrow, 0)
	now := time.Now()

	for _, b := range m.borrows {
		if b.Status == domain.StatusActive && b.DateTo.Before(now) {
			out = append(out, b)
		}
	}

	return out, len(out), nil
}

func TestExtendBorrowPeriod_OK(t *testing.T) {
	repo := newMockRepo()

	originalDateTo := time.Now().UTC().Add(24 * time.Hour)
	repo.borrows["b1"] = &domain.Borrow{
		ID:       "b1",
		Status:   domain.StatusActive,
		DateFrom: time.Now().UTC(),
		DateTo:   originalDateTo,
	}

	uc := NewBorrowUsecase(repo, nil, nil, nil, nil)

	b, err := uc.ExtendBorrowPeriod(context.Background(), "b1", 7)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedDateTo := originalDateTo.Add(7 * 24 * time.Hour)
	if !b.DateTo.Equal(expectedDateTo) {
		t.Fatalf("expected date_to %v, got %v", expectedDateTo, b.DateTo)
	}
}

func TestExtendBorrowPeriod_NotActive(t *testing.T) {
	repo := newMockRepo()

	repo.borrows["b1"] = &domain.Borrow{
		ID:     "b1",
		Status: domain.StatusReturned,
		DateTo: time.Now().UTC(),
	}

	uc := NewBorrowUsecase(repo, nil, nil, nil, nil)

	_, err := uc.ExtendBorrowPeriod(context.Background(), "b1", 7)
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestGetOverdueBorrows(t *testing.T) {
	repo := newMockRepo()

	repo.borrows["overdue"] = &domain.Borrow{
		ID:       "overdue",
		Status:   domain.StatusActive,
		DateFrom: time.Now().UTC().Add(-30 * 24 * time.Hour),
		DateTo:   time.Now().UTC().Add(-24 * time.Hour),
	}

	repo.borrows["active"] = &domain.Borrow{
		ID:       "active",
		Status:   domain.StatusActive,
		DateFrom: time.Now().UTC(),
		DateTo:   time.Now().UTC().Add(7 * 24 * time.Hour),
	}

	repo.borrows["returned"] = &domain.Borrow{
		ID:       "returned",
		Status:   domain.StatusReturned,
		DateFrom: time.Now().UTC().Add(-30 * 24 * time.Hour),
		DateTo:   time.Now().UTC().Add(-24 * time.Hour),
	}

	uc := NewBorrowUsecase(repo, nil, nil, nil, nil)

	items, total, err := uc.GetOverdueBorrows(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].ID != "overdue" {
		t.Fatalf("expected overdue borrow, got %s", items[0].ID)
	}
}

func TestCancelReservation_OK(t *testing.T) {
	repo := newMockRepo()

	repo.borrows["reservation"] = &domain.Borrow{
		ID:     "reservation",
		UserID: "u1",
		ExpID:  "exp1",
		Status: domain.StatusReserved,
	}

	uc := NewBorrowUsecase(repo, nil, nil, nil, nil)

	err := uc.CancelReservation(context.Background(), "reservation")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := repo.borrows["reservation"]
	if got.Status != domain.StatusCancelled {
		t.Fatalf("expected status CANCELLED, got %s", got.Status)
	}
}

func TestCancelReservation_NotReserved(t *testing.T) {
	repo := newMockRepo()

	repo.borrows["b1"] = &domain.Borrow{
		ID:     "b1",
		Status: domain.StatusActive,
	}

	uc := NewBorrowUsecase(repo, nil, nil, nil, nil)

	err := uc.CancelReservation(context.Background(), "b1")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestReturnSpecificCopy_OK(t *testing.T) {
	repo := newMockRepo()

	repo.borrows["b1"] = &domain.Borrow{
		ID:       "b1",
		UserID:   "u1",
		BookID:   "book1",
		ExpID:    "exp1",
		Status:   domain.StatusActive,
		DateFrom: time.Now().UTC().Add(-3 * 24 * time.Hour),
		DateTo:   time.Now().UTC().Add(11 * 24 * time.Hour),
	}

	uc := NewBorrowUsecase(repo, nil, nil, nil, nil)

	b, err := uc.ReturnSpecificCopy(context.Background(), "exp1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if b.Status != domain.StatusReturned {
		t.Fatalf("expected status RETURNED, got %s", b.Status)
	}

	if b.ReturnedAt == nil {
		t.Fatal("expected returned_at to be set")
	}
}