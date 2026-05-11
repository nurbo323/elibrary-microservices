package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"elibrary/borrow-service/internal/domain"
)

type BorrowRepo struct {
	pool *pgxpool.Pool
}

func NewBorrowRepo(pool *pgxpool.Pool) *BorrowRepo {
	return &BorrowRepo{pool: pool}
}

func (r *BorrowRepo) Create(ctx context.Context, b *domain.Borrow) error {
	const q = `
		INSERT INTO borrows (borrow_id, user_id, book_id, barcode, date_from, date_to, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`
	err := r.pool.QueryRow(ctx, q,
		b.ID, b.UserID, b.BookID, b.Barcode, b.DateFrom, b.DateTo, b.Status,
	).Scan(&b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create borrow: %w", err)
	}
	return nil
}

func (r *BorrowRepo) GetByID(ctx context.Context, id string) (*domain.Borrow, error) {
	const q = `
		SELECT borrow_id, user_id, book_id, COALESCE(exp_id::text, ''), barcode,
		       date_from, date_to, status, created_at, updated_at
		FROM borrows WHERE borrow_id = $1`
	b := &domain.Borrow{}
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&b.ID, &b.UserID, &b.BookID, &b.ExpID, &b.Barcode,
		&b.DateFrom, &b.DateTo, &b.Status, &b.CreatedAt, &b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrBorrowNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get borrow: %w", err)
	}
	return b, nil
}

func (r *BorrowRepo) UpdateStatus(ctx context.Context, id, newStatus string) error {
	const q = `UPDATE borrows SET status = $2, updated_at = NOW() WHERE borrow_id = $1`
	tag, err := r.pool.Exec(ctx, q, id, newStatus)
	if err != nil {
		return fmt.Errorf("update borrow status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBorrowNotFound
	}
	return nil
}

func (r *BorrowRepo) List(ctx context.Context, limit, offset int) ([]*domain.Borrow, int, error) {
	return r.queryList(ctx,
		`SELECT COUNT(*) FROM borrows`,
		`SELECT borrow_id, user_id, book_id, COALESCE(exp_id::text, ''), barcode,
		        date_from, date_to, status, created_at, updated_at
		 FROM borrows ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		[]any{limit, offset},
	)
}

func (r *BorrowRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Borrow, int, error) {
	return r.queryList(ctx,
		`SELECT COUNT(*) FROM borrows WHERE user_id = $1`,
		`SELECT borrow_id, user_id, book_id, COALESCE(exp_id::text, ''), barcode,
		        date_from, date_to, status, created_at, updated_at
		 FROM borrows WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		[]any{userID, limit, offset},
		userID,
	)
}

func (r *BorrowRepo) ListActive(ctx context.Context, limit, offset int) ([]*domain.Borrow, int, error) {
	return r.queryList(ctx,
		`SELECT COUNT(*) FROM borrows WHERE status = 'ACTIVE'`,
		`SELECT borrow_id, user_id, book_id, COALESCE(exp_id::text, ''), barcode,
		        date_from, date_to, status, created_at, updated_at
		 FROM borrows WHERE status = 'ACTIVE' ORDER BY date_to ASC LIMIT $1 OFFSET $2`,
		[]any{limit, offset},
	)
}

// helper: countArgs — это аргументы, которые используются в COUNT-запросе (если есть).
func (r *BorrowRepo) queryList(
	ctx context.Context, countQ, listQ string, listArgs []any, countArgs ...any,
) ([]*domain.Borrow, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, countQ, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count borrows: %w", err)
	}
	rows, err := r.pool.Query(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list borrows: %w", err)
	}
	defer rows.Close()

	out := make([]*domain.Borrow, 0)
	for rows.Next() {
		b := &domain.Borrow{}
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.BookID, &b.ExpID, &b.Barcode,
			&b.DateFrom, &b.DateTo, &b.Status, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan borrow: %w", err)
		}
		out = append(out, b)
	}
	return out, total, rows.Err()
}