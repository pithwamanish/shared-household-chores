package handlers

import (
	"bytes"
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
		// For new users in public alpha, auto-provision a personal household so magic link works seamlessly
		userName := strings.Title(strings.Split(cleanEmail, "@")[0])
		regReq := models.RegisterRequest{
			Name:          userName,
			Email:         cleanEmail,
			Password:      "supabase-managed-" + time.Now().Format("20060102150405"),
			HouseholdName: fmt.Sprintf("%s's Home", userName),
			Role:          models.MemberRoleAdmin,
		}
		_, _, regErr := h.Store.RegisterMember(regReq)
		if regErr == nil {
			token, member, _, err = h.Store.CreateMagicLink(cleanEmail)
		}
		if err != nil {
			respondError(w, "RESOURCE_NOT_FOUND", fmt.Sprintf("No member found with email '%s'", cleanEmail), http.StatusNotFound, nil)
			return
		}
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
		"message": fmt.Sprintf("Magic login link sent to %s. Please check your inbox.", cleanEmail),
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

	localValid := false
	if err == nil && member != nil {
		localValid = h.Store.VerifyPassword(member.ID, cleanPassword)
	}

	// If local check failed, check if Supabase Auth can verify this email & password
	if !localValid {
		supabaseURL := strings.TrimSpace(os.Getenv("SUPABASE_URL"))
		anonKey := strings.TrimSpace(os.Getenv("SUPABASE_ANON_KEY"))
		if supabaseURL != "" && anonKey != "" && !strings.Contains(supabaseURL, "your_supabase") {
			tokenEndpoint := strings.TrimRight(supabaseURL, "/")
			if !strings.HasSuffix(tokenEndpoint, "/auth/v1") {
				tokenEndpoint += "/auth/v1"
			}
			tokenEndpoint += "/token?grant_type=password"

			bodyBytes, _ := json.Marshal(map[string]string{
				"email":    cleanEmail,
				"password": cleanPassword,
			})

			client := &http.Client{Timeout: 5 * time.Second}
			vReq, err := http.NewRequest(http.MethodPost, tokenEndpoint, bytes.NewReader(bodyBytes))
			if err == nil {
				vReq.Header.Set("apikey", anonKey)
				vReq.Header.Set("Content-Type", "application/json")
				resp, vErr := client.Do(vReq)
				if vErr == nil {
					defer resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						var sbResp struct {
							AccessToken string `json:"access_token"`
							User        struct {
								ID           string            `json:"id"`
								Email        string            `json:"email"`
								UserMetadata map[string]string `json:"user_metadata"`
							} `json:"user"`
						}
						if err := json.NewDecoder(resp.Body).Decode(&sbResp); err == nil && sbResp.User.Email != "" {
							if member == nil {
								name := sbResp.User.UserMetadata["name"]
								if name == "" {
									name = strings.Title(strings.Split(cleanEmail, "@")[0])
								}
								regReq := models.RegisterRequest{
									Name:          name,
									Email:         cleanEmail,
									Password:      cleanPassword,
									HouseholdName: fmt.Sprintf("%s's Home", name),
									Role:          models.MemberRoleAdmin,
								}
								member, household, _ = h.Store.RegisterMember(regReq)
							} else {
								_ = h.Store.UpdatePassword(member.ID, cleanPassword)
							}
							localValid = true
						}
					}
				}
			}
		}
	}

	if !localValid || member == nil || household == nil {
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

// SupabaseLoginRequest holds the payload sent from frontend after Supabase Auth verification.
type SupabaseLoginRequest struct {
	AccessToken string `json:"access_token"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	Password    string `json:"password,omitempty"`
}

// SupabaseLogin handles sign-in or zero-friction auto-provisioning for users authenticated via Supabase.
func (h *AuthHandler) SupabaseLogin(w http.ResponseWriter, r *http.Request) {
	var req SupabaseLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest, nil)
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	if cleanEmail == "" {
		respondError(w, "BAD_REQUEST", "Email is required", http.StatusBadRequest, nil)
		return
	}

	// Verify token with Supabase if Supabase URL and AccessToken are present
	supabaseURL := strings.TrimSpace(os.Getenv("SUPABASE_URL"))
	anonKey := strings.TrimSpace(os.Getenv("SUPABASE_ANON_KEY"))
	if supabaseURL != "" && req.AccessToken != "" && !strings.Contains(supabaseURL, "your_supabase") {
		verifyEndpoint := strings.TrimRight(supabaseURL, "/")
		if !strings.HasSuffix(verifyEndpoint, "/auth/v1") {
			verifyEndpoint += "/auth/v1"
		}
		verifyEndpoint += "/user"

		client := &http.Client{Timeout: 5 * time.Second}
		vReq, err := http.NewRequest(http.MethodGet, verifyEndpoint, nil)
		if err == nil {
			vReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
			if anonKey != "" {
				vReq.Header.Set("apikey", anonKey)
			}
			resp, vErr := client.Do(vReq)
			if vErr == nil {
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					var sbUser struct {
						ID           string            `json:"id"`
						Email        string            `json:"email"`
						UserMetadata map[string]string `json:"user_metadata"`
					}
					if err := json.NewDecoder(resp.Body).Decode(&sbUser); err == nil && sbUser.Email != "" {
						cleanEmail = strings.ToLower(strings.TrimSpace(sbUser.Email))
						if metaName, ok := sbUser.UserMetadata["name"]; ok && metaName != "" && req.Name == "" {
							req.Name = metaName
						}
					}
				}
			}
		}
	}

	// Retrieve existing member or auto-provision a new household
	member, household, err := h.Store.GetMemberByEmail(cleanEmail)
	if err != nil {
		userName := strings.TrimSpace(req.Name)
		if userName == "" {
			parts := strings.Split(cleanEmail, "@")
			userName = strings.Title(parts[0])
		}
		regReq := models.RegisterRequest{
			Name:          userName,
			Email:         cleanEmail,
			Password:      "supabase-managed-" + time.Now().Format("20060102150405"),
			HouseholdName: fmt.Sprintf("%s's Home", userName),
			Role:          models.MemberRoleAdmin,
		}
		var regErr error
		member, household, regErr = h.Store.RegisterMember(regReq)
		if regErr != nil {
			respondError(w, "INTERNAL_ERROR", fmt.Sprintf("Failed to auto-provision user '%s': %v", cleanEmail, regErr), http.StatusInternalServerError, nil)
			return
		}
	} else if req.Password != "" {
		_ = h.Store.UpdatePassword(member.ID, req.Password)
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


