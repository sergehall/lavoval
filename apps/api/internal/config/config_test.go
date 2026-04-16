package config

import (
	"testing"
	"time"
)

func TestValidateProductionConfigAcceptsHealthyConfiguration(t *testing.T) {
	cfg := validProductionConfig()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected config to validate, got %v", err)
	}
}

func TestValidateProductionConfigRejectsUnsafeSettings(t *testing.T) {
	cfg := validProductionConfig()
	cfg.AppURL = "http://lavoval.com"
	cfg.DatabaseURL = "postgres://codex:codex@localhost:5432/lavoval?sslmode=disable"
	cfg.JWTSecret = "change-me"
	cfg.CookieSecure = false

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected invalid production config to fail validation")
	}
}

func TestHealthFlagsMarksHalfConfiguredOAuthAsInvalid(t *testing.T) {
	cfg := validProductionConfig()
	cfg.GoogleOAuthClientID = "google-client-id"
	cfg.GoogleOAuthSecret = ""

	flags := cfg.HealthFlags()
	if flags.GoogleOAuthValid {
		t.Fatal("expected half-configured Google OAuth to be invalid")
	}
	if flags.GoogleOAuthConfigured {
		t.Fatal("expected half-configured Google OAuth to be reported as not configured")
	}
}

func TestHealthFlagsReflectSelectedMailProvider(t *testing.T) {
	cfg := validProductionConfig()
	cfg.MailProvider = "gmail_api"
	cfg.GmailAPIAccessToken = ""
	cfg.GmailAPIRefreshToken = "refresh-token"
	cfg.GmailAPIClientID = "gmail-client-id"
	cfg.GmailAPIClientSecret = "gmail-client-secret"

	flags := cfg.HealthFlags()
	if !flags.MailConfigured || !flags.MailValid {
		t.Fatalf("expected Gmail API provider to be valid, got %+v", flags)
	}
	if !flags.GmailAPIRefreshConfigured {
		t.Fatal("expected refresh-based Gmail API flow to be reported as configured")
	}
}

func validProductionConfig() Config {
	return Config{
		AppEnv:               "production",
		AppName:              "Lavoval",
		AppURL:               "https://lavoval.com",
		HTTPAddr:             ":10000",
		DatabaseURL:          "postgres://user:strong-password@db.example.com:5432/lavoval?sslmode=require",
		JWTIssuer:            "lavoval",
		JWTAudience:          "lavoval-web",
		JWTSecret:            "0123456789abcdef0123456789abcdef",
		JWTAccessTTL:         15 * time.Minute,
		JWTRefreshTTL:        720 * time.Hour,
		CookieSecure:         true,
		EmailVerificationTTL: 24 * time.Hour,
		PasswordResetTTL:     time.Hour,
		MailProvider:         "smtp",
		SMTPHost:             "smtp.example.com",
		SMTPPort:             587,
		SMTPUsername:         "mailer@example.com",
		SMTPPassword:         "smtp-password",
		SMTPFromEmail:        "mailer@example.com",
		MailSendTimeout:      time.Second,
		MailLeaseTTL:         time.Second,
		MailPollInterval:     time.Second,
		MailWorkerCount:      2,
		MailMaxAttempts:      4,
		MailRateLimitBurst:   1,
		MailCleanupBatchSize: 100,
	}
}
