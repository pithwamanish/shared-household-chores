package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/choresync/backend/internal/models"
)

// ResendService sends transactional emails via the Resend HTTP API.
type ResendService struct {
	cfg        Config
	httpClient *http.Client
	devMock    *MockService
}

// NewResendService creates a new Resend email service client.
func NewResendService(cfg Config) *ResendService {
	return &ResendService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		devMock: NewMockService(cfg),
	}
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text"`
}

func (r *ResendService) sendViaAPI(toEmail, toName, subject, textBody, htmlBody, emailType, token string) error {
	// Always record in dev memory buffer as well
	r.devMock.recordEmail(toEmail, toName, subject, textBody, htmlBody, emailType, token)

	payload := resendPayload{
		From:    r.cfg.FromAddress,
		To:      []string{toEmail},
		Subject: subject,
		HTML:    htmlBody,
		Text:    textBody,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode resend payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create resend http request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+r.cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ChoreSync-Backend/1.0")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		log.Printf("[EMAIL][RESEND] Delivery network error to %s: %v", toEmail, err)
		return fmt.Errorf("resend network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("[EMAIL][RESEND] Delivery API error (%d): %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("resend api error: status %d: %s", resp.StatusCode, string(respBody))
	}

	log.Printf("[EMAIL][RESEND] Successfully sent email to %s (Subject: %s)", toEmail, subject)
	return nil
}

// SendMagicLink sends a magic link email via Resend.
func (r *ResendService) SendMagicLink(toEmail, toName, magicLink, token string) error {
	subject, textBody, htmlBody := RenderMagicLinkEmail(toName, magicLink, token)
	return r.sendViaAPI(toEmail, toName, subject, textBody, htmlBody, "magic_link", token)
}

// SendPasswordReset sends a password reset email via Resend.
func (r *ResendService) SendPasswordReset(toEmail, toName, resetLink, token string) error {
	subject, textBody, htmlBody := RenderPasswordResetEmail(toName, resetLink, token)
	return r.sendViaAPI(toEmail, toName, subject, textBody, htmlBody, "password_reset", token)
}

// SendChoreReminder sends a chore reminder nudge via Resend.
func (r *ResendService) SendChoreReminder(toEmail, toName, senderName, choreTitle, choreDueDate, householdName string) error {
	subject, textBody, htmlBody := RenderChoreReminderEmail(toName, senderName, choreTitle, choreDueDate, householdName)
	return r.sendViaAPI(toEmail, toName, subject, textBody, htmlBody, "chore_reminder", "")
}

// GetRecentDevEmails delegates to the internal mock tracker.
func (r *ResendService) GetRecentDevEmails() []models.DevEmail {
	return r.devMock.GetRecentDevEmails()
}

// ClearDevEmails clears the internal mock tracker buffer.
func (r *ResendService) ClearDevEmails() {
	r.devMock.ClearDevEmails()
}
