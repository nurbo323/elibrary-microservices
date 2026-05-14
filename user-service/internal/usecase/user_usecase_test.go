package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"elibrary/user-service/internal/domain"
)

// --- Hand-rolled mock ---

type mockRepo struct {
	users      map[string]*domain.User
	byTok      map[string]*domain.User
	byMail     map[string]*domain.User
	failCreate bool
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		users:  map[string]*domain.User{},
		byTok:  map[string]*domain.User{},
		byMail: map[string]*domain.User{},
	}
}

func (m *mockRepo) Create(_ context.Context, u *domain.User) error {
	if m.failCreate {
		return domain.ErrUserAlreadyExists
	}
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	m.users[u.ID] = u
	m.byMail[u.Email] = u
	if u.VerificationToken != "" {
		m.byTok[u.VerificationToken] = u
	}
	return nil
}
func (m *mockRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}
func (m *mockRepo) GetByEmail(_ context.Context, e string) (*domain.User, error) {
	if u, ok := m.byMail[e]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}
func (m *mockRepo) GetByVerificationToken(_ context.Context, t string) (*domain.User, error) {
	if u, ok := m.byTok[t]; ok {
		return u, nil
	}
	return nil, domain.ErrInvalidToken
}
func (m *mockRepo) Update(_ context.Context, u *domain.User) error { m.users[u.ID] = u; return nil }
func (m *mockRepo) UpdatePassword(_ context.Context, id, h string) error {
	m.users[id].PasswordHash = h
	return nil
}
func (m *mockRepo) MarkEmailVerified(_ context.Context, id string) error {
	m.users[id].EmailVerified = true
	return nil
}
func (m *mockRepo) Delete(_ context.Context, id string) error                     { delete(m.users, id); return nil }
func (m *mockRepo) List(_ context.Context, _, _ int) ([]*domain.User, int, error) { return nil, 0, nil }
func (m *mockRepo) Search(_ context.Context, _ string, _, _ int) ([]*domain.User, int, error) {
	return nil, 0, nil
}

// --- Tests ---

func TestCreate_OK(t *testing.T) {
	uc := NewUserUsecase(newMockRepo(), "secret", nil, nil, nil, "")
	u, err := uc.Create(context.Background(), "Nurbol", "n@e.com", "secret123")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if u.ID == "" || u.PasswordHash == "" || u.VerificationToken == "" {
		t.Fatalf("missing generated fields: %+v", u)
	}
}

func TestCreate_ShortPassword(t *testing.T) {
	uc := NewUserUsecase(newMockRepo(), "secret", nil, nil, nil, "")
	_, err := uc.Create(context.Background(), "N", "n@e.com", "123")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("want ErrInvalidArgument, got %v", err)
	}
}

func TestChangePassword_OK(t *testing.T) {
	repo := newMockRepo()
	uc := NewUserUsecase(repo, "secret", nil, nil, nil, "")
	_, _ = uc.Create(context.Background(), "N", "n@e.com", "oldpass1")
	var id string
	for k := range repo.users {
		id = k
	}
	if err := uc.ChangePassword(context.Background(), id, "oldpass1", "newpass1"); err != nil {
		t.Fatalf("change: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.users[id].PasswordHash), []byte("newpass1")); err != nil {
		t.Fatalf("password not updated: %v", err)
	}
}

func TestChangePassword_WrongOld(t *testing.T) {
	repo := newMockRepo()
	uc := NewUserUsecase(repo, "secret", nil, nil, nil, "")
	_, _ = uc.Create(context.Background(), "N", "n@e.com", "oldpass1")
	var id string
	for k := range repo.users {
		id = k
	}
	err := uc.ChangePassword(context.Background(), id, "WRONG", "newpass1")
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Fatalf("want ErrInvalidPassword, got %v", err)
	}
}

func TestVerifyEmail_OK(t *testing.T) {
	repo := newMockRepo()
	uc := NewUserUsecase(repo, "secret", nil, nil, nil, "")
	_, _ = uc.Create(context.Background(), "N", "n@e.com", "secret123")
	var tok string
	for k := range repo.byTok {
		tok = k
	}
	u, err := uc.VerifyEmail(context.Background(), tok)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !u.EmailVerified {
		t.Fatalf("not verified")
	}
}

func TestVerifyEmail_BadToken(t *testing.T) {
	uc := NewUserUsecase(newMockRepo(), "secret", nil, nil, nil, "")
	_, err := uc.VerifyEmail(context.Background(), "no-such-token")
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("want ErrInvalidToken, got %v", err)
	}
}

func TestLogin_OK(t *testing.T) {
	repo := newMockRepo()
	uc := NewUserUsecase(repo, "secret", nil, nil, nil, "")
	_, _ = uc.Create(context.Background(), "N", "n@e.com", "secret123")
	tok, u, err := uc.Login(context.Background(), "n@e.com", "secret123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if tok == "" || u == nil {
		t.Fatalf("empty token or user")
	}
}
