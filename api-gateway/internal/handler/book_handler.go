package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"elibrary/gen/bookpb"
)

type BookHandler struct{ client bookpb.BookServiceClient }

func NewBookHandler(c bookpb.BookServiceClient) *BookHandler { return &BookHandler{client: c} }

func (h *BookHandler) Register(r chi.Router) {
	r.Get("/books/_ping", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "book-handler stub"})
	})
}
