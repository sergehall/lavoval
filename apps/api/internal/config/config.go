package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv                string
	AppName               string
	AppURL                string
	HTTPAddr              string
	DatabaseURL           string
	JWTIssuer             string
	JWTAudience           string
	JWTSecret             string
	JWTAccessTTL          time.Duration
	JWTRefreshTTL         time.Duration
	CookieSecure          bool
	AdminSeedEmail        string
	AdminSeedSecret       string
	EmailVerificationTTL  time.Duration
	PasswordResetTTL      time.Duration
	SMTPHost              string
	SMTPPort              int
	SMTPUsername          string
	SMTPPassword          string
	SMTPFromEmail         string
	SMTPFromName          string
	SMTPRequireTLS        bool
	SMTPAllowInsecureAuth bool
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

	cfg := Config{
		AppEnv:                getEnv("APP_ENV", "development"),
		AppName:               getEnv("APP_NAME", "Lavoval"),
		AppURL:                getEnv("APP_URL", "http://localhost:3000"),
		HTTPAddr:              getEnv("BACKEND_HTTP_ADDR", ":8080"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://codex:codex@localhost:5432/lavoval?sslmode=disable"),
		JWTIssuer:             getEnv("JWT_ISSUER", "lavoval"),
		JWTAudience:           getEnv("JWT_AUDIENCE", "lavoval-web"),
		JWTSecret:             getEnv("JWT_SECRET", "change-me"),
		JWTAccessTTL:          accessTTL,
		JWTRefreshTTL:         refreshTTL,
		CookieSecure:          cookieSecure,
		AdminSeedEmail:        getEnv("ADMIN_SEED_EMAIL", "admin@lavoval.local"),
		AdminSeedSecret:       getEnv("ADMIN_SEED_PASSWORD", "ChangeMe123!"),
		EmailVerificationTTL:  emailVerificationTTL,
		PasswordResetTTL:      passwordResetTTL,
		SMTPHost:              getEnv("SMTP_HOST", ""),
		SMTPPort:              smtpPort,
		SMTPUsername:          getEnv("SMTP_USERNAME", ""),
		SMTPPassword:          getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:         getEnv("SMTP_FROM_EMAIL", ""),
		SMTPFromName:          getEnv("SMTP_FROM_NAME", "Lavoval"),
		SMTPRequireTLS:        smtpRequireTLS,
		SMTPAllowInsecureAuth: smtpAllowInsecureAuth,
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
