package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/sergehall/lavoval/apps/api/internal/auth"
	"github.com/sergehall/lavoval/apps/api/internal/config"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/mailer"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrEmailAlreadyExists = errors.New("email already exists")
var ErrEmailNotVerified = errors.New("email not verified")
var ErrVerificationTokenInvalid = errors.New("verification token is invalid")
var ErrVerificationTokenExpired = errors.New("verification token expired")
var ErrEmailAlreadyVerified = errors.New("email already verified")
var ErrPasswordResetTokenInvalid = errors.New("password reset token is invalid")
var ErrPasswordResetTokenExpired = errors.New("password reset token expired")

type AuthService struct {
	users          repository.UserStore
	profiles       repository.ProfileStore
	verifications  repository.EmailVerificationStore
	passwordResets repository.PasswordResetStore
	tokens         auth.TokenManager
	mailer         mailer.VerificationSender
	sessions       SessionRevoker
	cfg            config.Config
}

type AuthPayload struct {
	AccessToken  string      `json:"accessToken"`
	RefreshToken string      `json:"refreshToken"`
	User         SessionUser `json:"user"`
}

type RegisterResponse struct {
	Email                string `json:"email"`
	VerificationRequired bool   `json:"verificationRequired"`
}

type VerificationResponse struct {
	Email           string `json:"email"`
	AlreadyVerified bool   `json:"alreadyVerified"`
}

type SessionUser struct {
	ID        string      `json:"id"`
	Email     string      `json:"email"`
	Role      domain.Role `json:"role"`
	FirstName string      `json:"firstName,omitempty"`
	LastName  string      `json:"lastName,omitempty"`
}

