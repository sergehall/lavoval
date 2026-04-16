package mailer

import (
	"strings"

	"github.com/sergehall/lavoval/apps/api/internal/config"
)

func newConfiguredProvider(cfg config.Config) Provider {
	switch normalizeMailProvider(cfg.MailProvider) {
	case "gmail_api":
		return NewGmailAPIProvider(cfg)
	case "noop":
		return NewNoopProvider(cfg)
	default:
		return NewSMTPProvider(cfg)
	}
}

func normalizeMailProvider(provider string) string {
	value := strings.ToLower(strings.TrimSpace(provider))
	switch value {
	case "", "smtp":
		return "smtp"
	case "gmail-api", "gmail_api", "gmailapi":
		return "gmail_api"
	case "noop", "disabled":
		return "noop"
	default:
		return value
	}
}
