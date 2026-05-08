package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
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
	LoginIPRateLimitEnabled         bool
	LoginIPRateLimitWindow          time.Duration
	LoginIPRateLimitMaxAttempts     int
	LoginIPRateLimitSlowAfter       int
	LoginIPRateLimitBaseDelay       time.Duration
	LoginIPRateLimitMaxDelay        time.Duration
	LoginIPRateLimitBlockDuration   time.Duration
	LoginIPRateLimitStateTTL        time.Duration
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

type HealthFlags struct {
	AppEnv                    string `json:"app_env"`
	AppURLConfigured          bool   `json:"app_url_configured"`
	AppURLValid               bool   `json:"app_url_valid"`
	AppURLHTTPS               bool   `json:"app_url_https"`
	DatabaseConfigured        bool   `json:"database_configured"`
	DatabaseProductionSafe    bool   `json:"database_production_safe"`
	JWTConfigured             bool   `json:"jwt_configured"`
	JWTStrong                 bool   `json:"jwt_strong"`
	CookieSecure              bool   `json:"cookie_secure"`
	MFAConfigured             bool   `json:"mfa_configured"`
	GoogleOAuthConfigured     bool   `json:"google_oauth_configured"`
	GoogleOAuthValid          bool   `json:"google_oauth_valid"`
	GitHubOAuthConfigured     bool   `json:"github_oauth_configured"`
	GitHubOAuthValid          bool   `json:"github_oauth_valid"`
	MailProvider              string `json:"mail_provider"`
	MailConfigured            bool   `json:"mail_configured"`
	MailValid                 bool   `json:"mail_valid"`
	MailAlertingEnabled       bool   `json:"mail_alerting_enabled"`
	MailCleanupAutoEnabled    bool   `json:"mail_cleanup_auto_enabled"`
	SMTPConfigured            bool   `json:"smtp_configured"`
	GmailAPIEnabled           bool   `json:"gmail_api_enabled"`
	GmailAPIRefreshConfigured bool   `json:"gmail_api_refresh_configured"`
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

	authLoginIPRateLimitEnabled, err := strconv.ParseBool(getEnv("AUTH_LOGIN_IP_RATE_LIMIT_ENABLED", "true"))
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_LOGIN_IP_RATE_LIMIT_ENABLED: %w", err)
	}

	authLoginIPRateLimitWindow, err := time.ParseDuration(getEnv("AUTH_LOGIN_IP_RATE_LIMIT_WINDOW", "10m"))
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_LOGIN_IP_RATE_LIMIT_WINDOW: %w", err)
	}

	authLoginIPRateLimitMaxAttempts, err := strconv.Atoi(getEnv("AUTH_LOGIN_IP_RATE_LIMIT_MAX_ATTEMPTS", "8"))
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_LOGIN_IP_RATE_LIMIT_MAX_ATTEMPTS: %w", err)
	}

	authLoginIPRateLimitSlowAfter, err := strconv.Atoi(getEnv("AUTH_LOGIN_IP_RATE_LIMIT_SLOW_AFTER", "3"))
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_LOGIN_IP_RATE_LIMIT_SLOW_AFTER: %w", err)
	}

	authLoginIPRateLimitBaseDelay, err := time.ParseDuration(getEnv("AUTH_LOGIN_IP_RATE_LIMIT_BASE_DELAY", "500ms"))
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_LOGIN_IP_RATE_LIMIT_BASE_DELAY: %w", err)
	}

	authLoginIPRateLimitMaxDelay, err := time.ParseDuration(getEnv("AUTH_LOGIN_IP_RATE_LIMIT_MAX_DELAY", "4s"))
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_LOGIN_IP_RATE_LIMIT_MAX_DELAY: %w", err)
	}

	authLoginIPRateLimitBlockDuration, err := time.ParseDuration(getEnv("AUTH_LOGIN_IP_RATE_LIMIT_BLOCK_DURATION", "15m"))
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_LOGIN_IP_RATE_LIMIT_BLOCK_DURATION: %w", err)
	}

	authLoginIPRateLimitStateTTL, err := time.ParseDuration(getEnv("AUTH_LOGIN_IP_RATE_LIMIT_STATE_TTL", "30m"))
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_LOGIN_IP_RATE_LIMIT_STATE_TTL: %w", err)
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
		HTTPAddr:                        resolveHTTPAddr(),
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
		LoginIPRateLimitEnabled:         authLoginIPRateLimitEnabled,
		LoginIPRateLimitWindow:          authLoginIPRateLimitWindow,
		LoginIPRateLimitMaxAttempts:     authLoginIPRateLimitMaxAttempts,
		LoginIPRateLimitSlowAfter:       authLoginIPRateLimitSlowAfter,
		LoginIPRateLimitBaseDelay:       authLoginIPRateLimitBaseDelay,
		LoginIPRateLimitMaxDelay:        authLoginIPRateLimitMaxDelay,
		LoginIPRateLimitBlockDuration:   authLoginIPRateLimitBlockDuration,
		LoginIPRateLimitStateTTL:        authLoginIPRateLimitStateTTL,
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

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// resolveHTTPAddr returns the address the HTTP server should bind to.
// Prefers BACKEND_HTTP_ADDR; falls back to :PORT (Render injects PORT);
// defaults to :8080 for local development.
func resolveHTTPAddr() string {
	if addr := os.Getenv("BACKEND_HTTP_ADDR"); addr != "" {
		return addr
	}
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":8080"
}

