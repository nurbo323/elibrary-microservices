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

func (b *BookClient) GetBook(ctx context.Context, bookID string) (*bookpb.Book, error) {
	resp, err := b.c.GetBookById(ctx, &bookpb.GetBookByIdRequest{BookId: bookID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrBookNotFound
		}
		return nil, err
	}
	return resp.GetBook(), nil
}

func (b *BookClient) UpdateCopyStatus(ctx context.Context, expID, statusValue string) (*bookpb.BookCopy, error) {
	resp, err := b.c.UpdateCopyStatus(ctx, &bookpb.UpdateCopyStatusRequest{
		ExpId:  expID,
		Status: statusValue,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetCopy(), nil
}