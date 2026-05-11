package client

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"elibrary/borrow-service/internal/domain"
	"elibrary/gen/bookpb"
)

type BookClient struct {
	c bookpb.BookServiceClient
}

func NewBookClient(c bookpb.BookServiceClient) *BookClient {
	return &BookClient{c: c}
}

func (b *BookClient) Exists(ctx context.Context, bookID string) error {
	_, err := b.c.GetBookById(ctx, &bookpb.GetBookByIdRequest{BookId: bookID})
	if err == nil {
		return nil
	}
	if status.Code(err) == codes.NotFound {
		return domain.ErrBookNotFound
	}
	return err
}