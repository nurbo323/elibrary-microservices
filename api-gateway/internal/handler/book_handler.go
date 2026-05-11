package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"elibrary/gen/bookpb"
)

type BookHandler struct{ client bookpb.BookServiceClient }

func NewBookHandler(c bookpb.BookServiceClient) *BookHandler { return &BookHandler{client: c} }

func (h *BookHandler) Register(r chi.Router) {
	r.Post("/books", h.create)
	r.Get("/books/{id}", h.getByID)
	r.Put("/books/{id}", h.update)
	r.Delete("/books/{id}", h.delete)
	r.Get("/books", h.list)
	r.Post("/books/{id}/copies", h.addCopy)
}

type createBookReq struct {
	Name    string   `json:"name"`
	Authors []string `json:"authors"`
	Year    int      `json:"year"`
}

func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createBookReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	resp, err := h.client.CreateBook(r.Context(), &bookpb.CreateBookRequest{
		Name: req.Name, Authors: req.Authors, Year: int32(req.Year),
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, resp.GetBook())
}

func (h *BookHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.client.GetBookById(r.Context(), &bookpb.GetBookByIdRequest{BookId: id})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp.GetBook())
}

type updateBookReq struct {
	Name    string   `json:"name"`
	Authors []string `json:"authors"`
	Year    int      `json:"year"`
}

func (h *BookHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req updateBookReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	resp, err := h.client.UpdateBook(r.Context(), &bookpb.UpdateBookRequest{
		BookId: id, Name: req.Name, Authors: req.Authors, Year: int32(req.Year),
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp.GetBook())
}

func (h *BookHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := h.client.DeleteBook(r.Context(), &bookpb.DeleteBookRequest{BookId: id}); err != nil {
		WriteGrpcErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BookHandler) list(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	resp, err := h.client.ListBooks(r.Context(), &bookpb.ListBooksRequest{
		Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (h *BookHandler) addCopy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.client.AddBookCopy(r.Context(), &bookpb.AddBookCopyRequest{BookId: id})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, resp.GetCopy())
}