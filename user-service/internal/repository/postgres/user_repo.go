package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"elibrary/user-service/internal/domain"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	const q = `
		INSERT INTO users (user_id, name, email, password_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at`
	err := r.pool.QueryRow(ctx, q, u.ID, u.Name, u.Email, u.PasswordHash).
		Scan(&u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `
		SELECT user_id, name, email, password_hash, email_verified, created_at, updated_at
		FROM users WHERE user_id = $1`
	u := &domain.User{}
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.EmailVerified, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT user_id, name, email, password_hash, email_verified, created_at, updated_at
		FROM users WHERE email = $1`
	u := &domain.User{}
	err := r.pool.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.EmailVerified, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	const q = `
		UPDATE users
		SET name = $2, email = $3, updated_at = NOW()
		WHERE user_id = $1`
	tag, err := r.pool.Exec(ctx, q, u.ID, u.Name, u.Email)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("update user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM users WHERE user_id = $1`
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) List(ctx context.Context, limit, offset int) ([]*domain.User, int, error) {
	const countQ = `SELECT COUNT(*) FROM users`
	var total int
	if err := r.pool.QueryRow(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	const q = `
		SELECT user_id, name, email, password_hash, email_verified, created_at, updated_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0, limit)
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.EmailVerified, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func isUniqueViolation(err error) bool {
	type pgErr interface{ SQLState() string }
	var pe pgErr
	if errors.As(err, &pe) {
		return pe.SQLState() == "23505"
	}
	return false
}
