package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"elibrary/user-service/internal/domain"
)

type UserUsecase struct {
	repo      domain.UserRepository
	jwtSecret []byte
}

func NewUserUsecase(repo domain.UserRepository, jwtSecret string) *UserUsecase {
	return &UserUsecase{repo: repo, jwtSecret: []byte(jwtSecret)}
}

func (uc *UserUsecase) Create(ctx context.Context, name, email, password string) (*domain.User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if name == "" || email == "" || len(password) < 6 {
		return nil, fmt.Errorf("%w: name, email and password (min 6 chars) required", domain.ErrInvalidArgument)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	u := &domain.User{
		ID:           uuid.NewString(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := uc.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (uc *UserUsecase) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if id == "" {
		return nil, domain.ErrInvalidArgument
	}
	return uc.repo.GetByID(ctx, id)
}

func (uc *UserUsecase) Update(ctx context.Context, id, name, email string) (*domain.User, error) {
	if id == "" {
		return nil, domain.ErrInvalidArgument
	}
	u, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name = strings.TrimSpace(name); name != "" {
		u.Name = name
	}
	if email = strings.ToLower(strings.TrimSpace(email)); email != "" {
		u.Email = email
	}
	if err := uc.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (uc *UserUsecase) Delete(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrInvalidArgument
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *UserUsecase) List(ctx context.Context, limit, offset int) ([]*domain.User, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.repo.List(ctx, limit, offset)
}

func (uc *UserUsecase) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return "", nil, domain.ErrInvalidArgument
	}
	u, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", nil, domain.ErrInvalidPassword
		}
		return "", nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", nil, domain.ErrInvalidPassword
	}
	token, err := uc.issueToken(u.ID)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

func (uc *UserUsecase) issueToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(uc.jwtSecret)
}
