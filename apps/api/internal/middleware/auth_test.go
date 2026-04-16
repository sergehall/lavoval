package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/auth"
	"github.com/sergehall/lavoval/apps/api/internal/config"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type authMiddlewareUserStoreStub struct {
	user domain.User
}

func (s authMiddlewareUserStoreStub) Create(_ context.Context, u domain.User) (domain.User, error) {
	return u, nil
}
func (s authMiddlewareUserStoreStub) FindByEmail(_ context.Context, _ string) (domain.User, error) {
	return s.user, nil
}
func (s authMiddlewareUserStoreStub) FindByID(_ context.Context, _ string) (domain.User, error) {
	return s.user, nil
}
func (s authMiddlewareUserStoreStub) MarkEmailVerified(_ context.Context, _ string) (domain.User, error) {
	return s.user, nil
}
func (s authMiddlewareUserStoreStub) UpdatePasswordHash(_ context.Context, _, _ string) (domain.User, error) {
	return s.user, nil
}
func (s authMiddlewareUserStoreStub) UpdateRoleAndStatus(_ context.Context, _ string, _ domain.Role, _ domain.AccountStatus) (domain.User, error) {
	return s.user, nil
}
func (s authMiddlewareUserStoreStub) UpdateRoleStatusModeration(_ context.Context, _, _ string, _ domain.Role, _ domain.AccountStatus, _ *string) (domain.User, error) {
	return s.user, nil
}
func (s authMiddlewareUserStoreStub) BumpSessionVersion(_ context.Context, _ string) (domain.User, error) {
	u := s.user
	u.SessionVersion++
	return u, nil
}
func (s authMiddlewareUserStoreStub) GetStats(_ context.Context) (domain.AdminUserStats, error) {
	return domain.AdminUserStats{}, nil
}
func (s authMiddlewareUserStoreStub) StartTOTPEnrollment(_ context.Context, _ string, _ string) (domain.User, error) {
	return s.user, nil
}
func (s authMiddlewareUserStoreStub) CancelTOTPEnrollment(_ context.Context, _ string) (domain.User, error) {
	return s.user, nil
}
func (s authMiddlewareUserStoreStub) EnableTOTP(_ context.Context, _ string, _ string) (domain.User, error) {
	return s.user, nil
}
func (s authMiddlewareUserStoreStub) DisableTOTP(_ context.Context, _ string) (domain.User, error) {
	return s.user, nil
}
func (s authMiddlewareUserStoreStub) SoftDelete(_ context.Context, _ string) error {
	return nil
}
func (s authMiddlewareUserStoreStub) List(_ context.Context) ([]domain.User, error) {
	return []domain.User{s.user}, nil
}

func testMiddlewareTokenManager() auth.TokenManager {
	return auth.NewTokenManager(config.Config{
		JWTIssuer:     "test",
		JWTAudience:   "test",
		JWTSecret:     "super-secret",
		JWTAccessTTL:  time.Minute,
		JWTRefreshTTL: time.Hour,
	})
}

func TestAuthenticateAcceptsMatchingSessionVersion(t *testing.T) {
	user := domain.User{
		ID:             "user-1",
		Email:          "user@example.com",
		Role:           domain.RoleAdmin,
		SessionVersion: 3,
	}
	tokens := testMiddlewareTokenManager()
	pair, err := tokens.IssueTokens(user)
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}

	called := false
	handler := Authenticate(tokens, authMiddlewareUserStoreStub{user: user})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if !called {
		t.Fatal("expected downstream handler to be called")
	}
}

func TestAuthenticateRejectsRevokedSessionVersion(t *testing.T) {
	tokens := testMiddlewareTokenManager()
	pair, err := tokens.IssueTokens(domain.User{
		ID:             "user-1",
		Email:          "user@example.com",
		Role:           domain.RoleAdmin,
		SessionVersion: 2,
	})
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}

	handler := Authenticate(tokens, authMiddlewareUserStoreStub{user: domain.User{
		ID:             "user-1",
		Email:          "user@example.com",
		Role:           domain.RoleAdmin,
		SessionVersion: 3,
	}})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("did not expect downstream handler to be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
