package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/choresync/backend/internal/models"
	"github.com/choresync/backend/internal/store"
)

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, errCode string, message string, httpStatus int, details map[string]any) {
	if details == nil {
		details = map[string]any{}
	}
	resp := models.ErrorResponse{
		Error:   errCode,
		Message: message,
		Code:    httpStatus,
		Details: details,
	}
	respondJSON(w, httpStatus, resp)
}

func handleStoreError(w http.ResponseWriter, err error, notFoundMsg string) {
	if errors.Is(err, store.ErrNotFound) {
		respondError(w, "RESOURCE_NOT_FOUND", notFoundMsg, http.StatusNotFound, nil)
		return
	}
	if errors.Is(err, store.ErrForbidden) {
		respondError(w, "FORBIDDEN", err.Error(), http.StatusForbidden, nil)
		return
	}
	if errors.Is(err, store.ErrBadRequest) {
		respondError(w, "BAD_REQUEST", err.Error(), http.StatusBadRequest, nil)
		return
	}
	respondError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError, nil)
}
