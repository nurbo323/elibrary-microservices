package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	// ИСПРАВЛЕНО: Путь импорта соответствует твоей структуре
	"elibrary/elibrary/gen/bookpb"
)

type BookHandler struct {
	client bookpb.BookServiceClient
}

func NewBookHandler(c bookpb.BookServiceClient) *BookHandler {
	return &BookHandler{client: c}
}

func (h *BookHandler) Register(r chi.Router) {
	// Day 1 routes
	r.Post("/books", h.create)
	r.Get("/books/{id}", h.getByID)
	r.Put("/books/{id}", h.update)
	r.Delete("/books/{id}", h.delete)
	r.Get("/books", h.list)
	r.Post("/books/{id}/copies", h.addCopy)

	// Day 2 routes (ДОБАВЛЕНО)
	r.Get("/books/search", h.search)
	r.Get("/books/by-author", h.byAuthor)
	r.Get("/books/by-year", h.byYear)
	r.Get("/books/{id}/copies", h.copies)
	r.Get("/books/{id}/available-copies", h.availableCopies)
	r.Put("/copies/{exp_id}/status", h.updateCopyStatus)
}

// --- Обработчики Day 1 ---

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

// --- Обработчики Day 2 (НОВЫЕ) ---

func (h *BookHandler) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	resp, err := h.client.SearchBooks(r.Context(), &bookpb.SearchBooksRequest{
		Query: q, Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (h *BookHandler) byAuthor(w http.ResponseWriter, r *http.Request) {
	author := r.URL.Query().Get("author")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	resp, err := h.client.GetBooksByAuthor(r.Context(), &bookpb.GetBooksByAuthorRequest{
		Author: author, Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (h *BookHandler) byYear(w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	resp, err := h.client.GetBooksByYear(r.Context(), &bookpb.GetBooksByYearRequest{
		Year: int32(year), Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (h *BookHandler) copies(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.client.GetBookCopies(r.Context(), &bookpb.GetBookCopiesRequest{BookId: id})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (h *BookHandler) availableCopies(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.client.GetAvailableCopies(r.Context(), &bookpb.GetAvailableCopiesRequest{BookId: id})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

type updateStatusReq struct {
	Status string `json:"status"`
}

func (h *BookHandler) updateCopyStatus(w http.ResponseWriter, r *http.Request) {
	expID := chi.URLParam(r, "exp_id")
	var req updateStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	resp, err := h.client.UpdateCopyStatus(r.Context(), &bookpb.UpdateCopyStatusRequest{
		ExpId: expID, Status: req.Status,
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp.GetCopy())
}
