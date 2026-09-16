package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/choresync/backend/internal/auth"
	"github.com/choresync/backend/internal/email"
	"github.com/choresync/backend/internal/models"
	"github.com/choresync/backend/internal/store"
)

// AuthHandler handles authentication, magic links, demo tokens, and PIN verification.
type AuthHandler struct {
	Store store.Store
	Email email.Service
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(s store.Store, em email.Service) *AuthHandler {
	return &AuthHandler{Store: s, Email: em}
}

// RequestMagicLink dispatches or generates a magic login link for an email address.
func (h *AuthHandler) RequestMagicLink(w http.ResponseWriter, r *http.Request) {
	var req models.MagicLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	if cleanEmail == "" {
		respondError(w, "BAD_REQUEST", "Email address is required", http.StatusBadRequest, nil)
		return
	}

	token, member, _, err := h.Store.CreateMagicLink(cleanEmail)
	if err != nil {
		// For demo / test convenience, if email not found, give helpful message
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("No member found with email '%s'", cleanEmail), http.StatusNotFound, nil)
		return
	}

	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	magicLink := fmt.Sprintf("%s/?magic_token=%s", baseURL, token)

	memberName := ""
	if member != nil {
		memberName = member.Name
	}

	if h.Email != nil {
		_ = h.Email.SendMagicLink(cleanEmail, memberName, magicLink, token)
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Magic login code for %s: %s", cleanEmail, token),
		"token":   token,
		"link":    magicLink,
	})
}

// VerifyMagicLink verifies a token and returns a signed JWT with member and household claims.
func (h *AuthHandler) VerifyMagicLink(w http.ResponseWriter, r *http.Request) {
	var req models.MagicLinkVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.Token) == "" {
		respondError(w, "BAD_REQUEST", "Verification token is required", http.StatusBadRequest, nil)
		return
	}

	member, household, err := h.Store.VerifyMagicLink(req.Token)
	if err != nil {
		respondError(w, "UNAUTHORIZED", "Invalid or expired verification token", http.StatusUnauthorized, nil)
		return
	}

	claims := auth.Claims{
		Sub:         member.ID,
		HouseholdID: member.HouseholdID,
		Name:        member.Name,
		Role:        member.Role,
		Exp:         time.Now().Add(7 * 24 * time.Hour).Unix(),
		Iat:         time.Now().Unix(),
	}

	jwtToken, err := auth.GenerateToken(claims, auth.GetDefaultSecret())
	if err != nil {
		respondError(w, "INTERNAL_ERROR", "Failed to generate access token", http.StatusInternalServerError, nil)
		return
	}

	resp := models.AuthTokenResponse{
		AccessToken: jwtToken,
		TokenType:   "Bearer",
		ExpiresIn:   7 * 24 * 3600,
		Member:      member,
		Household:   household,
	}

	respondJSON(w, http.StatusOK, resp)
}

// DemoLogin mints a real signed JWT for seed demo personas (instant 1-click test switching).
func (h *AuthHandler) DemoLogin(w http.ResponseWriter, r *http.Request) {
	var req models.DemoLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	if strings.TrimSpace(req.MemberID) == "" {
		respondError(w, "BAD_REQUEST", "member_id is required", http.StatusBadRequest, nil)
		return
	}

	member, household, err := h.Store.GetDemoMember(req.MemberID)
	if err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Demo member with ID '%s' not found", req.MemberID), http.StatusNotFound, nil)
		return
	}

	claims := auth.Claims{
		Sub:         member.ID,
		HouseholdID: member.HouseholdID,
		Name:        member.Name,
		Role:        member.Role,
		Exp:         time.Now().Add(7 * 24 * time.Hour).Unix(),
		Iat:         time.Now().Unix(),
	}

	jwtToken, err := auth.GenerateToken(claims, auth.GetDefaultSecret())
	if err != nil {
		respondError(w, "INTERNAL_ERROR", "Failed to generate access token", http.StatusInternalServerError, nil)
		return
	}

	resp := models.AuthTokenResponse{
		AccessToken: jwtToken,
		TokenType:   "Bearer",
		ExpiresIn:   7 * 24 * 3600,
		Member:      member,
		Household:   household,
	}

	respondJSON(w, http.StatusOK, resp)
}

// VerifyPIN verifies the 4-digit household admin PIN.
func (h *AuthHandler) VerifyPIN(w http.ResponseWriter, r *http.Request) {
	householdID := chi.URLParam(r, "household_id")
	if householdID == "" {
		respondError(w, "BAD_REQUEST", "household_id path parameter is required", http.StatusBadRequest, nil)
		return
	}

	var req models.VerifyPINRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	cleanPIN := strings.TrimSpace(req.PIN)
	if cleanPIN == "" {
		respondError(w, "BAD_REQUEST", "4-digit PIN is required", http.StatusBadRequest, nil)
		return
	}

	valid, err := h.Store.VerifyAdminPIN(householdID, cleanPIN)
	if err != nil {
		respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("Household '%s' not found", householdID), http.StatusNotFound, nil)
		return
	}

	if !valid {
		respondError(w, "UNAUTHORIZED", "Invalid 4-digit admin PIN", http.StatusUnauthorized, nil)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"valid":   true,
		"message": "Admin PIN verified",
	})
}

