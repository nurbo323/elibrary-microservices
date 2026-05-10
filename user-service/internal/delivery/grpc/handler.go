package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"elibrary/gen/userpb"
	"elibrary/user-service/internal/domain"
	"elibrary/user-service/internal/usecase"
)

type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	uc *usecase.UserUsecase
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

func toProto(u *domain.User) *userpb.User {
	return &userpb.User{
		UserId:        u.ID,
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		CreatedAt:     timestamppb.New(u.CreatedAt),
	}
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInvalidPassword):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.UserResponse, error) {
	u, err := h.uc.Create(ctx, req.GetName(), req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapErr(err)
	}
	return &userpb.UserResponse{User: toProto(u)}, nil
}

func (h *UserHandler) GetUserById(ctx context.Context, req *userpb.GetUserByIdRequest) (*userpb.UserResponse, error) {
	u, err := h.uc.GetByID(ctx, req.GetUserId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &userpb.UserResponse{User: toProto(u)}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UserResponse, error) {
	u, err := h.uc.Update(ctx, req.GetUserId(), req.GetName(), req.GetEmail())
	if err != nil {
		return nil, mapErr(err)
	}
	return &userpb.UserResponse{User: toProto(u)}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*emptypb.Empty, error) {
	if err := h.uc.Delete(ctx, req.GetUserId()); err != nil {
		return nil, mapErr(err)
	}
	return &emptypb.Empty{}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	users, total, err := h.uc.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*userpb.User, len(users))
	for i, u := range users {
		out[i] = toProto(u)
	}
	return &userpb.ListUsersResponse{Users: out, Total: int32(total)}, nil
}

func (h *UserHandler) LoginUser(ctx context.Context, req *userpb.LoginUserRequest) (*userpb.LoginUserResponse, error) {
	token, u, err := h.uc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapErr(err)
	}
	return &userpb.LoginUserResponse{Token: token, User: toProto(u)}, nil
}
