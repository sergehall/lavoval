package auth

import (
	"testing"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/config"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

func testConfig() config.Config {
	return config.Config{
		JWTIssuer:     "test",
		JWTAudience:   "test",
		JWTSecret:     "super-secret",
		JWTAccessTTL:  time.Minute,
		JWTRefreshTTL: time.Hour,
	}
}

func testUser() domain.User {
	return domain.User{
		ID:             "user-1",
		Email:          "user@example.com",
		Role:           domain.RoleUser,
		SessionVersion: 1,
	}
}

func TestTokenManagerIssueAndParseRoundtrip(t *testing.T) {
	mgr := NewTokenManager(testConfig())

	pair, err := mgr.IssueTokens(testUser())
	if err != nil {
		t.Fatalf("expected no error issuing tokens, got %v", err)
	}
	if pair.AccessToken == "" {
		t.Fatal("expected non-empty access token")
	}
	if pair.RefreshToken == "" {
		t.Fatal("expected non-empty refresh token")
	}

	claims, err := mgr.Parse(pair.AccessToken)
	if err != nil {
		t.Fatalf("expected to parse access token, got %v", err)
	}
	if claims.UserID != "user-1" {
		t.Fatalf("expected UserID user-1, got %s", claims.UserID)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %s", claims.Email)
	}
	if claims.Type != "access" {
		t.Fatalf("expected token type access, got %s", claims.Type)
	}
	if claims.Role != domain.RoleUser {
		t.Fatalf("expected role user, got %s", claims.Role)
	}
	if claims.SessionVersion != 1 {
		t.Fatalf("expected session version 1, got %d", claims.SessionVersion)
	}
}

func TestTokenManagerRefreshTokenHasCorrectType(t *testing.T) {
	mgr := NewTokenManager(testConfig())

	pair, err := mgr.IssueTokens(testUser())
	if err != nil {
		t.Fatalf("expected no error issuing tokens, got %v", err)
	}

	claims, err := mgr.Parse(pair.RefreshToken)
	if err != nil {
		t.Fatalf("expected to parse refresh token, got %v", err)
	}
	if claims.Type != "refresh" {
		t.Fatalf("expected token type refresh, got %s", claims.Type)
	}
}

func TestTokenManagerRejectsInvalidToken(t *testing.T) {
	mgr := NewTokenManager(testConfig())

	_, err := mgr.Parse("not-a-valid-token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestTokenManagerRejectsTokenWithWrongSecret(t *testing.T) {
	issuer := NewTokenManager(testConfig())
	pair, err := issuer.IssueTokens(testUser())
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}

	cfg := testConfig()
	cfg.JWTSecret = "different-secret"
	parser := NewTokenManager(cfg)

	_, err = parser.Parse(pair.AccessToken)
	if err == nil {
		t.Fatal("expected error for token signed with wrong secret, got nil")
	}
}
