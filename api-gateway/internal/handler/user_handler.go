package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"elibrary/gen/userpb"
)

type UserHandler struct{ client userpb.UserServiceClient }

func NewUserHandler(c userpb.UserServiceClient) *UserHandler { return &UserHandler{client: c} }

func (h *UserHandler) Register(r chi.Router) {
	r.Post("/users", h.create)
	r.Get("/users/{id}", h.getByID)
	r.Put("/users/{id}", h.update)
	r.Delete("/users/{id}", h.delete)
	r.Get("/users", h.list)
	r.Post("/auth/login", h.login)

	r.Get("/users/search", h.search)
	r.Put("/users/{id}/password", h.changePassword)
	r.Get("/auth/verify", h.verifyEmail)
	r.Get("/users/{id}/profile", h.profile)
	r.Get("/users/{id}/active-borrows", h.activeBorrows)
	r.Get("/users/{id}/statistics", h.statistics)
}

type createUserReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	resp, err := h.client.CreateUser(r.Context(), &userpb.CreateUserRequest{
		Name: req.Name, Email: req.Email, Password: req.Password,
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, resp.GetUser())
}

func (h *UserHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.client.GetUserById(r.Context(), &userpb.GetUserByIdRequest{UserId: id})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp.GetUser())
}

type updateUserReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *UserHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req updateUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	resp, err := h.client.UpdateUser(r.Context(), &userpb.UpdateUserRequest{
		UserId: id, Name: req.Name, Email: req.Email,
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp.GetUser())
}

func (h *UserHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := h.client.DeleteUser(r.Context(), &userpb.DeleteUserRequest{UserId: id}); err != nil {
		WriteGrpcErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	resp, err := h.client.ListUsers(r.Context(), &userpb.ListUsersRequest{
		Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	resp, err := h.client.LoginUser(r.Context(), &userpb.LoginUserRequest{
		Email: req.Email, Password: req.Password,
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	resp, err := h.client.SearchUsers(r.Context(), &userpb.SearchUsersRequest{
		Query: q, Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

type changePassReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *UserHandler) changePassword(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req changePassReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if _, err := h.client.ChangePassword(r.Context(), &userpb.ChangePasswordRequest{
		UserId: id, OldPassword: req.OldPassword, NewPassword: req.NewPassword,
	}); err != nil {
		WriteGrpcErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) verifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	_, err := h.client.VerifyEmail(r.Context(), &userpb.VerifyEmailRequest{Token: token})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "email verified successfully"})
}

func (h *UserHandler) profile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.client.GetUserProfile(r.Context(), &userpb.GetUserByIdRequest{UserId: id})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) activeBorrows(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.client.GetUserActiveBorrows(r.Context(), &userpb.GetUserByIdRequest{UserId: id})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) statistics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.client.GetUserStatistics(r.Context(), &userpb.GetUserByIdRequest{UserId: id})
	if err != nil {
		WriteGrpcErr(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, resp)
}
