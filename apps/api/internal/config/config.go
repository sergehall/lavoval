package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv                          string
	AppName                         string
	AppURL                          string
	HTTPAddr                        string
	DatabaseURL                     string
	JWTIssuer                       string
	JWTAudience                     string
	JWTSecret                       string
	JWTAccessTTL                    time.Duration
	JWTRefreshTTL                   time.Duration
	CookieSecure                    bool
	AdminSeedEmail                  string
	AdminSeedSecret                 string
	EmailVerificationTTL            time.Duration
	PasswordResetTTL                time.Duration
	MFATOTPPeriod                   time.Duration
	MFATOTPIssuer                   string
	MFASecretKey                    string
	MFASignInChallengeTTL           time.Duration
	OAuthStateTTL                   time.Duration
	GoogleOAuthClientID             string
	GoogleOAuthSecret               string
	GitHubOAuthClientID             string
	GitHubOAuthSecret               string
	SMTPHost                        string
	SMTPPort                        int
	SMTPUsername                    string
	SMTPPassword                    string
	SMTPFromEmail                   string
	SMTPFromName                    string
	SMTPRequireTLS                  bool
	SMTPAllowInsecureAuth           bool
	SMTPUseSSL                      bool
	SMTPDialTimeout                 time.Duration
	MailProvider                    string
	MailSendTimeout                 time.Duration
	MailWorkerCount                 int
	MailMaxAttempts                 int
	MailRetryBaseDelay              time.Duration
	MailPollInterval                time.Duration
	MailLeaseTTL                    time.Duration
	MailRateLimitPerSecond          int
	MailRateLimitBurst              int
	MailJobsRetention               time.Duration
	MailEventsRetention             time.Duration
	MailCleanupBatchSize            int
	MailCleanupInterval             time.Duration
	MailCleanupDryRun               bool
	MailCleanupAlertJobsThreshold   int64
	MailCleanupAlertEventsThreshold int64
	MailCleanupAlertFailureStreak   int
	MailCleanupAlertStaleAfter      time.Duration
	MailAlertWebhookURL             string
	MailAlertWebhookTimeout         time.Duration
	MailAlertWebhookCooldown        time.Duration
	GmailAPIBaseURL                 string
	GmailAPIUser                    string
	GmailAPIAccessToken             string
	GmailAPIRefreshToken            string
	GmailAPIClientID                string
	GmailAPIClientSecret            string
	GmailAPITokenURL                string
}

