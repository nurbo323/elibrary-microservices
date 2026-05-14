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

func (h *UserHandler) SearchUsers(ctx context.Context, req *userpb.SearchUsersRequest) (*userpb.ListUsersResponse, error) {
	users, total, err := h.uc.Search(ctx, req.GetQuery(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*userpb.User, len(users))
	for i, u := range users {
		out[i] = toProto(u)
	}
	return &userpb.ListUsersResponse{Users: out, Total: int32(total)}, nil
}

func (h *UserHandler) ChangePassword(ctx context.Context, req *userpb.ChangePasswordRequest) (*emptypb.Empty, error) {
	if err := h.uc.ChangePassword(ctx, req.GetUserId(), req.GetOldPassword(), req.GetNewPassword()); err != nil {
		return nil, mapErr(err)
	}
	return &emptypb.Empty{}, nil
}

// Исправленная версия VerifyEmail
func (h *UserHandler) VerifyEmail(ctx context.Context, req *userpb.VerifyEmailRequest) (*emptypb.Empty, error) {
	// Вызываем usecase, результат (_) нам не важен для ответа gRPC
	_, err := h.uc.VerifyEmail(ctx, req.GetToken())
	if err != nil {
		return nil, mapErr(err)
	}

	// Возвращаем ПРАВИЛЬНЫЙ тип: emptypb.Empty
	return &emptypb.Empty{}, nil
}
func (h *UserHandler) GetUserProfile(ctx context.Context, req *userpb.GetUserByIdRequest) (*userpb.UserProfileResponse, error) {
	u, total, active, err := h.uc.GetProfile(ctx, req.GetUserId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &userpb.UserProfileResponse{
		User: toProto(u), TotalBorrows: int32(total), ActiveBorrows: int32(active),
	}, nil
}

func (h *UserHandler) GetUserActiveBorrows(ctx context.Context, req *userpb.GetUserByIdRequest) (*userpb.UserActiveBorrowsResponse, error) {
	list, err := h.uc.GetActiveBorrows(ctx, req.GetUserId())
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*userpb.ActiveBorrow, 0, len(list))
	for _, b := range list {
		out = append(out, &userpb.ActiveBorrow{
			BorrowId: b.BorrowID, BookId: b.BookID,
			DateFrom: b.DateFrom, DateTo: b.DateTo, Status: b.Status,
		})
	}
	return &userpb.UserActiveBorrowsResponse{Borrows: out}, nil
}

func (h *UserHandler) GetUserStatistics(ctx context.Context, req *userpb.GetUserByIdRequest) (*userpb.UserStatisticsResponse, error) {
	stats, err := h.uc.GetStatistics(ctx, req.GetUserId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &userpb.UserStatisticsResponse{
		TotalBorrows:    int32(stats.Total),
		ActiveBorrows:   int32(stats.Active),
		ReturnedBorrows: int32(stats.Returned),
		OverdueBorrows:  int32(stats.Overdue),
	}, nil
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
	case errors.Is(err, domain.ErrInvalidToken):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrSamePassword):
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
