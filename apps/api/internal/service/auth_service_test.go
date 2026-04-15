package service

import (
	"context"
	"testing"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/auth"
	"github.com/sergehall/lavoval/apps/api/internal/config"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/mailer"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type authUserRepoStub struct {
	user        domain.User
	findByEmail func(string) (domain.User, error)
}

func (s authUserRepoStub) Create(_ context.Context, user domain.User) (domain.User, error) {
	user.CreatedAt = now()
	user.UpdatedAt = now()
	return user, nil
}

func (s authUserRepoStub) FindByEmail(_ context.Context, email string) (domain.User, error) {
	if s.findByEmail != nil {
		return s.findByEmail(email)
	}
	return s.user, nil
}

func (s authUserRepoStub) FindByID(_ context.Context, _ string) (domain.User, error) {
	return s.user, nil
}

func (s authUserRepoStub) MarkEmailVerified(_ context.Context, _ string) (domain.User, error) {
	verified := now()
	s.user.EmailVerifiedAt = &verified
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

type verificationRepoStub struct {
	tokenCreated bool
	token        domain.EmailVerificationToken
}

func (s *verificationRepoStub) Create(_ context.Context, token domain.EmailVerificationToken) (domain.EmailVerificationToken, error) {
	s.tokenCreated = true
	s.token = token
	s.token.CreatedAt = now()
	return s.token, nil
}

func (s *verificationRepoStub) FindByTokenHash(_ context.Context, tokenHash string) (domain.EmailVerificationToken, error) {
	s.token.TokenHash = tokenHash
	return s.token, nil
}

func (s *verificationRepoStub) Consume(_ context.Context, _, _ string) error {
	return nil
}

func (s *verificationRepoStub) RevokeActiveByUserID(_ context.Context, _ string) error {
	return nil
}

type verificationMailerStub struct {
	lastEmail mailer.VerificationEmail
}

func (s *verificationMailerStub) SendVerificationEmail(_ context.Context, email mailer.VerificationEmail) error {
	s.lastEmail = email
	return nil
}

func TestAuthServiceRegisterReturnsVerificationRequirement(t *testing.T) {
	cfg := config.Config{
		AppName:              "Lavoval",
		AppURL:               "http://localhost:3000",
		JWTIssuer:            "test",
		JWTAudience:          "test",
		JWTSecret:            "super-secret",
		JWTAccessTTL:         time.Minute,
		JWTRefreshTTL:        time.Hour,
		EmailVerificationTTL: 24 * time.Hour,
	}
	verificationRepo := &verificationRepoStub{}
	verificationMailer := &verificationMailerStub{}
	service := NewAuthService(
		authUserRepoStub{findByEmail: func(string) (domain.User, error) { return domain.User{}, pgx.ErrNoRows }},
		authProfileRepoStub{},
		verificationRepo,
		auth.NewTokenManager(cfg),
		verificationMailer,
		cfg,
	)

	result, err := service.Register(context.Background(), RegisterInput{
		Email:     "user@example.com",
		Password:  "SuperSecurePass123",
		FirstName: "Ada",
		LastName:  "Lovelace",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.VerificationRequired {
		t.Fatal("expected registration to require email verification")
	}
	if result.Email != "user@example.com" {
		t.Fatalf("unexpected email %s", result.Email)
	}
	if !verificationRepo.tokenCreated {
		t.Fatal("expected email verification token to be created")
	}
	if verificationMailer.lastEmail.ToEmail != "user@example.com" {
		t.Fatalf("expected verification email to target registered email, got %s", verificationMailer.lastEmail.ToEmail)
	}
}

func TestAuthServiceLoginRejectsUnverifiedEmail(t *testing.T) {
	cfg := config.Config{
		JWTIssuer:     "test",
		JWTAudience:   "test",
		JWTSecret:     "super-secret",
		JWTAccessTTL:  time.Minute,
		JWTRefreshTTL: time.Hour,
	}

	service := NewAuthService(
		authUserRepoStub{
			user: domain.User{
				ID:           "user-1",
				Email:        "user@example.com",
				PasswordHash: mustHashPassword(t, "SuperSecurePass123"),
			},
		},
		authProfileRepoStub{},
		&verificationRepoStub{},
		auth.NewTokenManager(cfg),
		&verificationMailerStub{},
		cfg,
	)

	_, err := service.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "SuperSecurePass123",
	})
	if err == nil {
		t.Fatal("expected unverified login to fail")
	}
	if err != ErrEmailNotVerified {
		t.Fatalf("expected ErrEmailNotVerified, got %v", err)
	}
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	return string(hash)
}
