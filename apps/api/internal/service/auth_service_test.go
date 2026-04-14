package service

import (
	"context"
	"testing"

	"github.com/sergehall/lavoval/apps/api/internal/auth"
	"github.com/sergehall/lavoval/apps/api/internal/config"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type authUserRepoStub struct {
	user domain.User
}

func (s authUserRepoStub) Create(_ context.Context, user domain.User) (domain.User, error) {
	user.CreatedAt = now()
	user.UpdatedAt = now()
	return user, nil
}
func (s authUserRepoStub) FindByEmail(_ context.Context, _ string) (domain.User, error) {
	return s.user, nil
}
func (s authUserRepoStub) FindByID(_ context.Context, _ string) (domain.User, error) {
	return s.user, nil
}
func (s authUserRepoStub) List(_ context.Context) ([]domain.User, error) {
	return []domain.User{s.user}, nil
}

type authProfileRepoStub struct {
	profile domain.Profile
}

func (s authProfileRepoStub) Create(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	return profile, nil
}
func (s authProfileRepoStub) Update(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	return profile, nil
}
func (s authProfileRepoStub) FindByUserID(_ context.Context, _ string) (domain.Profile, error) {
	return s.profile, nil
}

func TestAuthServiceRegisterReturnsTokens(t *testing.T) {
	cfg := config.Config{JWTIssuer: "test", JWTAudience: "test", JWTSecret: "super-secret", JWTAccessTTL: 1, JWTRefreshTTL: 2}
	service := NewAuthService(authUserRepoStub{}, authProfileRepoStub{}, auth.NewTokenManager(cfg), cfg)

	result, err := service.Register(context.Background(), RegisterInput{
		Email:     "user@example.com",
		Password:  "SuperSecurePass123",
		FirstName: "Ada",
		LastName:  "Lovelace",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.AccessToken == "" || result.RefreshToken == "" {
		t.Fatal("expected token pair to be generated")
	}
	if result.User.Email != "user@example.com" {
		t.Fatalf("unexpected user email %s", result.User.Email)
	}
}
