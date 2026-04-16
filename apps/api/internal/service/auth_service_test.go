package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/sergehall/lavoval/apps/api/internal/auth"
	"github.com/sergehall/lavoval/apps/api/internal/config"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/mailer"
)

type authUserRepoStub struct {
	user                 domain.User
	findByEmail          func(string) (domain.User, error)
	updatePasswordHash   func(string, string) (domain.User, error)
	startTOTPEnrollment  func(string, string) (domain.User, error)
	cancelTOTPEnrollment func(string) (domain.User, error)
	enableTOTP           func(string, string) (domain.User, error)
	disableTOTP          func(string) (domain.User, error)
	softDeletedID        string
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

func (s authUserRepoStub) UpdateRoleAndStatus(_ context.Context, _ string, _ domain.Role, _ domain.AccountStatus) (domain.User, error) {
	return s.user, nil
}

func (s authUserRepoStub) UpdatePasswordHash(_ context.Context, userID string, passwordHash string) (domain.User, error) {
	if s.updatePasswordHash != nil {
		return s.updatePasswordHash(userID, passwordHash)
	}

	s.user.ID = userID
	s.user.PasswordHash = passwordHash
	return s.user, nil
}

func (s authUserRepoStub) StartTOTPEnrollment(_ context.Context, userID string, pendingSecretEncrypted string) (domain.User, error) {
	if s.startTOTPEnrollment != nil {
		return s.startTOTPEnrollment(userID, pendingSecretEncrypted)
	}

	s.user.ID = userID
	s.user.MFAPendingTOTPSecretEncrypted = &pendingSecretEncrypted
	return s.user, nil
}

func (s authUserRepoStub) CancelTOTPEnrollment(_ context.Context, userID string) (domain.User, error) {
	if s.cancelTOTPEnrollment != nil {
		return s.cancelTOTPEnrollment(userID)
	}

	s.user.ID = userID
	s.user.MFAPendingTOTPSecretEncrypted = nil
	return s.user, nil
}

func (s authUserRepoStub) EnableTOTP(_ context.Context, userID string, secretEncrypted string) (domain.User, error) {
	if s.enableTOTP != nil {
		return s.enableTOTP(userID, secretEncrypted)
	}

	enrolledAt := now()
	s.user.ID = userID
	s.user.MFAEnabled = true
	s.user.MFATOTPSecretEncrypted = &secretEncrypted
	s.user.MFAPendingTOTPSecretEncrypted = nil
	s.user.MFAEnrolledAt = &enrolledAt
	return s.user, nil
}

func (s authUserRepoStub) DisableTOTP(_ context.Context, userID string) (domain.User, error) {
	if s.disableTOTP != nil {
		return s.disableTOTP(userID)
	}

	s.user.ID = userID
	s.user.MFAEnabled = false
	s.user.MFATOTPSecretEncrypted = nil
	s.user.MFAPendingTOTPSecretEncrypted = nil
	s.user.MFAEnrolledAt = nil
	return s.user, nil
}

func (s authUserRepoStub) SoftDelete(_ context.Context, id string) error {
	s.softDeletedID = id
	return nil
}
func (s authUserRepoStub) UpdateRoleStatusModeration(_ context.Context, _, _ string, role domain.Role, status domain.AccountStatus, _ *string) (domain.User, error) {
	u := s.user
	u.Role = role
	u.Status = status
	return u, nil
}
func (s authUserRepoStub) BumpSessionVersion(_ context.Context, _ string) (domain.User, error) {
	u := s.user
	u.SessionVersion++
	return u, nil
}
func (authUserRepoStub) GetStats(_ context.Context) (domain.AdminUserStats, error) {
	return domain.AdminUserStats{}, nil
}

type authProfileRepoStub struct {
	profile              domain.Profile
	softDeletedForUserID string
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

func (s authProfileRepoStub) SoftDeleteByUserID(_ context.Context, userID string) error {
	s.softDeletedForUserID = userID
	return nil
}

type verificationRepoStub struct {
	tokenCreated bool
	token        domain.EmailVerificationToken
	consumeErr   error
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
	return s.consumeErr
}

func (s *verificationRepoStub) RevokeActiveByUserID(_ context.Context, _ string) error {
	return nil
}

type passwordResetRepoStub struct {
	tokenCreated bool
	token        domain.PasswordResetToken
}

func (s *passwordResetRepoStub) Create(_ context.Context, token domain.PasswordResetToken) (domain.PasswordResetToken, error) {
	s.tokenCreated = true
	s.token = token
	s.token.CreatedAt = now()
	return s.token, nil
}

func (s *passwordResetRepoStub) FindByTokenHash(_ context.Context, tokenHash string) (domain.PasswordResetToken, error) {
	s.token.TokenHash = tokenHash
	return s.token, nil
}

func (s *passwordResetRepoStub) Consume(_ context.Context, _, _ string) error {
	return nil
}

func (s *passwordResetRepoStub) RevokeActiveByUserID(_ context.Context, _ string) error {
	return nil
}

type mfaRecoveryCodeRepoStub struct {
	codes []domain.MFARecoveryCode
}

func (s *mfaRecoveryCodeRepoStub) ReplaceForUser(_ context.Context, _ string, codes []domain.MFARecoveryCode) error {
	s.codes = codes
	return nil
}

func (s *mfaRecoveryCodeRepoStub) FindActiveByCodeHash(_ context.Context, userID string, codeHash string) (domain.MFARecoveryCode, error) {
	for _, code := range s.codes {
		if code.UserID == userID && code.CodeHash == codeHash && code.ConsumedAt == nil {
			return code, nil
		}
	}
	return domain.MFARecoveryCode{}, pgx.ErrNoRows
}

func (s *mfaRecoveryCodeRepoStub) Consume(_ context.Context, id string, userID string) error {
	for index, code := range s.codes {
		if code.ID == id && code.UserID == userID {
			consumedAt := now()
			s.codes[index].ConsumedAt = &consumedAt
			return nil
		}
	}
	return pgx.ErrNoRows
}

func (s *mfaRecoveryCodeRepoStub) RevokeActiveByUserID(_ context.Context, userID string) error {
	for index, code := range s.codes {
		if code.UserID == userID && code.ConsumedAt == nil {
			consumedAt := now()
			s.codes[index].ConsumedAt = &consumedAt
		}
	}
	return nil
}

type signInChallengeRepoStub struct {
	challenge domain.AuthSignInChallenge
}

func (s *signInChallengeRepoStub) Create(_ context.Context, challenge domain.AuthSignInChallenge) (domain.AuthSignInChallenge, error) {
	s.challenge = challenge
	s.challenge.CreatedAt = now()
	return s.challenge, nil
}

func (s *signInChallengeRepoStub) FindByID(_ context.Context, id string) (domain.AuthSignInChallenge, error) {
	if s.challenge.ID == id {
		return s.challenge, nil
	}
	return domain.AuthSignInChallenge{}, pgx.ErrNoRows
}

func (s *signInChallengeRepoStub) Consume(_ context.Context, id string, userID string) error {
	if s.challenge.ID != id || s.challenge.UserID != userID {
		return pgx.ErrNoRows
	}
	consumedAt := now()
	s.challenge.ConsumedAt = &consumedAt
	return nil
}

func (s *signInChallengeRepoStub) RevokeActiveByUserID(_ context.Context, userID string) error {
	if s.challenge.UserID == userID && s.challenge.ConsumedAt == nil {
		consumedAt := now()
		s.challenge.ConsumedAt = &consumedAt
	}
	return nil
}

type oauthStateRepoStub struct {
	state domain.OAuthState
}

func (s *oauthStateRepoStub) Create(_ context.Context, state domain.OAuthState) (domain.OAuthState, error) {
	s.state = state
	s.state.CreatedAt = now()
	return s.state, nil
}

func (s *oauthStateRepoStub) FindByStateHash(_ context.Context, _ domain.OAuthProvider, stateHash string) (domain.OAuthState, error) {
	if s.state.StateHash == stateHash {
		return s.state, nil
	}
	return domain.OAuthState{}, pgx.ErrNoRows
}

func (s *oauthStateRepoStub) Consume(_ context.Context, id string) error {
	if s.state.ID != id {
		return pgx.ErrNoRows
	}
	consumedAt := now()
	s.state.ConsumedAt = &consumedAt
	return nil
}

type oauthIdentityRepoStub struct {
	identity domain.OAuthIdentity
}

func (s *oauthIdentityRepoStub) Create(_ context.Context, identity domain.OAuthIdentity) (domain.OAuthIdentity, error) {
	s.identity = identity
	s.identity.CreatedAt = now()
	s.identity.UpdatedAt = now()
	return s.identity, nil
}

func (s *oauthIdentityRepoStub) FindByProviderSubject(_ context.Context, provider domain.OAuthProvider, providerUserID string) (domain.OAuthIdentity, error) {
	if s.identity.Provider == provider && s.identity.ProviderUserID == providerUserID {
		return s.identity, nil
	}
	return domain.OAuthIdentity{}, pgx.ErrNoRows
}

func (s *oauthIdentityRepoStub) ListByUserID(_ context.Context, userID string) ([]domain.OAuthIdentity, error) {
	if s.identity.UserID == userID {
		return []domain.OAuthIdentity{s.identity}, nil
	}
	return nil, nil
}

type googleOAuthProviderStub struct {
	identity oauthIdentityProfile
}

func (s googleOAuthProviderStub) Enabled() bool {
	return true
}

func (s googleOAuthProviderStub) AuthorizationURL(state string) string {
	return "https://accounts.google.com/o/oauth2/v2/auth?state=" + state
}

func (s googleOAuthProviderStub) ExchangeCode(_ context.Context, _ string) (oauthIdentityProfile, error) {
	return s.identity, nil
}

type verificationMailerStub struct {
	lastEmail           mailer.VerificationEmail
	lastPasswordReset   mailer.PasswordResetEmail
	lastPasswordChanged mailer.PasswordChangedEmail
}

type sessionRevokerStub struct {
	lastUserID string
	lastReason string
}

func (s *sessionRevokerStub) RevokeAllSessionsForUser(_ context.Context, userID string, reason string) error {
	s.lastUserID = userID
	s.lastReason = reason
	return nil
}

func (s *verificationMailerStub) SendVerificationEmail(_ context.Context, email mailer.VerificationEmail) error {
	s.lastEmail = email
	return nil
}

func (s *verificationMailerStub) SendPasswordResetEmail(_ context.Context, email mailer.PasswordResetEmail) error {
	s.lastPasswordReset = email
	return nil
}

func (s *verificationMailerStub) SendPasswordChangedEmail(_ context.Context, email mailer.PasswordChangedEmail) error {
	s.lastPasswordChanged = email
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
		&passwordResetRepoStub{},
		nil,
		nil,
		nil,
		nil,
		auth.NewTokenManager(cfg),
		verificationMailer,
		NoopSessionRevoker{},
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
		&passwordResetRepoStub{},
		nil,
		nil,
		nil,
		nil,
		auth.NewTokenManager(cfg),
		&verificationMailerStub{},
		NoopSessionRevoker{},
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

func TestAuthServiceRegisterRejectsExistingEmail(t *testing.T) {
	cfg := config.Config{
		JWTIssuer:            "test",
		JWTAudience:          "test",
		JWTSecret:            "super-secret",
		JWTAccessTTL:         time.Minute,
		JWTRefreshTTL:        time.Hour,
		EmailVerificationTTL: 24 * time.Hour,
	}
	service := NewAuthService(
		authUserRepoStub{user: domain.User{ID: "user-1", Email: "existing@example.com"}},
		authProfileRepoStub{},
		&verificationRepoStub{},
		&passwordResetRepoStub{},
		nil,
		nil,
		nil,
		nil,
		auth.NewTokenManager(cfg),
		&verificationMailerStub{},
		NoopSessionRevoker{},
		cfg,
	)

	_, err := service.Register(context.Background(), RegisterInput{
		Email:     "existing@example.com",
		Password:  "SuperSecurePass123",
		FirstName: "Ada",
		LastName:  "Lovelace",
	})
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestAuthServiceLoginSucceeds(t *testing.T) {
	verified := now()
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
				ID:              "user-1",
				Email:           "user@example.com",
				PasswordHash:    mustHashPassword(t, "SuperSecurePass123"),
				EmailVerifiedAt: &verified,
			},
		},
		authProfileRepoStub{profile: domain.Profile{FirstName: "Ada", LastName: "Lovelace"}},
		&verificationRepoStub{},
		&passwordResetRepoStub{},
		nil,
		nil,
		nil,
		nil,
		auth.NewTokenManager(cfg),
		&verificationMailerStub{},
		NoopSessionRevoker{},
		cfg,
	)

	payload, err := service.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "SuperSecurePass123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if payload.AccessToken == "" {
		t.Fatal("expected non-empty access token")
	}
	if payload.User.Email != "user@example.com" {
		t.Fatalf("unexpected email %s", payload.User.Email)
	}
	if payload.User.FirstName != "Ada" {
		t.Fatalf("expected firstName Ada, got %s", payload.User.FirstName)
	}
}

