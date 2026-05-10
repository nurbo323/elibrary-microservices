package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"elibrary/gen/userpb"
)

type UserHandler struct{ client userpb.UserServiceClient }

func NewUserHandler(c userpb.UserServiceClient) *UserHandler { return &UserHandler{client: c} }

func (h *UserHandler) Register(r chi.Router) {
	r.Get("/users/_ping", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "user-handler stub"})
	})
}
