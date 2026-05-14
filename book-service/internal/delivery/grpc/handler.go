package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"elibrary/book-service/internal/domain"
	"elibrary/book-service/internal/usecase"

	// ВАЖНО: Убедись, что этот путь совпадает с тем, где лежат твои файлы
	"elibrary/elibrary/gen/bookpb"
)

type BookHandler struct {
	bookpb.UnimplementedBookServiceServer
	uc *usecase.BookUsecase
}

func NewBookHandler(uc *usecase.BookUsecase) *BookHandler {
	return &BookHandler{uc: uc}
}

// --- Вспомогательные функции (Mappers) ---

func toProto(b *domain.Book) *bookpb.Book {
	return &bookpb.Book{
		BookId:    b.ID,
		Name:      b.Name,
		Authors:   b.Authors,
		Year:      int32(b.Year),
		Status:    b.Status,
		CreatedAt: timestamppb.New(b.CreatedAt),
	}
}

func copyToProto(c *domain.BookCopy) *bookpb.BookCopy {
	return &bookpb.BookCopy{
		ExpId:  c.ExpID,
		BookId: c.BookID,
		Status: c.Status,
	}
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrBookNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrCopyNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInvalidStatus):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

// --- Реализация методов gRPC ---

func (h *BookHandler) CreateBook(ctx context.Context, req *bookpb.CreateBookRequest) (*bookpb.BookResponse, error) {
	b, err := h.uc.Create(ctx, req.GetName(), req.GetAuthors(), int(req.GetYear()))
	if err != nil {
		return nil, mapErr(err)
	}
	return &bookpb.BookResponse{Book: toProto(b)}, nil
}

func (h *BookHandler) GetBookById(ctx context.Context, req *bookpb.GetBookByIdRequest) (*bookpb.BookResponse, error) {
	b, err := h.uc.GetByID(ctx, req.GetBookId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &bookpb.BookResponse{Book: toProto(b)}, nil
}

func (h *BookHandler) UpdateBook(ctx context.Context, req *bookpb.UpdateBookRequest) (*bookpb.BookResponse, error) {
	b, err := h.uc.Update(ctx, req.GetBookId(), req.GetName(), req.GetAuthors(), int(req.GetYear()))
	if err != nil {
		return nil, mapErr(err)
	}
	return &bookpb.BookResponse{Book: toProto(b)}, nil
}

func (h *BookHandler) DeleteBook(ctx context.Context, req *bookpb.DeleteBookRequest) (*emptypb.Empty, error) {
	if err := h.uc.Delete(ctx, req.GetBookId()); err != nil {
		return nil, mapErr(err)
	}
	return &emptypb.Empty{}, nil
}

func (h *BookHandler) ListBooks(ctx context.Context, req *bookpb.ListBooksRequest) (*bookpb.ListBooksResponse, error) {
	books, total, err := h.uc.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*bookpb.Book, len(books))
	for i, b := range books {
		out[i] = toProto(b)
	}
	return &bookpb.ListBooksResponse{Books: out, Total: int32(total)}, nil
}

func (h *BookHandler) SearchBooks(ctx context.Context, req *bookpb.SearchBooksRequest) (*bookpb.ListBooksResponse, error) {
	books, total, err := h.uc.Search(ctx, req.GetQuery(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*bookpb.Book, len(books))
	for i, b := range books {
		out[i] = toProto(b)
	}
	return &bookpb.ListBooksResponse{Books: out, Total: int32(total)}, nil
}

func (h *BookHandler) AddBookCopy(ctx context.Context, req *bookpb.AddBookCopyRequest) (*bookpb.BookCopyResponse, error) {
	c, err := h.uc.AddCopy(ctx, req.GetBookId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &bookpb.BookCopyResponse{Copy: copyToProto(c)}, nil
}

func (h *BookHandler) GetBookCopies(ctx context.Context, req *bookpb.GetBookCopiesRequest) (*bookpb.ListBookCopiesResponse, error) {
	copies, err := h.uc.GetBookCopies(ctx, req.GetBookId())
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*bookpb.BookCopy, len(copies))
	for i, c := range copies {
		out[i] = copyToProto(c)
	}
	return &bookpb.ListBookCopiesResponse{Copies: out}, nil
}

func (h *BookHandler) UpdateCopyStatus(ctx context.Context, req *bookpb.UpdateCopyStatusRequest) (*bookpb.BookCopyResponse, error) {
	c, err := h.uc.UpdateCopyStatus(ctx, req.GetExpId(), req.GetStatus())
	if err != nil {
		return nil, mapErr(err)
	}
	return &bookpb.BookCopyResponse{Copy: copyToProto(c)}, nil
}

func (h *BookHandler) GetAvailableCopies(ctx context.Context, req *bookpb.GetAvailableCopiesRequest) (*bookpb.ListBookCopiesResponse, error) {
	// ИСПРАВЛЕНО: Теперь типы должны совпадать после правки .proto
	copies, err := h.uc.GetAvailableCopies(ctx, req.GetBookId())
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*bookpb.BookCopy, len(copies))
	for i, c := range copies {
		out[i] = copyToProto(c)
	}
	return &bookpb.ListBookCopiesResponse{Copies: out}, nil
}

func (h *BookHandler) GetBooksByAuthor(ctx context.Context, req *bookpb.GetBooksByAuthorRequest) (*bookpb.ListBooksResponse, error) {
	books, total, err := h.uc.ListByAuthor(ctx, req.GetAuthor(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*bookpb.Book, len(books))
	for i, b := range books {
		out[i] = toProto(b)
	}
	return &bookpb.ListBooksResponse{Books: out, Total: int32(total)}, nil
}

func (h *BookHandler) GetBooksByYear(ctx context.Context, req *bookpb.GetBooksByYearRequest) (*bookpb.ListBooksResponse, error) {
	books, total, err := h.uc.ListByYear(ctx, int(req.GetYear()), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*bookpb.Book, len(books))
	for i, b := range books {
		out[i] = toProto(b)
	}
	return &bookpb.ListBooksResponse{Books: out, Total: int32(total)}, nil
}
