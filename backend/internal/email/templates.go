package email

import (
	"fmt"
	"html"
	"strings"
)

// RenderMagicLinkEmail renders HTML and plain text for magic link sign-in.
func RenderMagicLinkEmail(toName, magicLink, token string) (subject, textBody, htmlBody string) {
	name := strings.TrimSpace(toName)
	if name == "" {
		name = "there"
	}
	subject = "Sign in to ChoreSync 🏠"
	textBody = fmt.Sprintf("Hi %s,\n\nUse this link to sign in to your ChoreSync account:\n%s\n\nOr enter this code directly: %s\n\nThis single-use magic link is valid for 15 minutes and will expire immediately upon login.\n\nIf you did not request this email, you can safely ignore it.\n\nBest regards,\nThe ChoreSync Team", name, magicLink, token)

	htmlBody = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; color: #1e293b; margin: 0; padding: 24px; }
    .card { max-width: 520px; margin: 0 auto; background: #ffffff; border-radius: 12px; padding: 32px; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); border: 1px solid #e2e8f0; }
    .logo { font-size: 24px; font-weight: 800; color: #4f46e5; margin-bottom: 24px; display: flex; align-items: center; gap: 8px; }
    .btn { display: inline-block; background-color: #4f46e5; color: #ffffff !important; text-decoration: none; font-weight: 600; padding: 12px 24px; border-radius: 8px; margin: 20px 0; }
    .code-box { background-color: #f1f5f9; border-radius: 8px; padding: 12px 16px; font-family: monospace; font-size: 18px; font-weight: 700; letter-spacing: 1px; color: #334155; margin: 16px 0; text-align: center; }
    .footer { margin-top: 32px; font-size: 12px; color: #94a3b8; text-align: center; border-top: 1px solid #f1f5f9; padding-top: 16px; }
  </style>
</head>
<body>
  <div class="card">
    <div class="logo">🧹 ChoreSync</div>
    <h2>Sign in to your household</h2>
    <p>Hi <strong>%s</strong>,</p>
    <p>Click the button below to instantly sign in to your ChoreSync account:</p>
    <div style="text-align: center;">
      <a href="%s" target="_top" class="btn">Sign In to ChoreSync</a>
    </div>
    <p>Or enter this verification code directly:</p>
    <div class="code-box">%s</div>
    <p style="font-size: 13px; color: #64748b;">This single-use link is valid for 15 minutes and will expire immediately upon sign-in.</p>
    <p style="font-size: 13px; color: #64748b;">Link: <a href="%s" target="_top" style="color: #4f46e5; word-break: break-all;">%s</a></p>
    <div class="footer">
      If you didn't request this link, you can safely ignore this email.<br>&copy; ChoreSync • Fair &amp; Transparent Household Coordination
    </div>
  </div>
</body>
</html>`, html.EscapeString(name), html.EscapeString(magicLink), html.EscapeString(token), html.EscapeString(magicLink), html.EscapeString(magicLink))

	return subject, textBody, htmlBody
}

// RenderPasswordResetEmail renders HTML and plain text for password reset requests.
func RenderPasswordResetEmail(toName, resetLink, token string) (subject, textBody, htmlBody string) {
	name := strings.TrimSpace(toName)
	if name == "" {
		name = "there"
	}
	subject = "Reset your ChoreSync password 🔐"
	textBody = fmt.Sprintf("Hi %s,\n\nWe received a request to reset your ChoreSync account password. Use the link below to set a new password:\n%s\n\nReset Token: %s\n\nThis link is valid for 1 hour. If you did not request a password reset, please ignore this email or check your account security.\n\nBest regards,\nThe ChoreSync Team", name, resetLink, token)

	htmlBody = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; color: #1e293b; margin: 0; padding: 24px; }
    .card { max-width: 520px; margin: 0 auto; background: #ffffff; border-radius: 12px; padding: 32px; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); border: 1px solid #e2e8f0; }
    .logo { font-size: 24px; font-weight: 800; color: #4f46e5; margin-bottom: 24px; }
    .btn { display: inline-block; background-color: #dc2626; color: #ffffff !important; text-decoration: none; font-weight: 600; padding: 12px 24px; border-radius: 8px; margin: 20px 0; }
    .token-box { background-color: #fef2f2; border: 1px solid #fecaca; border-radius: 8px; padding: 12px; font-family: monospace; font-size: 16px; font-weight: bold; color: #991b1b; margin: 16px 0; text-align: center; }
    .footer { margin-top: 32px; font-size: 12px; color: #94a3b8; text-align: center; border-top: 1px solid #f1f5f9; padding-top: 16px; }
  </style>
</head>
<body>
  <div class="card">
    <div class="logo">🧹 ChoreSync</div>
    <h2>Password Reset Request</h2>
    <p>Hi <strong>%s</strong>,</p>
    <p>We received a request to reset the password for your ChoreSync account. Click the button below to choose a new password:</p>
    <div style="text-align: center;">
      <a href="%s" target="_top" class="btn">Reset My Password</a>
    </div>
    <p>Reset verification token:</p>
    <div class="token-box">%s</div>
    <p style="font-size: 13px; color: #64748b;">Direct link: <a href="%s" target="_top" style="color: #4f46e5; word-break: break-all;">%s</a></p>
    <p style="font-size: 13px; color: #64748b;">This reset link will expire in 1 hour.</p>
    <div class="footer">
      If you did not make this request, you can safely ignore this email.<br>&copy; ChoreSync • Fair &amp; Transparent Household Coordination
    </div>
  </div>
</body>
</html>`, html.EscapeString(name), html.EscapeString(resetLink), html.EscapeString(token), html.EscapeString(resetLink), html.EscapeString(resetLink))

	return subject, textBody, htmlBody
}

// RenderChoreReminderEmail renders HTML and plain text for friendly chore reminder nudges.
func RenderChoreReminderEmail(toName, senderName, choreTitle, dueDate, householdName string) (subject, textBody, htmlBody string) {
	name := strings.TrimSpace(toName)
	if name == "" {
		name = "there"
	}
	sender := strings.TrimSpace(senderName)
	if sender == "" {
		sender = "A roommate"
	}
	subject = fmt.Sprintf("Friendly reminder: %s is due soon! ⏰", choreTitle)
	dueStr := dueDate
	if dueStr == "" {
		dueStr = "today"
	}

	textBody = fmt.Sprintf("Hi %s,\n\n%s sent you a friendly reminder from %s:\nChore: \"%s\"\nDue: %s\n\nLog in to ChoreSync to check it off or request a swap if needed!\n\nBest regards,\nThe ChoreSync Team", name, sender, householdName, choreTitle, dueStr)

	htmlBody = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; color: #1e293b; margin: 0; padding: 24px; }
    .card { max-width: 520px; margin: 0 auto; background: #ffffff; border-radius: 12px; padding: 32px; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); border: 1px solid #e2e8f0; }
    .logo { font-size: 24px; font-weight: 800; color: #4f46e5; margin-bottom: 24px; }
    .chore-box { background-color: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 8px; padding: 16px; margin: 20px 0; }
    .chore-title { font-size: 18px; font-weight: 700; color: #166534; margin: 0 0 8px 0; }
    .chore-due { font-size: 14px; color: #15803d; margin: 0; }
    .btn { display: inline-block; background-color: #16a34a; color: #ffffff !important; text-decoration: none; font-weight: 600; padding: 12px 24px; border-radius: 8px; margin: 16px 0; }
    .footer { margin-top: 32px; font-size: 12px; color: #94a3b8; text-align: center; border-top: 1px solid #f1f5f9; padding-top: 16px; }
  </style>
</head>
<body>
  <div class="card">
    <div class="logo">🧹 ChoreSync</div>
    <h2>Friendly Chore Reminder</h2>
    <p>Hi <strong>%s</strong>,</p>
    <p><strong>%s</strong> sent you a friendly nudge from <em>%s</em>:</p>
    <div class="chore-box">
      <div class="chore-title">%s</div>
      <div class="chore-due">📅 Due: %s</div>
    </div>
    <div style="text-align: center;">
      <a href="http://localhost:3000" class="btn">View &amp; Complete Chore</a>
    </div>
    <p style="font-size: 13px; color: #64748b;">Need someone else to take it? You can propose a peer swap directly in the app.</p>
    <div class="footer">
      Keep your household in sync • ChoreSync
    </div>
  </div>
</body>
</html>`, html.EscapeString(name), html.EscapeString(sender), html.EscapeString(householdName), html.EscapeString(choreTitle), html.EscapeString(dueStr))

	return subject, textBody, htmlBody
}
