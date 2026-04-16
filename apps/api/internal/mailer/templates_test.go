package mailer

import (
	"net/textproto"
	"strings"
	"testing"
)

func TestRenderVerificationEmailUsesLavovalBranding(t *testing.T) {
	rendered, err := renderVerificationEmail(VerificationEmail{
		ToEmail:     "serge@example.com",
		ToName:      "Serge Hall",
		VerifyURL:   "https://lavoval.test/verify-email?token=abc123",
		ProductName: "Lavoval",
	}, "https://lavoval.test")
	if err != nil {
		t.Fatalf("renderVerificationEmail returned error: %v", err)
	}

	if rendered.Subject != "Lavoval: confirm your email" {
		t.Fatalf("unexpected subject: %s", rendered.Subject)
	}

	for _, snippet := range []string{
		"Confirm your email to activate Lavoval",
		"https://lavoval.test/verify-email?token=abc123",
		"https://lavoval.test/email-brand-340-180.png",
		`width="216"`,
		`height="72"`,
		"#b44f23",
		"Human skill exchange for the AI era",
	} {
		if !strings.Contains(rendered.HTMLBody, snippet) {
			t.Fatalf("expected HTML body to contain %q", snippet)
		}
	}

	for _, snippet := range []string{
		"Email confirmation | Confirm your email to activate Lavoval",
		"Confirm email: https://lavoval.test/verify-email?token=abc123",
		"Lavoval: https://lavoval.test",
	} {
		if !strings.Contains(rendered.TextBody, snippet) {
			t.Fatalf("expected text body to contain %q", snippet)
		}
	}
}

func TestRenderVerificationEmailUsesPublicBrandFallbackForLocalhost(t *testing.T) {
	rendered, err := renderVerificationEmail(VerificationEmail{
		ToEmail:     "serge@example.com",
		ToName:      "Serge Hall",
		VerifyURL:   "http://localhost:3000/verify-email?token=abc123",
		ProductName: "Lavoval",
	}, "http://localhost:3000")
	if err != nil {
		t.Fatalf("renderVerificationEmail returned error: %v", err)
	}

	if !strings.Contains(rendered.HTMLBody, "https://lavoval.com/email-brand-340-180.png") {
		t.Fatalf("expected localhost emails to use public brand fallback, got %s", rendered.HTMLBody)
	}
}

func TestBuildMultipartMessageIncludesTextAndHTMLParts(t *testing.T) {
	headers := textproto.MIMEHeader{}
	headers.Set("X-Test-ID", "job-123")

	message, err := buildMultipartMessage("Lavoval", "noreply@lavoval.test", "serge@example.com", RenderedEmail{
		Subject:  "Lavoval: confirm your email",
		TextBody: "plain text body",
		HTMLBody: "<strong>html body</strong>",
	}, headers)
	if err != nil {
		t.Fatalf("buildMultipartMessage returned error: %v", err)
	}

	content := string(message)
	for _, snippet := range []string{
		"Content-Type: multipart/alternative;",
		`Content-Type: text/plain; charset="UTF-8"`,
		`Content-Type: text/html; charset="UTF-8"`,
		"X-Test-Id: job-123",
		"plain text body",
		"<strong>html body</strong>",
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected MIME message to contain %q", snippet)
		}
	}
}

func TestRenderPasswordResetEmailUsesLavovalBranding(t *testing.T) {
	rendered, err := renderPasswordResetEmail(PasswordResetEmail{
		ToEmail:     "serge@example.com",
		ToName:      "Serge Hall",
		ResetURL:    "https://lavoval.test/reset-password?token=abc123",
		ProductName: "Lavoval",
	}, "https://lavoval.test")
	if err != nil {
		t.Fatalf("renderPasswordResetEmail returned error: %v", err)
	}

	for _, snippet := range []string{
		"Reset your Lavoval password",
		"https://lavoval.test/reset-password?token=abc123",
		"https://lavoval.test/email-brand-340-180.png",
		"Password recovery",
	} {
		if !strings.Contains(rendered.HTMLBody, snippet) {
			t.Fatalf("expected HTML body to contain %q", snippet)
		}
	}
}

func TestRenderPasswordChangedEmailUsesSecurityNoticeCopy(t *testing.T) {
	rendered, err := renderPasswordChangedEmail(PasswordChangedEmail{
		ToEmail:     "serge@example.com",
		ToName:      "Serge Hall",
		SignInURL:   "https://lavoval.test/?auth=sign-in",
		ProductName: "Lavoval",
	}, "https://lavoval.test")
	if err != nil {
		t.Fatalf("renderPasswordChangedEmail returned error: %v", err)
	}

	for _, snippet := range []string{
		"Security notice",
		"Your Lavoval password was changed",
		"https://lavoval.test/?auth=sign-in",
	} {
		if !strings.Contains(rendered.HTMLBody, snippet) {
			t.Fatalf("expected HTML body to contain %q", snippet)
		}
	}
}
