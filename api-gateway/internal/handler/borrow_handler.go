package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"elibrary/gen/borrowpb"
)

type BorrowHandler struct{ client borrowpb.BorrowServiceClient }

func NewBorrowHandler(c borrowpb.BorrowServiceClient) *BorrowHandler { return &BorrowHandler{client: c} }

func (h *BorrowHandler) Register(r chi.Router) {
	r.Post("/borrows", h.borrow)
	r.Get("/borrows/active", h.active)               // ВАЖНО: до /borrows/{id}
	r.Get("/borrows/{id}", h.getByID)
	r.Post("/borrows/{id}/return", h.returnBook)
	r.Get("/borrows", h.list)
	r.Get("/users/{user_id}/borrows", h.userHistory) // история пользователя
}

type borrowReq struct {
	UserID string `json:"user_id"`
	BookID string `json:"book_id"`
}

func (h *BorrowHandler) borrow(w http.ResponseWriter, r *http.Request) {
	var req borrowReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	resp, err := h.client.BorrowBook(r.Context(), &borrowpb.BorrowBookRequest{
		UserId: req.UserID, BookId: req.BookID,
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, resp.GetBorrow())
}

func (h *BorrowHandler) returnBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.client.ReturnBook(r.Context(), &borrowpb.ReturnBookRequest{BorrowId: id})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp.GetBorrow())
}

func (h *BorrowHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.client.GetBorrowById(r.Context(), &borrowpb.GetBorrowByIdRequest{BorrowId: id})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp.GetBorrow())
}

func (h *BorrowHandler) list(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	resp, err := h.client.ListBorrows(r.Context(), &borrowpb.ListBorrowsRequest{
		Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (h *BorrowHandler) userHistory(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	resp, err := h.client.GetUserBorrowHistory(r.Context(), &borrowpb.GetUserBorrowHistoryRequest{
		UserId: userID, Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (h *BorrowHandler) active(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	resp, err := h.client.GetActiveBorrows(r.Context(), &borrowpb.GetActiveBorrowsRequest{
		Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}