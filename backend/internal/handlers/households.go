package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/choresync/backend/internal/models"
	"github.com/choresync/backend/internal/store"
	"github.com/choresync/backend/internal/telemetry"
)

type HouseholdHandler struct {
	Store store.Store
}

func NewHouseholdHandler(s store.Store) *HouseholdHandler {
	return &HouseholdHandler{Store: s}
}

// GetHouseholds returns all households.
func (h *HouseholdHandler) GetHouseholds(w http.ResponseWriter, r *http.Request) {
	households := h.Store.GetHouseholds()
	respondJSON(w, http.StatusOK, households)
}

// CreateHousehold creates a new household workspace.
func (h *HouseholdHandler) CreateHousehold(w http.ResponseWriter, r *http.Request) {
	var req models.HouseholdCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		respondError(w, "BAD_REQUEST", "Name is required and cannot be empty", http.StatusBadRequest, nil)
		return
	}

	household := h.Store.CreateHousehold(req)

	telemetry.RecordHouseholdCreated(r.Context(), string(household.Settings.DefaultMode))
	telemetry.Info(r.Context(), "Household created", "household_id", household.ID, "mode", household.Settings.DefaultMode)
	span := trace.SpanFromContext(r.Context())
	if span != nil {
		span.SetAttributes(
			attribute.String("household.id", household.ID),
			attribute.String("household.mode", string(household.Settings.DefaultMode)),
		)
	}

	respondJSON(w, http.StatusCreated, household)
}

// JoinHousehold joins using an invite code.
func (h *HouseholdHandler) JoinHousehold(w http.ResponseWriter, r *http.Request) {
	var req models.HouseholdJoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.InviteCode) == "" {
		respondError(w, "BAD_REQUEST", "Invite code is required", http.StatusBadRequest, nil)
		return
	}

	household, err := h.Store.JoinHousehold(req.InviteCode)
	if err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Household with invite code '%s' not found", req.InviteCode), http.StatusNotFound, nil)
		return
	}

	respondJSON(w, http.StatusOK, household)
}

// GetHousehold retrieves a household by its ID.
func (h *HouseholdHandler) GetHousehold(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "household_id")
	household, err := h.Store.GetHousehold(id)
	if err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Household with ID '%s' not found", id), http.StatusNotFound, nil)
		return
	}

	respondJSON(w, http.StatusOK, household)
}

// UpdateHouseholdSettings updates configuration preferences.
func (h *HouseholdHandler) UpdateHouseholdSettings(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "household_id")
	var req models.HouseholdSettingsUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if req.AdminPIN != nil {
		cleanPin := strings.TrimSpace(*req.AdminPIN)
		if len(cleanPin) != 4 {
			respondError(w, "BAD_REQUEST", "Admin PIN must be exactly 4 digits", http.StatusBadRequest, nil)
			return
		}
		for _, c := range cleanPin {
			if c < '0' || c > '9' {
				respondError(w, "BAD_REQUEST", "Admin PIN must contain only digits (0-9)", http.StatusBadRequest, nil)
				return
			}
		}
		req.AdminPIN = &cleanPin
	}

	household, err := h.Store.UpdateHouseholdSettings(id, req)
	if err != nil {
		handleStoreError(w, err, fmt.Sprintf("Household with ID '%s' not found", id))
		return
	}

	respondJSON(w, http.StatusOK, household)
}
