package handlers

import (
	"net/http"

	"github.com/choresync/backend/internal/models"
	"github.com/choresync/backend/internal/store"
)

type SystemHandler struct {
	Store store.Store
}

func NewSystemHandler(s store.Store) *SystemHandler {
	return &SystemHandler{Store: s}
}

// ResetDemoData resets data back to initial seed state.
func (h *SystemHandler) ResetDemoData(w http.ResponseWriter, r *http.Request) {
	h.Store.Reset()
	respondJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "System reset successfully to initial demo seed state",
	})
}

// HealthCheck provides uptime status for platform health probes and observability monitoring.
func (h *SystemHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"version": "1.0.0",
		"system":  "ChoreSync API",
	})
}
