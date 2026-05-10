package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"elibrary/book-service/internal/domain"
)

type CopyRepo struct {
	pool *pgxpool.Pool
}

func NewCopyRepo(pool *pgxpool.Pool) *CopyRepo {
	return &CopyRepo{pool: pool}
}

func (r *CopyRepo) Create(ctx context.Context, c *domain.BookCopy) error {
	const q = `
		INSERT INTO book_copies (exp_id, book_id, status)
		VALUES ($1, $2, $3)
		RETURNING created_at`
	err := r.pool.QueryRow(ctx, q, c.ExpID, c.BookID, c.Status).Scan(&c.CreatedAt)
	if err != nil {
		if isForeignKeyViolation(err) {
			return domain.ErrBookNotFound
		}
		return fmt.Errorf("create copy: %w", err)
	}
	return nil
}

func (r *CopyRepo) GetByID(ctx context.Context, id string) (*domain.BookCopy, error) {
	const q = `SELECT exp_id, book_id, status, created_at FROM book_copies WHERE exp_id = $1`
	c := &domain.BookCopy{}
	err := r.pool.QueryRow(ctx, q, id).Scan(&c.ExpID, &c.BookID, &c.Status, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCopyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get copy: %w", err)
	}
	return c, nil
}

func (r *CopyRepo) ListByBook(ctx context.Context, bookID string) ([]*domain.BookCopy, error) {
	const q = `
		SELECT exp_id, book_id, status, created_at
		FROM book_copies WHERE book_id = $1 ORDER BY created_at`
	rows, err := r.pool.Query(ctx, q, bookID)
	if err != nil {
		return nil, fmt.Errorf("list copies: %w", err)
	}
	defer rows.Close()

	copies := make([]*domain.BookCopy, 0)
	for rows.Next() {
		c := &domain.BookCopy{}
		if err := rows.Scan(&c.ExpID, &c.BookID, &c.Status, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan copy: %w", err)
		}
		copies = append(copies, c)
	}
	return copies, rows.Err()
}

func isForeignKeyViolation(err error) bool {
	type pgErr interface{ SQLState() string }
	var pe pgErr
	if errors.As(err, &pe) {
		return pe.SQLState() == "23503"
	}
	return false
}
