package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv          string
	AppName         string
	HTTPAddr        string
	DatabaseURL     string
	JWTIssuer       string
	JWTAudience     string
	JWTSecret       string
	JWTAccessTTL    time.Duration
	JWTRefreshTTL   time.Duration
	CookieSecure    bool
	AdminSeedEmail  string
	AdminSeedSecret string
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

	cfg := Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		AppName:         getEnv("APP_NAME", "Lavoval"),
		HTTPAddr:        getEnv("BACKEND_HTTP_ADDR", ":8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://codex:codex@localhost:5432/lavoval?sslmode=disable"),
		JWTIssuer:       getEnv("JWT_ISSUER", "lavoval"),
		JWTAudience:     getEnv("JWT_AUDIENCE", "lavoval-web"),
		JWTSecret:       getEnv("JWT_SECRET", "change-me"),
		JWTAccessTTL:    accessTTL,
		JWTRefreshTTL:   refreshTTL,
		CookieSecure:    cookieSecure,
		AdminSeedEmail:  getEnv("ADMIN_SEED_EMAIL", "admin@lavoval.local"),
		AdminSeedSecret: getEnv("ADMIN_SEED_PASSWORD", "ChangeMe123!"),
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
