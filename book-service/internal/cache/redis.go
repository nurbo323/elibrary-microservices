package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"elibrary/book-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

type BookCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func New(addr string, ttl time.Duration) (*BookCache, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr})

	// Проверка соединения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &BookCache{rdb: rdb, ttl: ttl}, nil
}

func (c *BookCache) Close() error { return c.rdb.Close() }

// --- Работа с одной книгой ---
func bookKey(id string) string { return "book:" + id }

func (c *BookCache) GetBook(ctx context.Context, id string) (*domain.Book, bool, error) {
	raw, err := c.rdb.Get(ctx, bookKey(id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	b := &domain.Book{}
	if err := json.Unmarshal(raw, b); err != nil {
		return nil, false, err
	}
	return b, true, nil
}

func (c *BookCache) SetBook(ctx context.Context, b *domain.Book) error {
	data, err := json.Marshal(b)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, bookKey(b.ID), data, c.ttl).Err()
}

func (c *BookCache) InvalidateBook(ctx context.Context, id string) error {
	return c.rdb.Del(ctx, bookKey(id)).Err()
}

// --- Работа со списками и поиском ---
func listKey(prefix string, limit, offset int) string {
	return fmt.Sprintf("books:%s:%d:%d", prefix, limit, offset)
}

type cachedList struct {
	Books []*domain.Book `json:"books"`
	Total int            `json:"total"`
}

func (c *BookCache) GetList(ctx context.Context, prefix string, limit, offset int) ([]*domain.Book, int, bool, error) {
	raw, err := c.rdb.Get(ctx, listKey(prefix, limit, offset)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, 0, false, nil
	}
	if err != nil {
		return nil, 0, false, err
	}
	cl := &cachedList{}
	if err := json.Unmarshal(raw, cl); err != nil {
		return nil, 0, false, err
	}
	return cl.Books, cl.Total, true, nil
}

func (c *BookCache) SetList(ctx context.Context, prefix string, limit, offset, total int, books []*domain.Book) error {
	data, err := json.Marshal(cachedList{Books: books, Total: total})
	if err != nil {
		return err
	}
	// Для списков ставим короткий TTL — 30 секунд (как в туториале)
	return c.rdb.Set(ctx, listKey(prefix, limit, offset), data, 30*time.Second).Err()
}

// Очистка всех списков при изменении данных (инвалидация)
func (c *BookCache) InvalidateAllLists(ctx context.Context) error {
	patterns := []string{
		"books:list:*",
		"books:search:*",
		"books:author:*",
		"books:year:*",
		"book:*:copies",
		"book:*:available",
	}
	for _, p := range patterns {
		iter := c.rdb.Scan(ctx, 0, p, 100).Iterator()
		for iter.Next(ctx) {
			if err := c.rdb.Del(ctx, iter.Val()).Err(); err != nil {
				return err
			}
		}
		if err := iter.Err(); err != nil {
			return err
		}
	}
	return nil
}