// Login handles email + password authentication.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	cleanPassword := strings.TrimSpace(req.Password)
	if cleanEmail == "" || cleanPassword == "" {
		respondError(w, "BAD_REQUEST", "Email and password are required", http.StatusBadRequest, nil)
		return
	}

	member, household, err := h.Store.GetMemberByEmail(cleanEmail)
	if err != nil {
		respondError(w, "UNAUTHORIZED", "Invalid email or password", http.StatusUnauthorized, nil)
		return
	}

	if !h.Store.VerifyPassword(member.ID, cleanPassword) {
		respondError(w, "UNAUTHORIZED", "Invalid email or password", http.StatusUnauthorized, nil)
		return
	}

	claims := auth.Claims{
		Sub:         member.ID,
		HouseholdID: member.HouseholdID,
		Name:        member.Name,
		Role:        member.Role,
		Exp:         time.Now().Add(7 * 24 * time.Hour).Unix(),
		Iat:         time.Now().Unix(),
	}

	jwtToken, err := auth.GenerateToken(claims, auth.GetDefaultSecret())
	if err != nil {
		respondError(w, "INTERNAL_ERROR", "Failed to generate access token", http.StatusInternalServerError, nil)
		return
	}

	resp := models.AuthTokenResponse{
		AccessToken: jwtToken,
		TokenType:   "Bearer",
		ExpiresIn:   7 * 24 * 3600,
		Member:      member,
		Household:   household,
	}

	respondJSON(w, http.StatusOK, resp)
}

// Register creates a new user account with email and password and creates or joins a household.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	cleanName := strings.TrimSpace(req.Name)
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	cleanPassword := strings.TrimSpace(req.Password)
	if cleanName == "" || cleanEmail == "" || cleanPassword == "" {
		respondError(w, "BAD_REQUEST", "Name, email, and password are required", http.StatusBadRequest, nil)
		return
	}

	member, household, err := h.Store.RegisterMember(req)
	if err != nil {
		if strings.Contains(err.Error(), "already registered") {
			respondError(w, "CONFLICT", err.Error(), http.StatusConflict, nil)
			return
		}
		respondError(w, "BAD_REQUEST", err.Error(), http.StatusBadRequest, nil)
		return
	}

	claims := auth.Claims{
		Sub:         member.ID,
		HouseholdID: member.HouseholdID,
		Name:        member.Name,
		Role:        member.Role,
		Exp:         time.Now().Add(7 * 24 * time.Hour).Unix(),
		Iat:         time.Now().Unix(),
	}

	jwtToken, err := auth.GenerateToken(claims, auth.GetDefaultSecret())
	if err != nil {
		respondError(w, "INTERNAL_ERROR", "Failed to generate access token", http.StatusInternalServerError, nil)
		return
	}

	resp := models.AuthTokenResponse{
		AccessToken: jwtToken,
		TokenType:   "Bearer",
		ExpiresIn:   7 * 24 * 3600,
		Member:      member,
		Household:   household,
	}

	respondJSON(w, http.StatusCreated, resp)
}

// RequestPasswordReset issues a password reset token and dispatches an email.
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	if cleanEmail == "" {
		respondError(w, "BAD_REQUEST", "Email address is required", http.StatusBadRequest, nil)
		return
	}

	token, member, err := h.Store.CreatePasswordResetToken(cleanEmail)
	if err != nil {
		// Avoid leaking account existence; return success message
		respondJSON(w, http.StatusOK, models.ForgotPasswordResponse{
			Message: fmt.Sprintf("If an account exists for %s, a password reset link has been sent.", cleanEmail),
		})
		return
	}

	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	resetLink := fmt.Sprintf("%s/?reset_token=%s", baseURL, token)

	if h.Email != nil {
		_ = h.Email.SendPasswordReset(cleanEmail, member.Name, resetLink, token)
	}

	respondJSON(w, http.StatusOK, models.ForgotPasswordResponse{
		Message: fmt.Sprintf("Password reset link sent to %s", cleanEmail),
		Token:   token,
	})
}

// ResetPassword verifies a password reset token and updates the password.
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	cleanToken := strings.TrimSpace(req.Token)
	cleanPassword := strings.TrimSpace(req.NewPassword)
	if cleanToken == "" || cleanPassword == "" {
		respondError(w, "BAD_REQUEST", "Reset token and new password are required", http.StatusBadRequest, nil)
		return
	}

	if len(cleanPassword) < 6 {
		respondError(w, "BAD_REQUEST", "Password must be at least 6 characters", http.StatusBadRequest, nil)
		return
	}

	member, household, err := h.Store.ResetPasswordWithToken(cleanToken, cleanPassword)
	if err != nil {
		respondError(w, "UNAUTHORIZED", "Invalid or expired password reset token", http.StatusUnauthorized, nil)
		return
	}

	claims := auth.Claims{
		Sub:         member.ID,
		HouseholdID: member.HouseholdID,
		Name:        member.Name,
		Role:        member.Role,
		Exp:         time.Now().Add(7 * 24 * time.Hour).Unix(),
		Iat:         time.Now().Unix(),
	}

	jwtToken, err := auth.GenerateToken(claims, auth.GetDefaultSecret())
	if err != nil {
		respondError(w, "INTERNAL_ERROR", "Failed to generate access token", http.StatusInternalServerError, nil)
		return
	}

	resp := models.AuthTokenResponse{
		AccessToken: jwtToken,
		TokenType:   "Bearer",
		ExpiresIn:   7 * 24 * 3600,
		Member:      member,
		Household:   household,
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetDevEmails returns the captured dev/test emails.
func (h *AuthHandler) GetDevEmails(w http.ResponseWriter, r *http.Request) {
	if h.Email == nil {
		respondJSON(w, http.StatusOK, []models.DevEmail{})
		return
	}
	respondJSON(w, http.StatusOK, h.Email.GetRecentDevEmails())
}

// ClearDevEmails clears the captured dev/test emails buffer.
func (h *AuthHandler) ClearDevEmails(w http.ResponseWriter, r *http.Request) {
	if h.Email != nil {
		h.Email.ClearDevEmails()
	}
	respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

