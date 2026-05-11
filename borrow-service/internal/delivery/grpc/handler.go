package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"elibrary/borrow-service/internal/domain"
	"elibrary/borrow-service/internal/usecase"
	"elibrary/gen/borrowpb"
)

type BorrowHandler struct {
	borrowpb.UnimplementedBorrowServiceServer
	uc *usecase.BorrowUsecase
}

func NewBorrowHandler(uc *usecase.BorrowUsecase) *BorrowHandler {
	return &BorrowHandler{uc: uc}
}

func toProto(b *domain.Borrow) *borrowpb.Borrow {
	return &borrowpb.Borrow{
		BorrowId: b.ID,
		ExpId:    b.ExpID,
		UserId:   b.UserID,
		Barcode:  b.Barcode,
		DateFrom: timestamppb.New(b.DateFrom),
		DateTo:   timestamppb.New(b.DateTo),
		Status:   b.Status,
	}
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrBorrowNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrBookNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrNotActive):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (h *BorrowHandler) BorrowBook(ctx context.Context, req *borrowpb.BorrowBookRequest) (*borrowpb.BorrowResponse, error) {
	b, err := h.uc.Borrow(ctx, req.GetUserId(), req.GetBookId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &borrowpb.BorrowResponse{Borrow: toProto(b)}, nil
}

func (h *BorrowHandler) ReturnBook(ctx context.Context, req *borrowpb.ReturnBookRequest) (*borrowpb.BorrowResponse, error) {
	b, err := h.uc.Return(ctx, req.GetBorrowId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &borrowpb.BorrowResponse{Borrow: toProto(b)}, nil
}

func (h *BorrowHandler) GetBorrowById(ctx context.Context, req *borrowpb.GetBorrowByIdRequest) (*borrowpb.BorrowResponse, error) {
	b, err := h.uc.GetByID(ctx, req.GetBorrowId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &borrowpb.BorrowResponse{Borrow: toProto(b)}, nil
}

func (h *BorrowHandler) ListBorrows(ctx context.Context, req *borrowpb.ListBorrowsRequest) (*borrowpb.ListBorrowsResponse, error) {
	items, total, err := h.uc.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	return &borrowpb.ListBorrowsResponse{Borrows: mapBorrows(items), Total: int32(total)}, nil
}

func (h *BorrowHandler) GetUserBorrowHistory(ctx context.Context, req *borrowpb.GetUserBorrowHistoryRequest) (*borrowpb.ListBorrowsResponse, error) {
	items, total, err := h.uc.UserHistory(ctx, req.GetUserId(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	return &borrowpb.ListBorrowsResponse{Borrows: mapBorrows(items), Total: int32(total)}, nil
}

func (h *BorrowHandler) GetActiveBorrows(ctx context.Context, req *borrowpb.GetActiveBorrowsRequest) (*borrowpb.ListBorrowsResponse, error) {
	items, total, err := h.uc.Active(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	return &borrowpb.ListBorrowsResponse{Borrows: mapBorrows(items), Total: int32(total)}, nil
}

func mapBorrows(in []*domain.Borrow) []*borrowpb.Borrow {
	out := make([]*borrowpb.Borrow, len(in))
	for i, b := range in {
		out[i] = toProto(b)
	}
	return out
}