type RegisterInput struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=12"`
	FirstName string `json:"firstName" validate:"required,min=2"`
	LastName  string `json:"lastName" validate:"required,min=2"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type VerifyEmailInput struct {
	Token string `json:"token" validate:"required,min=24"`
}

type ResendVerificationInput struct {
	Email string `json:"email" validate:"required,email"`
}

type ForgotPasswordInput struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordInput struct {
	Token       string `json:"token" validate:"required,min=24"`
	NewPassword string `json:"newPassword" validate:"required,min=12"`
}

type ForgotPasswordResponse struct {
	Email string `json:"email"`
	Sent  bool   `json:"sent"`
}

type ResetPasswordResponse struct {
	Email string `json:"email"`
	Reset bool   `json:"reset"`
}

func NewAuthService(
	users repository.UserStore,
	profiles repository.ProfileStore,
	verifications repository.EmailVerificationStore,
	passwordResets repository.PasswordResetStore,
	tokens auth.TokenManager,
	verificationMailer mailer.VerificationSender,
	sessionRevoker SessionRevoker,
	cfg config.Config,
) *AuthService {
	if sessionRevoker == nil {
		sessionRevoker = NoopSessionRevoker{}
	}

	return &AuthService{
		users:          users,
		profiles:       profiles,
		verifications:  verifications,
		passwordResets: passwordResets,
		tokens:         tokens,
		mailer:         verificationMailer,
		sessions:       sessionRevoker,
		cfg:            cfg,
	}
}

func (s *AuthService) ForgotPassword(ctx context.Context, input ForgotPasswordInput) (ForgotPasswordResponse, error) {
	user, err := s.users.FindByEmail(ctx, input.Email)
	if err != nil {
		return ForgotPasswordResponse{
			Email: input.Email,
			Sent:  true,
		}, nil
	}

	profile, err := s.profiles.FindByUserID(ctx, user.ID)
	if err != nil {
		return ForgotPasswordResponse{}, fmt.Errorf("find profile for password reset: %w", err)
	}

	if err := s.issuePasswordResetEmail(ctx, user, profile); err != nil {
		return ForgotPasswordResponse{}, err
	}

	return ForgotPasswordResponse{
		Email: user.Email,
		Sent:  true,
	}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, input ResetPasswordInput) (ResetPasswordResponse, error) {
	tokenHash := hashVerificationToken(input.Token)
	record, err := s.passwordResets.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return ResetPasswordResponse{}, ErrPasswordResetTokenInvalid
	}
	if record.ConsumedAt != nil {
		return ResetPasswordResponse{}, ErrPasswordResetTokenInvalid
	}
	if time.Now().After(record.ExpiresAt) {
		return ResetPasswordResponse{}, ErrPasswordResetTokenExpired
	}

	user, err := s.users.FindByID(ctx, record.UserID)
	if err != nil {
		return ResetPasswordResponse{}, fmt.Errorf("find user for password reset: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return ResetPasswordResponse{}, fmt.Errorf("hash new password: %w", err)
	}

	if _, err := s.users.UpdatePasswordHash(ctx, user.ID, string(hash)); err != nil {
		return ResetPasswordResponse{}, fmt.Errorf("update user password: %w", err)
	}
	if err := s.passwordResets.Consume(ctx, record.ID, user.ID); err != nil {
		return ResetPasswordResponse{}, fmt.Errorf("consume password reset token: %w", err)
	}
	if err := s.passwordResets.RevokeActiveByUserID(ctx, user.ID); err != nil {
		return ResetPasswordResponse{}, fmt.Errorf("revoke active password reset tokens: %w", err)
	}
	if err := s.sessions.RevokeAllSessionsForUser(ctx, user.ID, "password_reset"); err != nil {
		return ResetPasswordResponse{}, fmt.Errorf("revoke sessions after password reset: %w", err)
	}
	if err := s.mailer.SendPasswordChangedEmail(ctx, mailer.PasswordChangedEmail{
		ToEmail:     user.Email,
		ToName:      user.Email,
		ProductName: s.cfg.AppName,
		SignInURL:   fmt.Sprintf("%s/?auth=sign-in", s.cfg.AppURL),
	}); err != nil {
		return ResetPasswordResponse{}, fmt.Errorf("send password changed email: %w", err)
	}

	return ResetPasswordResponse{
		Email: user.Email,
		Reset: true,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (RegisterResponse, error) {
	existingUser, err := s.users.FindByEmail(ctx, input.Email)
	if err == nil && existingUser.ID != "" {
		return RegisterResponse{}, ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return RegisterResponse{}, fmt.Errorf("check existing user: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{
		ID:           uuid.NewString(),
		Email:        input.Email,
		PasswordHash: string(hash),
		Role:         domain.RoleUser,
		Status:       domain.AccountStatusActive,
	}

	createdUser, err := s.users.Create(ctx, user)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("create user: %w", err)
	}

	profile := domain.Profile{
		UserID:    createdUser.ID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Timezone:  "UTC",
	}

	createdProfile, err := s.profiles.Create(ctx, profile)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("create profile: %w", err)
	}

	if err := s.issueVerificationEmail(ctx, createdUser, createdProfile); err != nil {
		return RegisterResponse{}, err
	}

	return RegisterResponse{
		Email:                createdUser.Email,
		VerificationRequired: true,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (AuthPayload, error) {
	user, err := s.users.FindByEmail(ctx, input.Email)
	if err != nil {
		return AuthPayload{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return AuthPayload{}, ErrInvalidCredentials
	}
	if user.EmailVerifiedAt == nil {
		return AuthPayload{}, ErrEmailNotVerified
	}

	profile, err := s.profiles.FindByUserID(ctx, user.ID)
	if err != nil {
		return AuthPayload{}, fmt.Errorf("find profile: %w", err)
	}

	return s.buildAuthPayload(user, profile)
}

func (s *AuthService) VerifyEmail(ctx context.Context, input VerifyEmailInput) (VerificationResponse, error) {
	tokenHash := hashVerificationToken(input.Token)
	record, err := s.verifications.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return VerificationResponse{}, ErrVerificationTokenInvalid
	}
	if record.ConsumedAt != nil {
		return VerificationResponse{}, ErrVerificationTokenInvalid
	}
	if time.Now().After(record.ExpiresAt) {
		return VerificationResponse{}, ErrVerificationTokenExpired
	}

	user, err := s.users.FindByID(ctx, record.UserID)
	if err != nil {
		return VerificationResponse{}, fmt.Errorf("find user for verification: %w", err)
	}
	if user.EmailVerifiedAt != nil {
		return VerificationResponse{
			Email:           user.Email,
			AlreadyVerified: true,
		}, nil
	}

	if _, err := s.users.MarkEmailVerified(ctx, user.ID); err != nil {
		return VerificationResponse{}, fmt.Errorf("mark email verified: %w", err)
	}
	if err := s.verifications.Consume(ctx, record.ID, user.ID); err != nil {
		return VerificationResponse{}, fmt.Errorf("consume verification token: %w", err)
	}

	return VerificationResponse{
		Email:           user.Email,
		AlreadyVerified: false,
	}, nil
}

func (s *AuthService) ResendVerification(ctx context.Context, input ResendVerificationInput) (RegisterResponse, error) {
	user, err := s.users.FindByEmail(ctx, input.Email)
	if err != nil {
		return RegisterResponse{}, ErrInvalidCredentials
	}
	if user.EmailVerifiedAt != nil {
		return RegisterResponse{}, ErrEmailAlreadyVerified
	}

	profile, err := s.profiles.FindByUserID(ctx, user.ID)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("find profile for verification resend: %w", err)
	}

	if err := s.issueVerificationEmail(ctx, user, profile); err != nil {
		return RegisterResponse{}, err
	}

	return RegisterResponse{
		Email:                user.Email,
		VerificationRequired: true,
	}, nil
}

func (s *AuthService) buildAuthPayload(user domain.User, profile domain.Profile) (AuthPayload, error) {
	pair, err := s.tokens.IssueTokens(user)
	if err != nil {
		return AuthPayload{}, fmt.Errorf("issue tokens: %w", err)
	}

	return AuthPayload{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User: SessionUser{
			ID:        user.ID,
			Email:     user.Email,
			Role:      user.Role,
			FirstName: profile.FirstName,
			LastName:  profile.LastName,
		},
	}, nil
}

func (s *AuthService) issueVerificationEmail(ctx context.Context, user domain.User, profile domain.Profile) error {
	if s.mailer == nil {
		return fmt.Errorf("email delivery is not configured")
	}

	if err := s.verifications.RevokeActiveByUserID(ctx, user.ID); err != nil {
		return fmt.Errorf("revoke active verification tokens: %w", err)
	}

	plainToken, tokenHash, err := newVerificationToken()
	if err != nil {
		return fmt.Errorf("generate verification token: %w", err)
	}

	record := domain.EmailVerificationToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.cfg.EmailVerificationTTL),
	}

	if _, err := s.verifications.Create(ctx, record); err != nil {
		return fmt.Errorf("store verification token: %w", err)
	}

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s&email=%s", s.cfg.AppURL, url.QueryEscape(plainToken), url.QueryEscape(user.Email))
	if err := s.mailer.SendVerificationEmail(ctx, mailer.VerificationEmail{
		ToEmail:     user.Email,
		ToName:      strings.TrimSpace(profile.FirstName + " " + profile.LastName),
		VerifyURL:   verifyURL,
		ProductName: s.cfg.AppName,
	}); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}

func (s *AuthService) issuePasswordResetEmail(ctx context.Context, user domain.User, profile domain.Profile) error {
	if s.mailer == nil {
		return fmt.Errorf("email delivery is not configured")
	}

	if err := s.passwordResets.RevokeActiveByUserID(ctx, user.ID); err != nil {
		return fmt.Errorf("revoke active password reset tokens: %w", err)
	}

	plainToken, tokenHash, err := newVerificationToken()
	if err != nil {
		return fmt.Errorf("generate password reset token: %w", err)
	}

	record := domain.PasswordResetToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.cfg.PasswordResetTTL),
	}

	if _, err := s.passwordResets.Create(ctx, record); err != nil {
		return fmt.Errorf("store password reset token: %w", err)
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s&email=%s", s.cfg.AppURL, url.QueryEscape(plainToken), url.QueryEscape(user.Email))
	if err := s.mailer.SendPasswordResetEmail(ctx, mailer.PasswordResetEmail{
		ToEmail:     user.Email,
		ToName:      strings.TrimSpace(profile.FirstName + " " + profile.LastName),
		ResetURL:    resetURL,
		ProductName: s.cfg.AppName,
	}); err != nil {
		return fmt.Errorf("send password reset email: %w", err)
	}

	return nil
}

func newVerificationToken() (plain string, hashed string, err error) {
	raw := uuid.NewString() + "." + uuid.NewString()
	return raw, hashVerificationToken(raw), nil
}

func hashVerificationToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
