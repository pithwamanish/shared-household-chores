package email

import (
	"log"
	"os"
	"strings"

	"github.com/choresync/backend/internal/models"
)

// Service defines the transactional email operations for ChoreSync.
type Service interface {
	SendMagicLink(toEmail, toName, magicLink, token string) error
	SendPasswordReset(toEmail, toName, resetLink, token string) error
	SendChoreReminder(toEmail, toName, senderName, choreTitle, choreDueDate, householdName string) error
	GetRecentDevEmails() []models.DevEmail
	ClearDevEmails()
}

// Config holds email client configuration.
type Config struct {
	Provider    string // "resend", "smtp", "mock"
	FromAddress string // e.g. "ChoreSync <notifications@choresync.app>" or "onboarding@resend.dev"
	AppBaseURL  string // e.g. "http://localhost:3000"

	// Resend specific
	ResendAPIKey string

	// SMTP specific
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string

	// Supabase specific
	SupabaseURL        string
	SupabaseAnonKey    string
	SupabaseServiceKey string
}

// LoadConfigFromEnv reads email configuration from environment variables.
func LoadConfigFromEnv() Config {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("EMAIL_PROVIDER")))
	resendKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	smtpHost := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	from := strings.TrimSpace(os.Getenv("EMAIL_FROM"))
	baseURL := strings.TrimSpace(os.Getenv("APP_BASE_URL"))
	supabaseURL := strings.TrimSpace(os.Getenv("SUPABASE_URL"))
	supabaseAnonKey := strings.TrimSpace(os.Getenv("SUPABASE_ANON_KEY"))
	supabaseServiceKey := strings.TrimSpace(os.Getenv("SUPABASE_SERVICE_ROLE_KEY"))

	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	if from == "" {
		if resendKey != "" {
			from = "ChoreSync <onboarding@resend.dev>"
		} else {
			from = "ChoreSync <notifications@choresync.app>"
		}
	}

	// Auto-detect provider if not explicitly configured
	if provider == "" {
		if supabaseURL != "" && supabaseAnonKey != "" {
			provider = "supabase"
		} else if resendKey != "" {
			provider = "resend"
		} else if smtpHost != "" {
			provider = "smtp"
		} else {
			provider = "mock"
		}
	}

	smtpPass := strings.TrimSpace(os.Getenv("SMTP_PASS"))
	if strings.Contains(smtpHost, "gmail") {
		smtpPass = strings.ReplaceAll(smtpPass, " ", "")
	}

	return Config{
		Provider:           provider,
		FromAddress:        from,
		AppBaseURL:         baseURL,
		ResendAPIKey:       resendKey,
		SMTPHost:           smtpHost,
		SMTPPort:           os.Getenv("SMTP_PORT"),
		SMTPUser:           os.Getenv("SMTP_USER"),
		SMTPPass:           smtpPass,
		SupabaseURL:        supabaseURL,
		SupabaseAnonKey:    supabaseAnonKey,
		SupabaseServiceKey: supabaseServiceKey,
	}
}

// NewServiceFromEnv initializes an email service based on environment variables.
func NewServiceFromEnv() Service {
	cfg := LoadConfigFromEnv()

	switch cfg.Provider {
	case "supabase":
		if cfg.SupabaseURL == "" || cfg.SupabaseAnonKey == "" {
			log.Println("[EMAIL] Warning: SUPABASE_URL or SUPABASE_ANON_KEY is not set. Falling back to Mock email provider.")
			return NewMockService(cfg)
		}
		log.Printf("[EMAIL] Initialized Supabase Auth Email Service (Endpoint: %s)", cfg.SupabaseURL)
		return NewSupabaseService(cfg)

	case "resend":
		if cfg.ResendAPIKey == "" {
			log.Println("[EMAIL] Warning: RESEND_API_KEY is not set. Falling back to Mock email provider.")
			return NewMockService(cfg)
		}
		log.Printf("[EMAIL] Initialized Resend Email Service (Sender: %s)", cfg.FromAddress)
		return NewResendService(cfg)

	case "smtp":
		if cfg.SMTPHost == "" {
			log.Println("[EMAIL] Warning: SMTP_HOST is not set. Falling back to Mock email provider.")
			return NewMockService(cfg)
		}
		log.Printf("[EMAIL] Initialized SMTP Email Service (Host: %s, Sender: %s)", cfg.SMTPHost, cfg.FromAddress)
		return NewSMTPService(cfg)

	default:
		log.Printf("[EMAIL] Initialized Mock / Development Email Service (Outgoing emails captured in memory)")
		return NewMockService(cfg)
	}
}
