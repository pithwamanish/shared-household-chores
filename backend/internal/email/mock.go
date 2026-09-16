package email

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/choresync/backend/internal/models"
)

// MockService implements Service in-memory for testing and local development.
type MockService struct {
	mu     sync.RWMutex
	emails []models.DevEmail
	cfg    Config
}

// NewMockService creates a new mock/dev email service.
func NewMockService(cfg Config) *MockService {
	return &MockService{
		emails: make([]models.DevEmail, 0),
		cfg:    cfg,
	}
}

func (m *MockService) recordEmail(toEmail, toName, subject, textBody, htmlBody, emailType, token string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("email-%d", time.Now().UnixNano())
	devEmail := models.DevEmail{
		ID:       id,
		To:       toEmail,
		ToName:   toName,
		From:     m.cfg.FromAddress,
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
		Type:     emailType,
		SentAt:   time.Now().Format(time.RFC3339),
		Provider: "mock",
		Token:    token,
	}

	m.emails = append(m.emails, devEmail)
	if len(m.emails) > 100 {
		m.emails = m.emails[len(m.emails)-100:]
	}

	log.Printf("[EMAIL][MOCK] To: %s <%s> | Subject: %s | Type: %s | Token: %s", toName, toEmail, subject, emailType, token)
}

// SendMagicLink records and logs a magic login link email.
func (m *MockService) SendMagicLink(toEmail, toName, magicLink, token string) error {
	subject, textBody, htmlBody := RenderMagicLinkEmail(toName, magicLink, token)
	m.recordEmail(toEmail, toName, subject, textBody, htmlBody, "magic_link", token)
	return nil
}

// SendPasswordReset records and logs a password reset email.
func (m *MockService) SendPasswordReset(toEmail, toName, resetLink, token string) error {
	subject, textBody, htmlBody := RenderPasswordResetEmail(toName, resetLink, token)
	m.recordEmail(toEmail, toName, subject, textBody, htmlBody, "password_reset", token)
	return nil
}

// SendChoreReminder records and logs a chore nudge reminder email.
func (m *MockService) SendChoreReminder(toEmail, toName, senderName, choreTitle, choreDueDate, householdName string) error {
	subject, textBody, htmlBody := RenderChoreReminderEmail(toName, senderName, choreTitle, choreDueDate, householdName)
	m.recordEmail(toEmail, toName, subject, textBody, htmlBody, "chore_reminder", "")
	return nil
}

// GetRecentDevEmails returns copies of all recorded dev emails in reverse-chronological order.
func (m *MockService) GetRecentDevEmails() []models.DevEmail {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]models.DevEmail, len(m.emails))
	for i, e := range m.emails {
		res[len(m.emails)-1-i] = e
	}
	return res
}

// ClearDevEmails empties the dev email log buffer.
func (m *MockService) ClearDevEmails() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.emails = make([]models.DevEmail, 0)
}