func TestAuthServiceLoginRejectsBlockedAccount(t *testing.T) {
	verified := now()
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
				ID:              "user-1",
				Email:           "user@example.com",
				PasswordHash:    mustHashPassword(t, "SuperSecurePass123"),
				EmailVerifiedAt: &verified,
				Status:          domain.AccountStatusBlocked,
			},
		},
		authProfileRepoStub{profile: domain.Profile{FirstName: "Ada", LastName: "Lovelace"}},
		&verificationRepoStub{},
		&passwordResetRepoStub{},
		nil,
		nil,
		nil,
		nil,
		auth.NewTokenManager(cfg),
		&verificationMailerStub{},
		NoopSessionRevoker{},
		cfg,
	)

	_, err := service.Login(context.Background(), LoginInput{
		Email:    "user@example.com",
		Password: "SuperSecurePass123",
	})
	if !errors.Is(err, ErrAccountBlocked) {
		t.Fatalf("expected ErrAccountBlocked, got %v", err)
	}
}

func TestAuthServiceForgotPasswordIssuesResetEmail(t *testing.T) {
	cfg := config.Config{
		AppName:              "Lavoval",
		AppURL:               "http://localhost:3000",
		JWTIssuer:            "test",
		JWTAudience:          "test",
		JWTSecret:            "super-secret",
		JWTAccessTTL:         time.Minute,
		JWTRefreshTTL:        time.Hour,
		EmailVerificationTTL: 24 * time.Hour,
		PasswordResetTTL:     30 * time.Minute,
	}

	passwordResetRepo := &passwordResetRepoStub{}
	verificationMailer := &verificationMailerStub{}
	service := NewAuthService(
		authUserRepoStub{
			user: domain.User{
				ID:    "user-1",
				Email: "user@example.com",
			},
		},
		authProfileRepoStub{profile: domain.Profile{FirstName: "Ada", LastName: "Lovelace"}},
		&verificationRepoStub{},
		passwordResetRepo,
		nil,
		nil,
		nil,
		nil,
		auth.NewTokenManager(cfg),
		verificationMailer,
		NoopSessionRevoker{},
		cfg,
	)

	result, err := service.ForgotPassword(context.Background(), ForgotPasswordInput{Email: "user@example.com"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Sent {
		t.Fatal("expected forgot password response to mark email as sent")
	}
	if !passwordResetRepo.tokenCreated {
		t.Fatal("expected password reset token to be created")
	}
	if verificationMailer.lastPasswordReset.ToEmail != "user@example.com" {
		t.Fatalf("expected password reset email to target requested email, got %s", verificationMailer.lastPasswordReset.ToEmail)
	}
	if verificationMailer.lastPasswordReset.ResetURL == "" {
		t.Fatal("expected password reset email to include reset URL")
	}
}

func TestAuthServiceVerifyEmailExpiredToken(t *testing.T) {
	cfg := config.Config{
		JWTIssuer:     "test",
		JWTAudience:   "test",
		JWTSecret:     "super-secret",
		JWTAccessTTL:  time.Minute,
		JWTRefreshTTL: time.Hour,
	}
	pastTime := now().Add(-time.Hour)
	service := NewAuthService(
		authUserRepoStub{},
		authProfileRepoStub{},
		&verificationRepoStub{
			token: domain.EmailVerificationToken{
				ID:        "token-1",
				UserID:    "user-1",
				ExpiresAt: pastTime,
			},
		},
		&passwordResetRepoStub{},
		nil,
		nil,
		nil,
		nil,
		auth.NewTokenManager(cfg),
		&verificationMailerStub{},
		NoopSessionRevoker{},
		cfg,
	)

	_, err := service.VerifyEmail(context.Background(), VerifyEmailInput{Token: "some-plain-token-value"})
	if !errors.Is(err, ErrVerificationTokenExpired) {
		t.Fatalf("expected ErrVerificationTokenExpired, got %v", err)
	}
}

func TestAuthServiceVerifyEmailReturnsAlreadyVerifiedForConsumedToken(t *testing.T) {
	cfg := config.Config{
		JWTIssuer:     "test",
		JWTAudience:   "test",
		JWTSecret:     "super-secret",
		JWTAccessTTL:  time.Minute,
		JWTRefreshTTL: time.Hour,
	}
	verified := now()
	service := NewAuthService(
		authUserRepoStub{
			user: domain.User{
				ID:              "user-1",
				Email:           "user@example.com",
				EmailVerifiedAt: &verified,
			},
		},
		authProfileRepoStub{},
		&verificationRepoStub{
			token: domain.EmailVerificationToken{
				ID:         "token-1",
				UserID:     "user-1",
				ConsumedAt: &verified,
				ExpiresAt:  now().Add(time.Hour),
			},
		},
		&passwordResetRepoStub{},
		nil,
		nil,
		nil,
		nil,
		auth.NewTokenManager(cfg),
		&verificationMailerStub{},
		NoopSessionRevoker{},
		cfg,
	)

	result, err := service.VerifyEmail(context.Background(), VerifyEmailInput{Token: "some-plain-token-value"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.AlreadyVerified {
		t.Fatal("expected already verified response")
	}
	if result.Email != "user@example.com" {
		t.Fatalf("unexpected email %s", result.Email)
	}
}

func TestAuthServiceVerifyEmailReturnsAlreadyVerifiedWhenConsumeRaces(t *testing.T) {
	cfg := config.Config{
		JWTIssuer:     "test",
		JWTAudience:   "test",
		JWTSecret:     "super-secret",
		JWTAccessTTL:  time.Minute,
		JWTRefreshTTL: time.Hour,
	}
	verified := now()
	service := NewAuthService(
		authUserRepoStub{
			user: domain.User{
				ID:              "user-1",
				Email:           "user@example.com",
				EmailVerifiedAt: &verified,
			},
		},
		authProfileRepoStub{},
		&verificationRepoStub{
			token: domain.EmailVerificationToken{
				ID:        "token-1",
				UserID:    "user-1",
				ExpiresAt: now().Add(time.Hour),
			},
			consumeErr: errors.New("consume email verification token: no rows affected"),
		},
		&passwordResetRepoStub{},
		nil,
		nil,
		nil,
		nil,
		auth.NewTokenManager(cfg),
		&verificationMailerStub{},
		NoopSessionRevoker{},
		cfg,
	)

	result, err := service.VerifyEmail(context.Background(), VerifyEmailInput{Token: "some-plain-token-value"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.AlreadyVerified {
		t.Fatal("expected already verified response")
	}
	if result.Email != "user@example.com" {
		t.Fatalf("unexpected email %s", result.Email)
	}
}

func TestAuthServiceEnrollMFAStartsPendingEnrollment(t *testing.T) {
	cfg := config.Config{
		AppName:       "Lavoval",
		MFATOTPIssuer: "Lavoval",
		MFATOTPPeriod: 30 * time.Second,
		JWTSecret:     "super-secret",
	}

	var storedPendingSecret string
	service := NewAuthService(
		authUserRepoStub{
			user: domain.User{
				ID:    "user-1",
				Email: "user@example.com",
			},
			startTOTPEnrollment: func(_ string, pending string) (domain.User, error) {
				storedPendingSecret = pending
				return domain.User{
					ID:                            "user-1",
					Email:                         "user@example.com",
					MFAPendingTOTPSecretEncrypted: &pending,
				}, nil
			},
		},
		authProfileRepoStub{},
		&verificationRepoStub{},
		&passwordResetRepoStub{},
		nil,
		nil,
		nil,
		nil,
		auth.NewTokenManager(cfg),
		&verificationMailerStub{},
		NoopSessionRevoker{},
		cfg,
	)

	result, err := service.EnrollMFA(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Secret == "" {
		t.Fatal("expected a generated MFA secret")
	}
	if result.ProvisionURL == "" {
		t.Fatal("expected an otpauth provisioning URL")
	}
	if storedPendingSecret == "" {
		t.Fatal("expected encrypted pending secret to be stored")
	}
}

func TestAuthServiceCancelMFAEnrollmentClearsPendingState(t *testing.T) {
	pending := "pending-secret"
	service := NewAuthService(
		authUserRepoStub{
			user: domain.User{
				ID:                            "user-1",
				Email:                         "member@example.com",
				MFAPendingTOTPSecretEncrypted: &pending,
			},
		},
		authProfileRepoStub{},
		&verificationRepoStub{},
		&passwordResetRepoStub{},
		nil,
		nil,
		nil,
		nil,
		auth.NewTokenManager(config.Config{JWTIssuer: "test", JWTAudience: "test", JWTSecret: "secret", JWTAccessTTL: time.Minute, JWTRefreshTTL: time.Hour}),
		&verificationMailerStub{},
		NoopSessionRevoker{},
		config.Config{JWTIssuer: "test", JWTAudience: "test", JWTSecret: "secret", JWTAccessTTL: time.Minute, JWTRefreshTTL: time.Hour},
	)

	result, err := service.CancelMFAEnrollment(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.PendingEnrollment {
		t.Fatal("expected pending enrollment to be cleared")
	}
	if result.Enabled {
		t.Fatal("expected mfa to remain disabled")
	}
}

func TestAuthServiceVerifyMFAEnrollmentEnablesMFA(t *testing.T) {
	cfg := config.Config{
		AppName:       "Lavoval",
		MFATOTPIssuer: "Lavoval",
		MFATOTPPeriod: 30 * time.Second,
		JWTSecret:     "super-secret",
	}

	totp := newTOTPManager(cfg)
	secret := "JBSWY3DPEHPK3PXP"
	encryptedSecret, err := totp.EncryptSecret(secret)
	if err != nil {
		t.Fatalf("encrypt secret: %v", err)
	}

	code, err := generateTOTPCode(secret, time.Now().UTC().Unix()/30)
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}

	service := NewAuthService(
		authUserRepoStub{
			user: domain.User{
				ID:                            "user-1",
				Email:                         "user@example.com",
				MFAPendingTOTPSecretEncrypted: &encryptedSecret,
			},
		},
		authProfileRepoStub{},
		&verificationRepoStub{},
		&passwordResetRepoStub{},
		&mfaRecoveryCodeRepoStub{},
		nil,
		nil,
		nil,
		auth.NewTokenManager(cfg),
		&verificationMailerStub{},
		NoopSessionRevoker{},
		cfg,
	)

	result, err := service.VerifyMFAEnrollment(context.Background(), "user-1", MFAVerifyEnrollmentInput{Code: code})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Enabled {
		t.Fatal("expected MFA to be enabled")
	}
	if result.PendingEnrollment {
		t.Fatal("expected pending enrollment to be cleared")
	}
	if result.EnrolledAt == nil {
		t.Fatal("expected enrolledAt to be set")
	}
}

func TestAuthServiceDisableMFARequiresPasswordAndCode(t *testing.T) {
	cfg := config.Config{
		AppName:       "Lavoval",
		MFATOTPIssuer: "Lavoval",
		MFATOTPPeriod: 30 * time.Second,
		JWTSecret:     "super-secret",
	}

	totp := newTOTPManager(cfg)
	secret := "JBSWY3DPEHPK3PXP"
	encryptedSecret, err := totp.EncryptSecret(secret)
	if err != nil {
		t.Fatalf("encrypt secret: %v", err)
	}

	code, err := generateTOTPCode(secret, time.Now().UTC().Unix()/30)
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}

	sessionRevoker := &sessionRevokerStub{}
	service := NewAuthService(
		authUserRepoStub{
			user: domain.User{
				ID:                     "user-1",
				Email:                  "user@example.com",
				PasswordHash:           mustHashPassword(t, "SuperSecurePass123"),
				MFAEnabled:             true,
				MFATOTPSecretEncrypted: &encryptedSecret,
				MFAEnrolledAt:          ptrTime(now()),
			},
		},
		authProfileRepoStub{},
		&verificationRepoStub{},
		&passwordResetRepoStub{},
		&mfaRecoveryCodeRepoStub{},
		nil,
		nil,
		nil,
		auth.NewTokenManager(cfg),
		&verificationMailerStub{},
		sessionRevoker,
		cfg,
	)

	result, err := service.DisableMFA(context.Background(), "user-1", MFADisableInput{
		Password: "SuperSecurePass123",
		Code:     code,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Enabled {
		t.Fatal("expected MFA to be disabled")
	}
	if sessionRevoker.lastReason != "mfa_disabled" {
		t.Fatalf("expected sessions to be revoked with mfa_disabled, got %q", sessionRevoker.lastReason)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func TestAuthServiceCompleteGoogleOAuthCreatesIdentityAndSession(t *testing.T) {
	cfg := config.Config{
		AppName:             "Lavoval",
		AppURL:              "http://localhost:3000",
		JWTIssuer:           "test",
		JWTAudience:         "test",
		JWTSecret:           "super-secret",
		JWTAccessTTL:        time.Minute,
		JWTRefreshTTL:       time.Hour,
		OAuthStateTTL:       10 * time.Minute,
		GoogleOAuthClientID: "google-client-id",
		GoogleOAuthSecret:   "google-secret",
	}

	state, stateHash, err := newOAuthStateToken()
	if err != nil {
		t.Fatalf("new oauth state token: %v", err)
	}

	oauthStates := &oauthStateRepoStub{
		state: domain.OAuthState{
			ID:        "state-1",
			Provider:  domain.OAuthProviderGoogle,
			StateHash: stateHash,
			ExpiresAt: time.Now().Add(5 * time.Minute),
		},
	}
	oauthIdentities := &oauthIdentityRepoStub{}

	service := NewAuthService(
		authUserRepoStub{findByEmail: func(string) (domain.User, error) { return domain.User{}, pgx.ErrNoRows }},
		authProfileRepoStub{},
		&verificationRepoStub{},
		&passwordResetRepoStub{},
		nil,
		nil,
		oauthStates,
		oauthIdentities,
		auth.NewTokenManager(cfg),
		&verificationMailerStub{},
		NoopSessionRevoker{},
		cfg,
	)
	service.googleOAuth = googleOAuthProviderStub{
		identity: oauthIdentityProfile{
			Subject:       "google-subject-1",
			Email:         "google.user@example.com",
			EmailVerified: true,
			GivenName:     "Google",
			FamilyName:    "User",
		},
	}

	payload, err := service.CompleteGoogleOAuth(context.Background(), CompleteGoogleOAuthInput{
		Code:  "oauth-code-value",
		State: state,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if payload.User.Email != "google.user@example.com" {
		t.Fatalf("unexpected oauth user email %q", payload.User.Email)
	}
	if oauthIdentities.identity.ProviderUserID != "google-subject-1" {
		t.Fatalf("expected oauth identity to be stored, got %#v", oauthIdentities.identity)
	}
	if oauthStates.state.ConsumedAt == nil {
		t.Fatal("expected oauth state to be consumed")
	}
}