func getEnv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return fallback
	}
	return value
}

func (c Config) Validate() error {
	var problems []string

	flags := c.HealthFlags()

	if !flags.AppURLConfigured || !flags.AppURLValid {
		problems = append(problems, "APP_URL must be a valid absolute http(s) URL")
	}
	if !flags.DatabaseConfigured {
		problems = append(problems, "DATABASE_URL must be configured")
	}
	if !flags.JWTConfigured {
		problems = append(problems, "JWT_ISSUER, JWT_AUDIENCE, and JWT_SECRET must be configured")
	}
	if !flags.JWTStrong {
		problems = append(problems, "JWT_SECRET must not use a placeholder value and should be at least 32 characters")
	}
	if c.JWTAccessTTL <= 0 || c.JWTRefreshTTL <= 0 {
		problems = append(problems, "JWT_ACCESS_TTL and JWT_REFRESH_TTL must be greater than zero")
	}
	if c.EmailVerificationTTL <= 0 || c.PasswordResetTTL <= 0 {
		problems = append(problems, "EMAIL_VERIFICATION_TTL and PASSWORD_RESET_TTL must be greater than zero")
	}
	if c.LoginIPRateLimitEnabled {
		if c.LoginIPRateLimitWindow <= 0 || c.LoginIPRateLimitBaseDelay <= 0 || c.LoginIPRateLimitMaxDelay <= 0 || c.LoginIPRateLimitBlockDuration <= 0 || c.LoginIPRateLimitStateTTL <= 0 {
			problems = append(problems, "AUTH_LOGIN_IP_RATE_LIMIT_WINDOW, AUTH_LOGIN_IP_RATE_LIMIT_BASE_DELAY, AUTH_LOGIN_IP_RATE_LIMIT_MAX_DELAY, AUTH_LOGIN_IP_RATE_LIMIT_BLOCK_DURATION, and AUTH_LOGIN_IP_RATE_LIMIT_STATE_TTL must be greater than zero")
		}
		if c.LoginIPRateLimitMaxAttempts <= 0 {
			problems = append(problems, "AUTH_LOGIN_IP_RATE_LIMIT_MAX_ATTEMPTS must be greater than zero")
		}
		if c.LoginIPRateLimitSlowAfter < 0 {
			problems = append(problems, "AUTH_LOGIN_IP_RATE_LIMIT_SLOW_AFTER must not be negative")
		}
		if c.LoginIPRateLimitSlowAfter > c.LoginIPRateLimitMaxAttempts {
			problems = append(problems, "AUTH_LOGIN_IP_RATE_LIMIT_SLOW_AFTER must be less than or equal to AUTH_LOGIN_IP_RATE_LIMIT_MAX_ATTEMPTS")
		}
	}
	if c.MailSendTimeout <= 0 || c.MailLeaseTTL <= 0 || c.MailPollInterval <= 0 {
		problems = append(problems, "MAIL_SEND_TIMEOUT, MAIL_LEASE_TTL, and MAIL_POLL_INTERVAL must be greater than zero")
	}
	if c.MailWorkerCount <= 0 || c.MailMaxAttempts <= 0 || c.MailRateLimitBurst <= 0 {
		problems = append(problems, "MAIL_WORKER_COUNT, MAIL_MAX_ATTEMPTS, and MAIL_RATE_LIMIT_BURST must be greater than zero")
	}
	if c.MailCleanupBatchSize <= 0 {
		problems = append(problems, "MAIL_CLEANUP_BATCH_SIZE must be greater than zero")
	}
	if !flags.MailValid {
		problems = append(problems, "mail provider configuration is incomplete for MAIL_PROVIDER="+normalizedMailProvider(c.MailProvider))
	}
	if !flags.GoogleOAuthValid {
		problems = append(problems, "GOOGLE_OAUTH_CLIENT_ID and GOOGLE_OAUTH_CLIENT_SECRET must either both be set or both be empty")
	}
	if !flags.GitHubOAuthValid {
		problems = append(problems, "GITHUB_OAUTH_CLIENT_ID and GITHUB_OAUTH_CLIENT_SECRET must either both be set or both be empty")
	}

	if c.isProduction() {
		if !flags.AppURLHTTPS {
			problems = append(problems, "APP_URL must use https in production")
		}
		if !flags.DatabaseProductionSafe {
			problems = append(problems, "DATABASE_URL must not use a localhost/default development DSN in production")
		}
		if !c.CookieSecure {
			problems = append(problems, "COOKIE_SECURE must be true in production")
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid runtime configuration: %s", strings.Join(problems, "; "))
	}

	return nil
}

func (c Config) HealthFlags() HealthFlags {
	appURLValid, appURLHTTPS := validateAppURL(c.AppURL)
	mailProvider := normalizedMailProvider(c.MailProvider)
	smtpConfigured := isPresent(c.SMTPHost) && c.SMTPPort > 0 && isPresent(c.SMTPUsername) && isPresent(c.SMTPPassword) && isPresent(c.SMTPFromEmail)
	gmailAPIDirectConfigured := isPresent(c.GmailAPIAccessToken)
	gmailAPIRefreshConfigured := isPresent(c.GmailAPIRefreshToken) && isPresent(c.GmailAPIClientID) && isPresent(c.GmailAPIClientSecret)
	googleConfigured, googleValid := oauthPairState(c.GoogleOAuthClientID, c.GoogleOAuthSecret)
	githubConfigured, githubValid := oauthPairState(c.GitHubOAuthClientID, c.GitHubOAuthSecret)

	mailConfigured := false
	mailValid := true
	switch mailProvider {
	case "smtp":
		mailConfigured = smtpConfigured
		mailValid = smtpConfigured
	case "gmail_api":
		mailConfigured = isPresent(c.SMTPFromEmail) && (gmailAPIDirectConfigured || gmailAPIRefreshConfigured)
		mailValid = mailConfigured
	case "noop":
		mailConfigured = true
		mailValid = true
	default:
		mailValid = false
	}

	return HealthFlags{
		AppEnv:                    c.AppEnv,
		AppURLConfigured:          isPresent(c.AppURL),
		AppURLValid:               appURLValid,
		AppURLHTTPS:               appURLHTTPS,
		DatabaseConfigured:        isPresent(c.DatabaseURL),
		DatabaseProductionSafe:    !looksLikeDefaultDatabaseURL(c.DatabaseURL),
		JWTConfigured:             isPresent(c.JWTIssuer) && isPresent(c.JWTAudience) && isPresent(c.JWTSecret),
		JWTStrong:                 isStrongSecret(c.JWTSecret),
		CookieSecure:              c.CookieSecure,
		MFAConfigured:             isPresent(c.MFASecretKey),
		GoogleOAuthConfigured:     googleConfigured,
		GoogleOAuthValid:          googleValid,
		GitHubOAuthConfigured:     githubConfigured,
		GitHubOAuthValid:          githubValid,
		MailProvider:              mailProvider,
		MailConfigured:            mailConfigured,
		MailValid:                 mailValid,
		MailAlertingEnabled:       isPresent(c.MailAlertWebhookURL),
		MailCleanupAutoEnabled:    c.MailCleanupInterval > 0,
		SMTPConfigured:            smtpConfigured,
		GmailAPIEnabled:           gmailAPIDirectConfigured,
		GmailAPIRefreshConfigured: gmailAPIRefreshConfigured,
	}
}

func (c Config) isProduction() bool {
	return strings.EqualFold(strings.TrimSpace(c.AppEnv), "production")
}

func validateAppURL(raw string) (valid bool, https bool) {
	if !isPresent(raw) {
		return false, false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed == nil || parsed.Host == "" {
		return false, false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false, false
	}
	return true, parsed.Scheme == "https"
}

func normalizedMailProvider(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		return "smtp"
	}
	return normalized
}

func oauthPairState(clientID string, clientSecret string) (configured bool, valid bool) {
	idConfigured := isPresent(clientID)
	secretConfigured := isPresent(clientSecret)
	if !idConfigured && !secretConfigured {
		return false, true
	}
	return idConfigured && secretConfigured, idConfigured && secretConfigured
}

func looksLikeDefaultDatabaseURL(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return strings.Contains(normalized, "localhost") ||
		strings.Contains(normalized, "127.0.0.1") ||
		strings.Contains(normalized, "codex:codex@") ||
		strings.Contains(normalized, "lavoval:lavoval@postgres")
}

func isStrongSecret(value string) bool {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < 32 {
		return false
	}
	lower := strings.ToLower(trimmed)
	return !strings.Contains(lower, "change-me") && !strings.Contains(lower, "replace-me")
}

func isPresent(value string) bool {
	return strings.TrimSpace(value) != ""
}
