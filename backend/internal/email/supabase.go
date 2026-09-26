package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/choresync/backend/internal/models"
)

// SupabaseService dispatches transactional auth emails via the Supabase Auth (GoTrue) API.
// This allows sending magic links and password resets using Supabase's pre-configured
// zero-domain mailer (noreply@mail.app.supabase.io) without requiring custom domain DNS.
type SupabaseService struct {
	cfg        Config
	httpClient *http.Client
	devMock    *MockService
}

// NewSupabaseService creates a new Supabase email service client.
func NewSupabaseService(cfg Config) *SupabaseService {
	return &SupabaseService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		devMock: NewMockService(cfg),
	}
}

func (s *SupabaseService) getAuthEndpoint(path string) string {
	base := strings.TrimRight(s.cfg.SupabaseURL, "/")
	if !strings.HasSuffix(base, "/auth/v1") {
		base = base + "/auth/v1"
	}
	return base + path
}

func (s *SupabaseService) postToSupabase(endpoint string, payload interface{}) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode supabase payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create supabase request: %w", err)
	}

	apiKey := s.cfg.SupabaseAnonKey
	if s.cfg.SupabaseServiceKey != "" {
		apiKey = s.cfg.SupabaseServiceKey
	}

	req.Header.Set("apikey", apiKey)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ChoreSync-Backend/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("supabase network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase api error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// SendMagicLink dispatches an OTP / Magic Link email via Supabase Auth.
func (s *SupabaseService) SendMagicLink(toEmail, toName, magicLink, token string) error {
	subject, textBody, htmlBody := RenderMagicLinkEmail(toName, magicLink, token)
	s.devMock.recordEmail(toEmail, toName, subject, textBody, htmlBody, "magic_link", token)

	endpoint := s.getAuthEndpoint("/otp")
	payload := map[string]interface{}{
		"email":       toEmail,
		"create_user": true,
		"data": map[string]string{
			"name":       toName,
			"magic_link": magicLink,
		},
	}

	err := s.postToSupabase(endpoint, payload)
	if err != nil {
		log.Printf("[EMAIL][SUPABASE] Delivery error to %s: %v", toEmail, err)
		return err
	}

	log.Printf("[EMAIL][SUPABASE] Successfully dispatched Magic Link to %s via Supabase Auth mailer", toEmail)
	return nil
}

// SendPasswordReset dispatches a password recovery email via Supabase Auth.
func (s *SupabaseService) SendPasswordReset(toEmail, toName, resetLink, token string) error {
	subject, textBody, htmlBody := RenderPasswordResetEmail(toName, resetLink, token)
	s.devMock.recordEmail(toEmail, toName, subject, textBody, htmlBody, "password_reset", token)

	endpoint := s.getAuthEndpoint("/recover")
	payload := map[string]interface{}{
		"email": toEmail,
	}

	err := s.postToSupabase(endpoint, payload)
	if err != nil {
		log.Printf("[EMAIL][SUPABASE] Delivery error to %s: %v", toEmail, err)
		return err
	}

	log.Printf("[EMAIL][SUPABASE] Successfully dispatched Password Reset to %s via Supabase Auth mailer", toEmail)
	return nil
}

// SendChoreReminder records and logs chore reminders (delegated to devMock buffer and console).
func (s *SupabaseService) SendChoreReminder(toEmail, toName, senderName, choreTitle, choreDueDate, householdName string) error {
	subject, textBody, htmlBody := RenderChoreReminderEmail(toName, senderName, choreTitle, choreDueDate, householdName)
	s.devMock.recordEmail(toEmail, toName, subject, textBody, htmlBody, "chore_reminder", "")
	log.Printf("[EMAIL][SUPABASE] Chore reminder queued for %s (Chore: %s, Due: %s)", toEmail, choreTitle, choreDueDate)
	return nil
}

// GetRecentDevEmails returns copies of all recorded dev emails in reverse-chronological order.
func (s *SupabaseService) GetRecentDevEmails() []models.DevEmail {
	return s.devMock.GetRecentDevEmails()
}

// ClearDevEmails clears the internal mock tracker buffer.
func (s *SupabaseService) ClearDevEmails() {
	s.devMock.ClearDevEmails()
}
