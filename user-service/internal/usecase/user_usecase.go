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

	"crypto/rand"
	"encoding/hex"
	"log"

	"elibrary/pkg/eventbus"
	"elibrary/user-service/internal/clients"
	"elibrary/user-service/internal/mailer"
)

type UserUsecase struct {
	repo         domain.UserRepository
	jwtSecret    []byte
	publisher    *eventbus.Publisher
	mailer       *mailer.Mailer
	borrowClient *clients.BorrowClient
	verifyURL    string // e.g. http://localhost:8080/api/auth/verify?token=
}

func NewUserUsecase(
	repo domain.UserRepository,
	jwtSecret string,
	publisher *eventbus.Publisher,
	m *mailer.Mailer,
	bc *clients.BorrowClient,
	verifyURL string,
) *UserUsecase {
	return &UserUsecase{
		repo:         repo,
		jwtSecret:    []byte(jwtSecret),
		publisher:    publisher,
		mailer:       m,
		borrowClient: bc,
		verifyURL:    verifyURL,
	}
}

func (uc *UserUsecase) Create(ctx context.Context, name, email, password string) (*domain.User, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	if name == "" || email == "" || len(password) < 6 {
		return nil, fmt.Errorf("%w: name, email and password (min 6 chars) are required", domain.ErrInvalidArgument)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	token, err := randomToken(32)
	if err != nil {
		return nil, fmt.Errorf("gen token: %w", err)
	}
	u := &domain.User{
		ID:                uuid.NewString(),
		Name:              name,
		Email:             email,
		PasswordHash:      string(hash),
		VerificationToken: token,
	}
	if err := uc.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	// fire-and-forget event + email (logging errors, not failing user creation)
	go func() {
		if uc.publisher != nil {
			ev := eventbus.UserCreatedEvent{
				UserID: u.ID, Name: u.Name, Email: u.Email,
				VerificationToken: token, CreatedAt: u.CreatedAt,
			}
			if err := uc.publisher.Publish(context.Background(), "user.created", ev); err != nil {
				log.Printf("publish user.created: %v", err)
			}
		}
		if uc.mailer != nil {
			link := fmt.Sprintf("%s%s", uc.verifyURL, token)
			if err := uc.mailer.SendWelcome(u.Email, u.Name, link); err != nil {
				log.Printf("send welcome email: %v", err)
			}
		}
	}()

	return u, nil
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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

func (uc *UserUsecase) Search(ctx context.Context, query string, limit, offset int) ([]*domain.User, int, error) {
	if strings.TrimSpace(query) == "" {
		return nil, 0, fmt.Errorf("%w: query is required", domain.ErrInvalidArgument)
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return uc.repo.Search(ctx, query, limit, offset)
}

func (uc *UserUsecase) ChangePassword(ctx context.Context, id, oldPass, newPass string) error {
	if id == "" || oldPass == "" || len(newPass) < 6 {
		return fmt.Errorf("%w: id, old and new (min 6 chars) password required", domain.ErrInvalidArgument)
	}
	if oldPass == newPass {
		return domain.ErrSamePassword
	}
	u, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPass)); err != nil {
		return domain.ErrInvalidPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash: %w", err)
	}
	return uc.repo.UpdatePassword(ctx, id, string(hash))
}

func (uc *UserUsecase) VerifyEmail(ctx context.Context, token string) (*domain.User, error) {
	if token == "" {
		return nil, domain.ErrInvalidArgument
	}
	u, err := uc.repo.GetByVerificationToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if u.EmailVerified {
		return u, nil
	}
	if err := uc.repo.MarkEmailVerified(ctx, u.ID); err != nil {
		return nil, err
	}
	u.EmailVerified = true
	u.VerificationToken = ""
	return u, nil
}

func (uc *UserUsecase) GetProfile(ctx context.Context, id string) (*domain.User, int, int, error) {
	u, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, 0, 0, err
	}
	if uc.borrowClient == nil {
		return u, 0, 0, nil
	}
	active, _ := uc.borrowClient.GetActiveByUser(ctx, id)
	history, _ := uc.borrowClient.GetHistoryByUser(ctx, id)
	activeCount, total := 0, 0
	if active != nil {
		activeCount = len(active.GetBorrows())
	}
	if history != nil {
		total = len(history.GetBorrows())
	}
	return u, total, activeCount, nil
}

type ActiveBorrowDTO struct {
	BorrowID string
	BookID   string
	DateFrom string
	DateTo   string
	Status   string
}

func (uc *UserUsecase) GetActiveBorrows(ctx context.Context, userID string) ([]ActiveBorrowDTO, error) {
	if uc.borrowClient == nil {
		return nil, nil
	}
	resp, err := uc.borrowClient.GetActiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ActiveBorrowDTO, 0, len(resp.GetBorrows()))
	for _, b := range resp.GetBorrows() {
		out = append(out, ActiveBorrowDTO{
			BorrowID: b.GetBorrowId(),
			BookID:   b.GetExpId(),
			DateFrom: b.GetDateFrom().AsTime().Format("2006-01-02"),
			DateTo:   b.GetDateTo().AsTime().Format("2006-01-02"),
			Status:   b.GetStatus(),
		})
	}
	return out, nil
}

type UserStatsDTO struct {
	Total    int
	Active   int
	Returned int
	Overdue  int
}

func (uc *UserUsecase) GetStatistics(ctx context.Context, userID string) (UserStatsDTO, error) {
	if uc.borrowClient == nil {
		return UserStatsDTO{}, nil
	}
	history, err := uc.borrowClient.GetHistoryByUser(ctx, userID)
	if err != nil {
		return UserStatsDTO{}, err
	}
	stats := UserStatsDTO{}
	for _, b := range history.GetBorrows() {
		stats.Total++
		switch strings.ToUpper(b.GetStatus()) {
		case "ACTIVE", "BORROWED":
			stats.Active++
		case "RETURNED":
			stats.Returned++
		case "OVERDUE":
			stats.Overdue++
		}
	}
	return stats, nil
}
