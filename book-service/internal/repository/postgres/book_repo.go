package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"elibrary/book-service/internal/domain"
)

type BookRepo struct {
	pool *pgxpool.Pool
}

func NewBookRepo(pool *pgxpool.Pool) *BookRepo {
	return &BookRepo{pool: pool}
}

func (r *BookRepo) Create(ctx context.Context, b *domain.Book) error {
	const q = `
		INSERT INTO books (book_id, name, authors, year, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`
	err := r.pool.QueryRow(ctx, q, b.ID, b.Name, b.Authors, b.Year, b.Status).
		Scan(&b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create book: %w", err)
	}
	return nil
}

func (r *BookRepo) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	const q = `
		SELECT book_id, name, authors, year, status, created_at, updated_at
		FROM books WHERE book_id = $1`
	b := &domain.Book{}
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&b.ID, &b.Name, &b.Authors, &b.Year, &b.Status, &b.CreatedAt, &b.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrBookNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get book by id: %w", err)
	}
	return b, nil
}

func (r *BookRepo) Update(ctx context.Context, b *domain.Book) error {
	const q = `
		UPDATE books
		SET name = $2, authors = $3, year = $4, updated_at = NOW()
		WHERE book_id = $1`
	tag, err := r.pool.Exec(ctx, q, b.ID, b.Name, b.Authors, b.Year)
	if err != nil {
		return fmt.Errorf("update book: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBookNotFound
	}
	return nil
}

func (r *BookRepo) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM books WHERE book_id = $1`
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete book: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBookNotFound
	}
	return nil
}

func (r *BookRepo) List(ctx context.Context, limit, offset int) ([]*domain.Book, int, error) {
	const countQ = `SELECT COUNT(*) FROM books`
	var total int
	if err := r.pool.QueryRow(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count books: %w", err)
	}

	const q = `
		SELECT book_id, name, authors, year, status, created_at, updated_at
		FROM books ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()

	books := make([]*domain.Book, 0, limit)
	for rows.Next() {
		b := &domain.Book{}
		if err := rows.Scan(&b.ID, &b.Name, &b.Authors, &b.Year, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan book: %w", err)
		}
		books = append(books, b)
	}
	return books, total, rows.Err()
}