func Load() (Config, error) {
	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return Config{}, fmt.Errorf("parse JWT_ACCESS_TTL: %w", err)
	}

	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "720h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse JWT_REFRESH_TTL: %w", err)
	}

	cookieSecure, err := strconv.ParseBool(getEnv("COOKIE_SECURE", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("parse COOKIE_SECURE: %w", err)
	}

	emailVerificationTTL, err := time.ParseDuration(getEnv("EMAIL_VERIFICATION_TTL", "24h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse EMAIL_VERIFICATION_TTL: %w", err)
	}

	passwordResetTTL, err := time.ParseDuration(getEnv("PASSWORD_RESET_TTL", "30m"))
	if err != nil {
		return Config{}, fmt.Errorf("parse PASSWORD_RESET_TTL: %w", err)
	}

	mfaTOTPPeriod, err := time.ParseDuration(getEnv("MFA_TOTP_PERIOD", "30s"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MFA_TOTP_PERIOD: %w", err)
	}

	mfaSignInChallengeTTL, err := time.ParseDuration(getEnv("MFA_SIGN_IN_CHALLENGE_TTL", "10m"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MFA_SIGN_IN_CHALLENGE_TTL: %w", err)
	}

	oauthStateTTL, err := time.ParseDuration(getEnv("OAUTH_STATE_TTL", "10m"))
	if err != nil {
		return Config{}, fmt.Errorf("parse OAUTH_STATE_TTL: %w", err)
	}

	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		return Config{}, fmt.Errorf("parse SMTP_PORT: %w", err)
	}

	smtpRequireTLS, err := strconv.ParseBool(getEnv("SMTP_REQUIRE_TLS", "true"))
	if err != nil {
		return Config{}, fmt.Errorf("parse SMTP_REQUIRE_TLS: %w", err)
	}

	smtpAllowInsecureAuth, err := strconv.ParseBool(getEnv("SMTP_ALLOW_INSECURE_AUTH", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("parse SMTP_ALLOW_INSECURE_AUTH: %w", err)
	}

	smtpUseSSL, err := strconv.ParseBool(getEnv("SMTP_USE_SSL", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("parse SMTP_USE_SSL: %w", err)
	}

	smtpDialTimeout, err := time.ParseDuration(getEnv("SMTP_DIAL_TIMEOUT", "10s"))
	if err != nil {
		return Config{}, fmt.Errorf("parse SMTP_DIAL_TIMEOUT: %w", err)
	}

	mailSendTimeout, err := time.ParseDuration(getEnv("MAIL_SEND_TIMEOUT", "15s"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_SEND_TIMEOUT: %w", err)
	}

	mailWorkerCount, err := strconv.Atoi(getEnv("MAIL_WORKER_COUNT", "4"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_WORKER_COUNT: %w", err)
	}

	mailMaxAttempts, err := strconv.Atoi(getEnv("MAIL_MAX_ATTEMPTS", "4"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_MAX_ATTEMPTS: %w", err)
	}

	mailRetryBaseDelay, err := time.ParseDuration(getEnv("MAIL_RETRY_BASE_DELAY", "1s"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_RETRY_BASE_DELAY: %w", err)
	}

	mailPollInterval, err := time.ParseDuration(getEnv("MAIL_POLL_INTERVAL", "500ms"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_POLL_INTERVAL: %w", err)
	}

	mailLeaseTTL, err := time.ParseDuration(getEnv("MAIL_LEASE_TTL", "30s"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_LEASE_TTL: %w", err)
	}

	mailRateLimitPerSecond, err := strconv.Atoi(getEnv("MAIL_RATE_LIMIT_PER_SECOND", "0"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_RATE_LIMIT_PER_SECOND: %w", err)
	}

	mailRateLimitBurst, err := strconv.Atoi(getEnv("MAIL_RATE_LIMIT_BURST", "1"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_RATE_LIMIT_BURST: %w", err)
	}

	mailJobsRetention, err := time.ParseDuration(getEnv("MAIL_JOBS_RETENTION", "720h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_JOBS_RETENTION: %w", err)
	}

	mailEventsRetention, err := time.ParseDuration(getEnv("MAIL_EVENTS_RETENTION", "336h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_EVENTS_RETENTION: %w", err)
	}

	mailCleanupBatchSize, err := strconv.Atoi(getEnv("MAIL_CLEANUP_BATCH_SIZE", "500"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_CLEANUP_BATCH_SIZE: %w", err)
	}

	mailCleanupInterval, err := time.ParseDuration(getEnv("MAIL_CLEANUP_INTERVAL", "1h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_CLEANUP_INTERVAL: %w", err)
	}

	mailCleanupDryRun, err := strconv.ParseBool(getEnv("MAIL_CLEANUP_DRY_RUN", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_CLEANUP_DRY_RUN: %w", err)
	}

	mailCleanupAlertJobsThreshold, err := strconv.ParseInt(getEnv("MAIL_CLEANUP_ALERT_JOBS_THRESHOLD", "5000"), 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_CLEANUP_ALERT_JOBS_THRESHOLD: %w", err)
	}

	mailCleanupAlertEventsThreshold, err := strconv.ParseInt(getEnv("MAIL_CLEANUP_ALERT_EVENTS_THRESHOLD", "20000"), 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_CLEANUP_ALERT_EVENTS_THRESHOLD: %w", err)
	}

	mailCleanupAlertFailureStreak, err := strconv.Atoi(getEnv("MAIL_CLEANUP_ALERT_FAILURE_STREAK", "3"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_CLEANUP_ALERT_FAILURE_STREAK: %w", err)
	}

	mailCleanupAlertStaleAfter, err := time.ParseDuration(getEnv("MAIL_CLEANUP_ALERT_STALE_AFTER", "6h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_CLEANUP_ALERT_STALE_AFTER: %w", err)
	}

	mailAlertWebhookTimeout, err := time.ParseDuration(getEnv("MAIL_ALERT_WEBHOOK_TIMEOUT", "5s"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_ALERT_WEBHOOK_TIMEOUT: %w", err)
	}

	mailAlertWebhookCooldown, err := time.ParseDuration(getEnv("MAIL_ALERT_WEBHOOK_COOLDOWN", "30m"))
	if err != nil {
		return Config{}, fmt.Errorf("parse MAIL_ALERT_WEBHOOK_COOLDOWN: %w", err)
	}

	cfg := Config{
		AppEnv:                          getEnv("APP_ENV", "development"),
		AppName:                         getEnv("APP_NAME", "Lavoval"),
		AppURL:                          getEnv("APP_URL", "http://localhost:3000"),
		HTTPAddr:                        getEnv("BACKEND_HTTP_ADDR", ":8080"),
		DatabaseURL:                     getEnv("DATABASE_URL", "postgres://codex:codex@localhost:5432/lavoval?sslmode=disable"),
		JWTIssuer:                       getEnv("JWT_ISSUER", "lavoval"),
		JWTAudience:                     getEnv("JWT_AUDIENCE", "lavoval-web"),
		JWTSecret:                       getEnv("JWT_SECRET", "change-me"),
		JWTAccessTTL:                    accessTTL,
		JWTRefreshTTL:                   refreshTTL,
		CookieSecure:                    cookieSecure,
		AdminSeedEmail:                  getEnv("ADMIN_SEED_EMAIL", "admin@lavoval.local"),
		AdminSeedSecret:                 getEnv("ADMIN_SEED_PASSWORD", "ChangeMe123!"),
		EmailVerificationTTL:            emailVerificationTTL,
		PasswordResetTTL:                passwordResetTTL,
		MFATOTPPeriod:                   mfaTOTPPeriod,
		MFATOTPIssuer:                   getEnv("MFA_TOTP_ISSUER", getEnv("APP_NAME", "Lavoval")),
		MFASecretKey:                    getEnv("MFA_SECRET_KEY", ""),
		MFASignInChallengeTTL:           mfaSignInChallengeTTL,
		OAuthStateTTL:                   oauthStateTTL,
		GoogleOAuthClientID:             getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
		GoogleOAuthSecret:               getEnv("GOOGLE_OAUTH_CLIENT_SECRET", ""),
		GitHubOAuthClientID:             getEnv("GITHUB_OAUTH_CLIENT_ID", ""),
		GitHubOAuthSecret:               getEnv("GITHUB_OAUTH_CLIENT_SECRET", ""),
		SMTPHost:                        getEnv("SMTP_HOST", ""),
		SMTPPort:                        smtpPort,
		SMTPUsername:                    getEnv("SMTP_USERNAME", ""),
		SMTPPassword:                    getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:                   getEnv("SMTP_FROM_EMAIL", ""),
		SMTPFromName:                    getEnv("SMTP_FROM_NAME", "Lavoval"),
		SMTPRequireTLS:                  smtpRequireTLS,
		SMTPAllowInsecureAuth:           smtpAllowInsecureAuth,
		SMTPUseSSL:                      smtpUseSSL,
		SMTPDialTimeout:                 smtpDialTimeout,
		MailProvider:                    getEnv("MAIL_PROVIDER", "smtp"),
		MailSendTimeout:                 mailSendTimeout,
		MailWorkerCount:                 mailWorkerCount,
		MailMaxAttempts:                 mailMaxAttempts,
		MailRetryBaseDelay:              mailRetryBaseDelay,
		MailPollInterval:                mailPollInterval,
		MailLeaseTTL:                    mailLeaseTTL,
		MailRateLimitPerSecond:          mailRateLimitPerSecond,
		MailRateLimitBurst:              mailRateLimitBurst,
		MailJobsRetention:               mailJobsRetention,
		MailEventsRetention:             mailEventsRetention,
		MailCleanupBatchSize:            mailCleanupBatchSize,
		MailCleanupInterval:             mailCleanupInterval,
		MailCleanupDryRun:               mailCleanupDryRun,
		MailCleanupAlertJobsThreshold:   mailCleanupAlertJobsThreshold,
		MailCleanupAlertEventsThreshold: mailCleanupAlertEventsThreshold,
		MailCleanupAlertFailureStreak:   mailCleanupAlertFailureStreak,
		MailCleanupAlertStaleAfter:      mailCleanupAlertStaleAfter,
		MailAlertWebhookURL:             getEnv("MAIL_ALERT_WEBHOOK_URL", ""),
		MailAlertWebhookTimeout:         mailAlertWebhookTimeout,
		MailAlertWebhookCooldown:        mailAlertWebhookCooldown,
		GmailAPIBaseURL:                 getEnv("GMAIL_API_BASE_URL", "https://gmail.googleapis.com/gmail/v1"),
		GmailAPIUser:                    getEnv("GMAIL_API_USER", "me"),
		GmailAPIAccessToken:             getEnv("GMAIL_API_ACCESS_TOKEN", ""),
		GmailAPIRefreshToken:            getEnv("GMAIL_API_REFRESH_TOKEN", ""),
		GmailAPIClientID:                getEnv("GMAIL_API_CLIENT_ID", ""),
		GmailAPIClientSecret:            getEnv("GMAIL_API_CLIENT_SECRET", ""),
		GmailAPITokenURL:                getEnv("GMAIL_API_TOKEN_URL", "https://oauth2.googleapis.com/token"),
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return fallback
	}
	return value
}
