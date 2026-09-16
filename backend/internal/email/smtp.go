package email

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"github.com/choresync/backend/internal/models"
)

// SMTPService sends emails using standard SMTP (compatible with SendGrid, Mailgun, Brevo, AWS SES, Gmail).
type SMTPService struct {
	cfg     Config
	devMock *MockService
}

// NewSMTPService creates a new SMTP service.
func NewSMTPService(cfg Config) *SMTPService {
	return &SMTPService{
		cfg:     cfg,
		devMock: NewMockService(cfg),
	}
}

func (s *SMTPService) sendViaSMTP(toEmail, toName, subject, textBody, htmlBody, emailType, token string) error {
	s.devMock.recordEmail(toEmail, toName, subject, textBody, htmlBody, emailType, token)

	port := s.cfg.SMTPPort
	if port == "" {
		port = "587"
	}
	addr := fmt.Sprintf("%s:%s", s.cfg.SMTPHost, port)

	var auth smtp.Auth
	if s.cfg.SMTPUser != "" && s.cfg.SMTPPass != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	}

	fromAddr := s.cfg.FromAddress
	if strings.Contains(fromAddr, "<") && strings.Contains(fromAddr, ">") {
		start := strings.Index(fromAddr, "<")
		end := strings.Index(fromAddr, ">")
		if start < end {
			fromAddr = fromAddr[start+1 : end]
		}
	}

	boundary := "ChoreSyncBoundary789456"
	header := make(map[string]string)
	header["From"] = s.cfg.FromAddress
	header["To"] = fmt.Sprintf("%s <%s>", toName, toEmail)
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = fmt.Sprintf("multipart/alternative; boundary=%s", boundary)

	var msg strings.Builder
	for k, v := range header {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")

	// Text part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
	msg.WriteString(textBody)
	msg.WriteString("\r\n\r\n")

	// HTML part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n\r\n")

	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	err := smtp.SendMail(addr, auth, fromAddr, []string{toEmail}, []byte(msg.String()))
	if err != nil {
		log.Printf("[EMAIL][SMTP] Delivery error to %s: %v", toEmail, err)
		return fmt.Errorf("smtp send failed: %w", err)
	}

	log.Printf("[EMAIL][SMTP] Successfully delivered email to %s via %s", toEmail, addr)
	return nil
}

// SendMagicLink sends a magic link email via SMTP.
func (s *SMTPService) SendMagicLink(toEmail, toName, magicLink, token string) error {
	subject, textBody, htmlBody := RenderMagicLinkEmail(toName, magicLink, token)
	return s.sendViaSMTP(toEmail, toName, subject, textBody, htmlBody, "magic_link", token)
}

// SendPasswordReset sends a password reset email via SMTP.
func (s *SMTPService) SendPasswordReset(toEmail, toName, resetLink, token string) error {
	subject, textBody, htmlBody := RenderPasswordResetEmail(toName, resetLink, token)
	return s.sendViaSMTP(toEmail, toName, subject, textBody, htmlBody, "password_reset", token)
}

// SendChoreReminder sends a chore reminder nudge via SMTP.
func (s *SMTPService) SendChoreReminder(toEmail, toName, senderName, choreTitle, choreDueDate, householdName string) error {
	subject, textBody, htmlBody := RenderChoreReminderEmail(toName, senderName, choreTitle, choreDueDate, householdName)
	return s.sendViaSMTP(toEmail, toName, subject, textBody, htmlBody, "chore_reminder", "")
}

// GetRecentDevEmails delegates to internal mock tracker.
func (s *SMTPService) GetRecentDevEmails() []models.DevEmail {
	return s.devMock.GetRecentDevEmails()
}

// ClearDevEmails clears internal mock tracker buffer.
func (s *SMTPService) ClearDevEmails() {
	s.devMock.ClearDevEmails()
}
