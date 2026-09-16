package email

import (
	"strings"
	"testing"
)

func TestRenderMagicLinkEmail(t *testing.T) {
	sub, txt, html := RenderMagicLinkEmail("Sarah Chen", "http://localhost:3000/?magic_token=MAGIC-123", "MAGIC-123")
	if !strings.Contains(sub, "Sign in to ChoreSync") {
		t.Fatalf("unexpected subject: %s", sub)
	}
	if !strings.Contains(txt, "MAGIC-123") || !strings.Contains(txt, "http://localhost:3000/?magic_token=MAGIC-123") {
		t.Fatalf("text body missing token or link: %s", txt)
	}
	if !strings.Contains(txt, "15 minutes") || !strings.Contains(txt, "single-use") {
		t.Fatalf("text body missing expiration notice: %s", txt)
	}
	if !strings.Contains(html, "MAGIC-123") || !strings.Contains(html, "Sarah Chen") {
		t.Fatalf("html body missing token or name: %s", html)
	}
	if !strings.Contains(html, "15 minutes") || !strings.Contains(html, "single-use") {
		t.Fatalf("html body missing expiration notice: %s", html)
	}
}

func TestRenderPasswordResetEmail(t *testing.T) {
	sub, txt, html := RenderPasswordResetEmail("Liam Vance", "http://localhost:3000/?reset_token=RST-456", "RST-456")
	if !strings.Contains(sub, "Reset your ChoreSync password") {
		t.Fatalf("unexpected subject: %s", sub)
	}
	if !strings.Contains(txt, "RST-456") {
		t.Fatalf("text body missing reset token: %s", txt)
	}
	if !strings.Contains(html, "RST-456") || !strings.Contains(html, "Reset My Password") {
		t.Fatalf("html body missing token or button: %s", html)
	}
}

func TestRenderChoreReminderEmail(t *testing.T) {
	sub, txt, html := RenderChoreReminderEmail("Maya Patel", "Liam Vance", "Clean Kitchen Sink", "Tonight 8 PM", "Apartment 4B")
	if !strings.Contains(sub, "Clean Kitchen Sink") {
		t.Fatalf("unexpected subject: %s", sub)
	}
	if !strings.Contains(txt, "Clean Kitchen Sink") || !strings.Contains(txt, "Liam Vance") {
		t.Fatalf("text body missing chore or sender: %s", txt)
	}
	if !strings.Contains(html, "Apartment 4B") {
		t.Fatalf("html body missing household: %s", html)
	}
}

func TestMockServiceRecording(t *testing.T) {
	cfg := Config{FromAddress: "test@example.com"}
	svc := NewMockService(cfg)

	err := svc.SendMagicLink("sarah@example.com", "Sarah", "http://localhost:3000/?token=abc", "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = svc.SendPasswordReset("sarah@example.com", "Sarah", "http://localhost:3000/reset?token=xyz", "xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = svc.SendChoreReminder("sarah@example.com", "Sarah", "Liam", "Dishes", "today", "Flat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	emails := svc.GetRecentDevEmails()
	if len(emails) != 3 {
		t.Fatalf("expected 3 emails, got %d", len(emails))
	}

	if emails[0].Type != "chore_reminder" {
		t.Fatalf("expected newest email to be chore_reminder, got %s", emails[0].Type)
	}

	svc.ClearDevEmails()
	if len(svc.GetRecentDevEmails()) != 0 {
		t.Fatalf("expected 0 emails after clear")
	}
}
