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
		INSERT INTO borrows (borrow_id, user_id, book_id, exp_id, barcode, date_from, date_to, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`

	err := r.pool.QueryRow(ctx, q,
		b.ID, b.UserID, b.BookID, b.ExpID, b.Barcode, b.DateFrom, b.DateTo, b.Status,
	).Scan(&b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create borrow: %w", err)
	}

	return nil
}

func (r *BorrowRepo) GetByID(ctx context.Context, id string) (*domain.Borrow, error) {
	const q = `
		SELECT borrow_id, user_id, book_id, COALESCE(exp_id::text, ''), barcode,
		       date_from, date_to, returned_at, status, created_at, updated_at
		FROM borrows
		WHERE borrow_id = $1`

	b := &domain.Borrow{}

	err := r.pool.QueryRow(ctx, q, id).Scan(
		&b.ID,
		&b.UserID,
		&b.BookID,
		&b.ExpID,
		&b.Barcode,
		&b.DateFrom,
		&b.DateTo,
		&b.ReturnedAt,
		&b.Status,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrBorrowNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get borrow: %w", err)
	}

	return b, nil
}

func (r *BorrowRepo) GetActiveByExpID(ctx context.Context, expID string) (*domain.Borrow, error) {
	const q = `
		SELECT borrow_id, user_id, book_id, COALESCE(exp_id::text, ''), barcode,
		       date_from, date_to, returned_at, status, created_at, updated_at
		FROM borrows
		WHERE exp_id = $1 AND status IN ('ACTIVE', 'RESERVED')
		ORDER BY date_from DESC
		LIMIT 1`

	b := &domain.Borrow{}

	err := r.pool.QueryRow(ctx, q, expID).Scan(
		&b.ID,
		&b.UserID,
		&b.BookID,
		&b.ExpID,
		&b.Barcode,
		&b.DateFrom,
		&b.DateTo,
		&b.ReturnedAt,
		&b.Status,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrBorrowNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get active borrow by exp_id: %w", err)
	}

	return b, nil
}

func (r *BorrowRepo) Update(ctx context.Context, b *domain.Borrow) error {
	const q = `
		UPDATE borrows
		SET user_id = $2,
		    book_id = $3,
		    exp_id = $4,
		    barcode = $5,
		    date_from = $6,
		    date_to = $7,
		    returned_at = $8,
		    status = $9,
		    updated_at = NOW()
		WHERE borrow_id = $1
		RETURNING updated_at`

	err := r.pool.QueryRow(ctx, q,
		b.ID,
		b.UserID,
		b.BookID,
		b.ExpID,
		b.Barcode,
		b.DateFrom,
		b.DateTo,
		b.ReturnedAt,
		b.Status,
	).Scan(&b.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrBorrowNotFound
	}
	if err != nil {
		return fmt.Errorf("update borrow: %w", err)
	}

	return nil
}

func (r *BorrowRepo) UpdateStatus(ctx context.Context, id, newStatus string) error {
	const q = `
		UPDATE borrows
		SET status = $2, updated_at = NOW()
		WHERE borrow_id = $1`

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
		        date_from, date_to, returned_at, status, created_at, updated_at
		 FROM borrows
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		[]any{limit, offset},
	)
}

func (r *BorrowRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Borrow, int, error) {
	return r.queryList(ctx,
		`SELECT COUNT(*) FROM borrows WHERE user_id = $1`,
		`SELECT borrow_id, user_id, book_id, COALESCE(exp_id::text, ''), barcode,
		        date_from, date_to, returned_at, status, created_at, updated_at
		 FROM borrows
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		[]any{userID, limit, offset},
		userID,
	)
}

func (r *BorrowRepo) ListActive(ctx context.Context, limit, offset int) ([]*domain.Borrow, int, error) {
	return r.queryList(ctx,
		`SELECT COUNT(*) FROM borrows WHERE status = 'ACTIVE'`,
		`SELECT borrow_id, user_id, book_id, COALESCE(exp_id::text, ''), barcode,
		        date_from, date_to, returned_at, status, created_at, updated_at
		 FROM borrows
		 WHERE status = 'ACTIVE'
		 ORDER BY date_to ASC
		 LIMIT $1 OFFSET $2`,
		[]any{limit, offset},
	)
}

func (r *BorrowRepo) ListOverdue(ctx context.Context, limit, offset int) ([]*domain.Borrow, int, error) {
	return r.queryList(ctx,
		`SELECT COUNT(*) FROM borrows WHERE status = 'ACTIVE' AND date_to < NOW()`,
		`SELECT borrow_id, user_id, book_id, COALESCE(exp_id::text, ''), barcode,
		        date_from, date_to, returned_at, status, created_at, updated_at
		 FROM borrows
		 WHERE status = 'ACTIVE' AND date_to < NOW()
		 ORDER BY date_to ASC
		 LIMIT $1 OFFSET $2`,
		[]any{limit, offset},
	)
}

func (r *BorrowRepo) queryList(
	ctx context.Context,
	countQ string,
	listQ string,
	listArgs []any,
	countArgs ...any,
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
			&b.ID,
			&b.UserID,
			&b.BookID,
			&b.ExpID,
			&b.Barcode,
			&b.DateFrom,
			&b.DateTo,
			&b.ReturnedAt,
			&b.Status,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan borrow: %w", err)
		}

		out = append(out, b)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return out, total, nil
}