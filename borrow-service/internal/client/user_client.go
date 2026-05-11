package client

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"elibrary/borrow-service/internal/domain"
	"elibrary/gen/userpb"
)

type UserClient struct {
	c userpb.UserServiceClient
}

func NewUserClient(c userpb.UserServiceClient) *UserClient {
	return &UserClient{c: c}
}

func (u *UserClient) Exists(ctx context.Context, userID string) error {
	_, err := u.c.GetUserById(ctx, &userpb.GetUserByIdRequest{UserId: userID})
	if err == nil {
		return nil
	}
	if status.Code(err) == codes.NotFound {
		return domain.ErrUserNotFound
	}
	return err
}