package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid or malformed token")
	ErrTokenExpired = errors.New("token has expired")
	ErrUnauthorized = errors.New("unauthorized")
)

type contextKey string

const (
	CallerContextKey contextKey = "choresync_caller"
)

// Claims contains JWT claims representing the authenticated household member.
type Claims struct {
	Sub         string `json:"sub"`          // Member ID
	HouseholdID string `json:"household_id"` // Household ID
	Name        string `json:"name"`         // Member Name
	Role        string `json:"role"`         // Role: admin, member, child
	Exp         int64  `json:"exp"`          // Expiration timestamp
	Iat         int64  `json:"iat"`          // Issued at timestamp
}

// GetDefaultSecret returns the JWT HMAC secret from environment or dev fallback.
func GetDefaultSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "choresync-jwt-dev-secret-key-2026-very-secure"
	}
	return []byte(secret)
}

// GenerateToken creates a signed HMAC-SHA256 JWT for the given claims.
func GenerateToken(claims Claims, secret []byte) (string, error) {
	if len(secret) == 0 {
		secret = GetDefaultSecret()
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	signingInput := encodedHeader + "." + encodedPayload

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(signingInput))
	signature := h.Sum(nil)
	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + encodedSignature, nil
}

// ValidateToken parses and verifies an HMAC-SHA256 JWT.
func ValidateToken(tokenStr string, secret []byte) (*Claims, error) {
	if len(secret) == 0 {
		secret = GetDefaultSecret()
	}

	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(signingInput))
	expectedSignature := h.Sum(nil)
	expectedSignatureEncoded := base64.RawURLEncoding.EncodeToString(expectedSignature)

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignatureEncoded)) {
		return nil, ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}

// WithCaller injects the caller Claims into a context.
func WithCaller(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, CallerContextKey, claims)
}

// GetCaller extracts the caller Claims from a context, if present.
func GetCaller(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(CallerContextKey).(*Claims)
	return claims, ok && claims != nil
}

// Middleware returns Chi-compatible middleware that validates Bearer JWT tokens.
// If an Authorization header is provided, it must be a valid Bearer token.
// If valid, the claims are attached to request context.
func Middleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No token provided - continue (public endpoints or fallback mode)
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				respondAuthError(w, "Invalid Authorization header format. Expected 'Bearer <token>'", http.StatusUnauthorized)
				return
			}

			claims, err := ValidateToken(strings.TrimSpace(parts[1]), secret)
			if err != nil {
				respondAuthError(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
				return
			}

			ctx := WithCaller(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func respondAuthError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":   "UNAUTHORIZED",
		"message": message,
		"code":    status,
		"details": map[string]any{},
	})
}
