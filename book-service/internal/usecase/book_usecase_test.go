package usecase

import (
	"context"
	"errors"
	"testing"

	"elibrary/book-service/internal/domain"
)

// --- Mocks ---

type mockBookRepo struct {
	books      map[string]*domain.Book
	failCreate bool
}

func newMockBookRepo() *mockBookRepo {
	return &mockBookRepo{books: map[string]*domain.Book{}}
}

func (m *mockBookRepo) Create(_ context.Context, b *domain.Book) error {
	if m.failCreate {
		return domain.ErrBookAlreadyExists
	}
	m.books[b.ID] = b
	return nil
}
func (m *mockBookRepo) GetByID(_ context.Context, id string) (*domain.Book, error) {
	if b, ok := m.books[id]; ok {
		return b, nil
	}
	return nil, domain.ErrBookNotFound
}
func (m *mockBookRepo) Update(_ context.Context, b *domain.Book) error { m.books[b.ID] = b; return nil }
func (m *mockBookRepo) Delete(_ context.Context, id string) error      { delete(m.books, id); return nil }
func (m *mockBookRepo) List(_ context.Context, _, _ int) ([]*domain.Book, int, error) {
	return nil, 0, nil
}
func (m *mockBookRepo) Search(_ context.Context, _ string, _, _ int) ([]*domain.Book, int, error) {
	return nil, 0, nil
}
func (m *mockBookRepo) ListByAuthor(_ context.Context, _ string, _, _ int) ([]*domain.Book, int, error) {
	return nil, 0, nil
}
func (m *mockBookRepo) ListByYear(_ context.Context, _, _, _ int) ([]*domain.Book, int, error) {
	return nil, 0, nil
}

type mockCopyRepo struct {
	copies map[string]*domain.BookCopy
}

func newMockCopyRepo() *mockCopyRepo {
	return &mockCopyRepo{copies: map[string]*domain.BookCopy{}}
}

func (m *mockCopyRepo) Add(_ context.Context, c *domain.BookCopy) error {
	m.copies[c.ExpID] = c
	return nil
}
func (m *mockCopyRepo) ListByBook(_ context.Context, bid string) ([]*domain.BookCopy, error) {
	out := []*domain.BookCopy{}
	for _, c := range m.copies {
		if c.BookID == bid {
			out = append(out, c)
		}
	}
	return out, nil
}
func (m *mockCopyRepo) GetByID(_ context.Context, id string) (*domain.BookCopy, error) {
	if c, ok := m.copies[id]; ok {
		return c, nil
	}
	return nil, domain.ErrCopyNotFound
}
func (m *mockCopyRepo) UpdateStatus(_ context.Context, id, s string) error {
	c, ok := m.copies[id]
	if !ok {
		return domain.ErrCopyNotFound
	}
	c.Status = s
	return nil
}
func (m *mockCopyRepo) ListAvailable(_ context.Context, bid string) ([]*domain.BookCopy, error) {
	out := []*domain.BookCopy{}
	for _, c := range m.copies {
		if c.Status == domain.CopyStatusAvailable && (bid == "" || c.BookID == bid) {
			out = append(out, c)
		}
	}
	return out, nil
}

// --- Tests ---

func TestCreate_OK(t *testing.T) {
	uc := New(newMockBookRepo(), newMockCopyRepo(), nil, nil)
	b, err := uc.Create(context.Background(), "Go Programming", []string{"Alan Donovan"}, 2015)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if b.ID == "" {
		t.Fatalf("empty id")
	}
}

func TestCreate_InvalidArgs(t *testing.T) {
	uc := New(newMockBookRepo(), newMockCopyRepo(), nil, nil)
	_, err := uc.Create(context.Background(), "", nil, 0)
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("want ErrInvalidArgument, got %v", err)
	}
}

func TestUpdateCopyStatus_OK(t *testing.T) {
	copies := newMockCopyRepo()
	copies.copies["exp-1"] = &domain.BookCopy{
		ExpID: "exp-1", BookID: "b-1", Status: domain.CopyStatusAvailable,
	}
	uc := New(newMockBookRepo(), copies, nil, nil)
	c, err := uc.UpdateCopyStatus(context.Background(), "exp-1", domain.CopyStatusBorrowed)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if c.Status != domain.CopyStatusBorrowed {
		t.Fatalf("status not updated: %s", c.Status)
	}
}
