package handler

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteErr(w http.ResponseWriter, code int, msg string) {
	WriteJSON(w, code, map[string]string{"error": msg})
}

func WriteGrpcErr(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		WriteErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	switch st.Code() {
	case codes.NotFound:
		WriteErr(w, http.StatusNotFound, st.Message())
	case codes.AlreadyExists:
		WriteErr(w, http.StatusConflict, st.Message())
	case codes.InvalidArgument:
		WriteErr(w, http.StatusBadRequest, st.Message())
	case codes.Unauthenticated:
		WriteErr(w, http.StatusUnauthorized, st.Message())
	case codes.PermissionDenied:
		WriteErr(w, http.StatusForbidden, st.Message())
	default:
		WriteErr(w, http.StatusInternalServerError, st.Message())
	}
}
