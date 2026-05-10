package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"elibrary/gen/borrowpb"
)

type BorrowHandler struct{ client borrowpb.BorrowServiceClient }

func NewBorrowHandler(c borrowpb.BorrowServiceClient) *BorrowHandler {
	return &BorrowHandler{client: c}
}

func (h *BorrowHandler) Register(r chi.Router) {
	r.Get("/borrows/_ping", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "borrow-handler stub"})
	})
}
