package service

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
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
var ErrMFAAlreadyEnabled = errors.New("mfa already enabled")
var ErrMFANotEnabled = errors.New("mfa not enabled")
var ErrMFAPendingEnrollmentMissing = errors.New("mfa pending enrollment missing")
var ErrMFACodeInvalid = errors.New("mfa code is invalid")
var ErrMFARecoveryCodeInvalid = errors.New("mfa recovery code is invalid")
var ErrMFASignInChallengeInvalid = errors.New("mfa sign-in challenge is invalid")
var ErrMFASignInChallengeExpired = errors.New("mfa sign-in challenge expired")
var ErrOAuthStateInvalid = errors.New("oauth state is invalid")
var ErrOAuthStateExpired = errors.New("oauth state expired")
var ErrOAuthEmailNotVerified = errors.New("oauth email not verified")
var ErrOAuthNotConfigured = errors.New("oauth is not configured")
var ErrOAuthMFASignInNotSupported = errors.New("oauth mfa sign-in not supported")

type AuthService struct {
	users            repository.UserStore
	profiles         repository.ProfileStore
	verifications    repository.EmailVerificationStore
	passwordResets   repository.PasswordResetStore
	recoveryCodes    repository.MFARecoveryCodeStore
	signInChallenges repository.SignInChallengeStore
	oauthStates      repository.OAuthStateStore
	oauthIdentities  repository.OAuthIdentityStore
	tokens           auth.TokenManager
	mailer           mailer.VerificationSender
	sessions         SessionRevoker
	cfg              config.Config
	totp             totpManager
	googleOAuth      googleOAuthProvider
	githubOAuth      githubOAuthProvider
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

type MFAStatusResponse struct {
	Enabled           bool       `json:"enabled"`
	PendingEnrollment bool       `json:"pendingEnrollment"`
	EnrolledAt        *time.Time `json:"enrolledAt,omitempty"`
	RecoveryCodes     []string   `json:"recoveryCodes,omitempty"`
}

type MFAEnrollResponse struct {
	Secret       string `json:"secret"`
	ProvisionURL string `json:"provisionUrl"`
}

type MFAVerifyEnrollmentInput struct {
	Code string `json:"code" validate:"required,len=6,numeric"`
}

type MFADisableInput struct {
	Password string `json:"password" validate:"required,min=8"`
	Code     string `json:"code" validate:"required,len=6,numeric"`
}

type MFARegenerateRecoveryCodesInput struct {
	Password string `json:"password" validate:"required,min=8"`
	Code     string `json:"code" validate:"required,len=6,numeric"`
}

type CompleteMFASignInInput struct {
	ChallengeID  string `json:"challengeId" validate:"required,uuid4"`
	Code         string `json:"code,omitempty"`
	RecoveryCode string `json:"recoveryCode,omitempty"`
}

type MFARequiredError struct {
	ChallengeID string
}

func (e *MFARequiredError) Error() string {
	return "mfa required"
}

func NewAuthService(
	users repository.UserStore,
	profiles repository.ProfileStore,
	verifications repository.EmailVerificationStore,
	passwordResets repository.PasswordResetStore,
	recoveryCodes repository.MFARecoveryCodeStore,
	signInChallenges repository.SignInChallengeStore,
	oauthStates repository.OAuthStateStore,
	oauthIdentities repository.OAuthIdentityStore,
	tokens auth.TokenManager,
	verificationMailer mailer.VerificationSender,
	sessionRevoker SessionRevoker,
	cfg config.Config,
) *AuthService {
	if sessionRevoker == nil {
		sessionRevoker = NoopSessionRevoker{}
	}
	if recoveryCodes == nil {
		recoveryCodes = noopMFARecoveryCodeStore{}
	}
	if signInChallenges == nil {
		signInChallenges = noopSignInChallengeStore{}
	}
	if oauthStates == nil {
		oauthStates = noopOAuthStateStore{}
	}
	if oauthIdentities == nil {
		oauthIdentities = noopOAuthIdentityStore{}
	}

	return &AuthService{
		users:            users,
		profiles:         profiles,
		verifications:    verifications,
		passwordResets:   passwordResets,
		recoveryCodes:    recoveryCodes,
		signInChallenges: signInChallenges,
		oauthStates:      oauthStates,
		oauthIdentities:  oauthIdentities,
		tokens:           tokens,
		mailer:           verificationMailer,
		sessions:         sessionRevoker,
		cfg:              cfg,
		totp:             newTOTPManager(cfg),
		googleOAuth:      newGoogleOAuthProvider(cfg),
		githubOAuth:      newGitHubOAuthProvider(cfg),
	}
}

type noopMFARecoveryCodeStore struct{}

func (noopMFARecoveryCodeStore) ReplaceForUser(context.Context, string, []domain.MFARecoveryCode) error {
	return nil
}
func (noopMFARecoveryCodeStore) FindActiveByCodeHash(context.Context, string, string) (domain.MFARecoveryCode, error) {
	return domain.MFARecoveryCode{}, pgx.ErrNoRows
}
func (noopMFARecoveryCodeStore) Consume(context.Context, string, string) error {
	return nil
}
func (noopMFARecoveryCodeStore) RevokeActiveByUserID(context.Context, string) error {
	return nil
}

type noopSignInChallengeStore struct{}

func (noopSignInChallengeStore) Create(_ context.Context, challenge domain.AuthSignInChallenge) (domain.AuthSignInChallenge, error) {
	return challenge, nil
}
func (noopSignInChallengeStore) FindByID(context.Context, string) (domain.AuthSignInChallenge, error) {
	return domain.AuthSignInChallenge{}, pgx.ErrNoRows
}
func (noopSignInChallengeStore) Consume(context.Context, string, string) error {
	return nil
}
func (noopSignInChallengeStore) RevokeActiveByUserID(context.Context, string) error {
	return nil
}

type noopOAuthStateStore struct{}

func (noopOAuthStateStore) Create(_ context.Context, state domain.OAuthState) (domain.OAuthState, error) {
	return state, nil
}
func (noopOAuthStateStore) FindByStateHash(context.Context, domain.OAuthProvider, string) (domain.OAuthState, error) {
	return domain.OAuthState{}, pgx.ErrNoRows
}
func (noopOAuthStateStore) Consume(context.Context, string) error {
	return nil
}

type noopOAuthIdentityStore struct{}

func (noopOAuthIdentityStore) Create(_ context.Context, identity domain.OAuthIdentity) (domain.OAuthIdentity, error) {
	return identity, nil
}
func (noopOAuthIdentityStore) FindByProviderSubject(context.Context, domain.OAuthProvider, string) (domain.OAuthIdentity, error) {
	return domain.OAuthIdentity{}, pgx.ErrNoRows
}
func (noopOAuthIdentityStore) ListByUserID(context.Context, string) ([]domain.OAuthIdentity, error) {
	return nil, nil
}

func (s *AuthService) MFAStatus(ctx context.Context, userID string) (MFAStatusResponse, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return MFAStatusResponse{}, fmt.Errorf("find user for mfa status: %w", err)
	}

	return MFAStatusResponse{
		Enabled:           user.MFAEnabled,
		PendingEnrollment: user.MFAPendingTOTPSecretEncrypted != nil && *user.MFAPendingTOTPSecretEncrypted != "",
		EnrolledAt:        user.MFAEnrolledAt,
	}, nil
}

func (s *AuthService) EnrollMFA(ctx context.Context, userID string) (MFAEnrollResponse, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return MFAEnrollResponse{}, fmt.Errorf("find user for mfa enrollment: %w", err)
	}
	if user.MFAEnabled {
		return MFAEnrollResponse{}, ErrMFAAlreadyEnabled
	}

	secret, err := s.totp.GenerateSecret()
	if err != nil {
		return MFAEnrollResponse{}, err
	}

	encryptedSecret, err := s.totp.EncryptSecret(secret)
	if err != nil {
		return MFAEnrollResponse{}, err
	}

	if _, err := s.users.StartTOTPEnrollment(ctx, user.ID, encryptedSecret); err != nil {
		return MFAEnrollResponse{}, fmt.Errorf("start mfa enrollment: %w", err)
	}

	return MFAEnrollResponse{
		Secret:       secret,
		ProvisionURL: s.totp.ProvisioningURI(user.Email, secret),
	}, nil
}

func (s *AuthService) CancelMFAEnrollment(ctx context.Context, userID string) (MFAStatusResponse, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return MFAStatusResponse{}, fmt.Errorf("find user for mfa cancel: %w", err)
	}
	if user.MFAEnabled {
		return MFAStatusResponse{}, ErrMFAAlreadyEnabled
	}
	if user.MFAPendingTOTPSecretEncrypted == nil || *user.MFAPendingTOTPSecretEncrypted == "" {
		return MFAStatusResponse{}, ErrMFAPendingEnrollmentMissing
	}

	updatedUser, err := s.users.CancelTOTPEnrollment(ctx, user.ID)
	if err != nil {
		return MFAStatusResponse{}, fmt.Errorf("cancel mfa enrollment: %w", err)
	}

	return MFAStatusResponse{
		Enabled:           updatedUser.MFAEnabled,
		PendingEnrollment: updatedUser.MFAPendingTOTPSecretEncrypted != nil && *updatedUser.MFAPendingTOTPSecretEncrypted != "",
		EnrolledAt:        updatedUser.MFAEnrolledAt,
	}, nil
}

func (s *AuthService) VerifyMFAEnrollment(ctx context.Context, userID string, input MFAVerifyEnrollmentInput) (MFAStatusResponse, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return MFAStatusResponse{}, fmt.Errorf("find user for mfa verification: %w", err)
	}
	if user.MFAEnabled {
		return MFAStatusResponse{}, ErrMFAAlreadyEnabled
	}
	if user.MFAPendingTOTPSecretEncrypted == nil || *user.MFAPendingTOTPSecretEncrypted == "" {
		return MFAStatusResponse{}, ErrMFAPendingEnrollmentMissing
	}

	secret, err := s.totp.DecryptSecret(*user.MFAPendingTOTPSecretEncrypted)
	if err != nil {
		return MFAStatusResponse{}, fmt.Errorf("decrypt pending mfa secret: %w", err)
	}

	if !s.totp.VerifyCode(secret, input.Code, time.Now()) {
		return MFAStatusResponse{}, ErrMFACodeInvalid
	}

	updatedUser, err := s.users.EnableTOTP(ctx, user.ID, *user.MFAPendingTOTPSecretEncrypted)
	if err != nil {
		return MFAStatusResponse{}, fmt.Errorf("enable mfa: %w", err)
	}

	recoveryCodes, recoveryRecords, err := s.generateRecoveryCodes(user.ID)
	if err != nil {
		return MFAStatusResponse{}, fmt.Errorf("generate recovery codes: %w", err)
	}
	if err := s.recoveryCodes.ReplaceForUser(ctx, user.ID, recoveryRecords); err != nil {
		return MFAStatusResponse{}, fmt.Errorf("store recovery codes: %w", err)
	}

	return MFAStatusResponse{
		Enabled:           updatedUser.MFAEnabled,
		PendingEnrollment: false,
		EnrolledAt:        updatedUser.MFAEnrolledAt,
		RecoveryCodes:     recoveryCodes,
	}, nil
}

func (s *AuthService) DisableMFA(ctx context.Context, userID string, input MFADisableInput) (MFAStatusResponse, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return MFAStatusResponse{}, fmt.Errorf("find user for mfa disable: %w", err)
	}
	if !user.MFAEnabled || user.MFATOTPSecretEncrypted == nil || *user.MFATOTPSecretEncrypted == "" {
		return MFAStatusResponse{}, ErrMFANotEnabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return MFAStatusResponse{}, ErrInvalidCredentials
	}

	secret, err := s.totp.DecryptSecret(*user.MFATOTPSecretEncrypted)
	if err != nil {
		return MFAStatusResponse{}, fmt.Errorf("decrypt active mfa secret: %w", err)
	}
	if !s.totp.VerifyCode(secret, input.Code, time.Now()) {
		return MFAStatusResponse{}, ErrMFACodeInvalid
	}

	updatedUser, err := s.users.DisableTOTP(ctx, user.ID)
	if err != nil {
		return MFAStatusResponse{}, fmt.Errorf("disable mfa: %w", err)
	}
	if err := s.recoveryCodes.RevokeActiveByUserID(ctx, user.ID); err != nil {
		return MFAStatusResponse{}, fmt.Errorf("revoke recovery codes after mfa disable: %w", err)
	}
	if err := s.sessions.RevokeAllSessionsForUser(ctx, user.ID, "mfa_disabled"); err != nil {
		return MFAStatusResponse{}, fmt.Errorf("revoke sessions after mfa disable: %w", err)
	}

	return MFAStatusResponse{
		Enabled:           updatedUser.MFAEnabled,
		PendingEnrollment: false,
		EnrolledAt:        updatedUser.MFAEnrolledAt,
	}, nil
}

func (s *AuthService) RegenerateMFARecoveryCodes(ctx context.Context, userID string, input MFARegenerateRecoveryCodesInput) (MFAStatusResponse, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return MFAStatusResponse{}, fmt.Errorf("find user for recovery code regeneration: %w", err)
	}
	if !user.MFAEnabled || user.MFATOTPSecretEncrypted == nil || *user.MFATOTPSecretEncrypted == "" {
		return MFAStatusResponse{}, ErrMFANotEnabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return MFAStatusResponse{}, ErrInvalidCredentials
	}

	secret, err := s.totp.DecryptSecret(*user.MFATOTPSecretEncrypted)
	if err != nil {
		return MFAStatusResponse{}, fmt.Errorf("decrypt active mfa secret: %w", err)
	}
	if !s.totp.VerifyCode(secret, input.Code, time.Now()) {
		return MFAStatusResponse{}, ErrMFACodeInvalid
	}

	recoveryCodes, recoveryRecords, err := s.generateRecoveryCodes(user.ID)
	if err != nil {
		return MFAStatusResponse{}, fmt.Errorf("generate recovery codes: %w", err)
	}
	if err := s.recoveryCodes.ReplaceForUser(ctx, user.ID, recoveryRecords); err != nil {
		return MFAStatusResponse{}, fmt.Errorf("replace recovery codes: %w", err)
	}

	return MFAStatusResponse{
		Enabled:           user.MFAEnabled,
		PendingEnrollment: false,
		EnrolledAt:        user.MFAEnrolledAt,
		RecoveryCodes:     recoveryCodes,
	}, nil
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
		if cleanupErr := s.users.SoftDelete(ctx, createdUser.ID); cleanupErr != nil {
			log.Printf("auth: cleanup failed after profile create error for user %s: %v", createdUser.ID, cleanupErr)
		}
		return RegisterResponse{}, fmt.Errorf("create profile: %w", err)
	}

	go func() {
		if err := s.issueVerificationEmail(context.Background(), createdUser, createdProfile); err != nil {
			log.Printf("auth: verification email failed for %s: %v", createdUser.Email, err)
		}
	}()

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
	if user.MFAEnabled {
		challenge, err := s.startMFASignInChallenge(ctx, user.ID)
		if err != nil {
			return AuthPayload{}, fmt.Errorf("start mfa sign-in challenge: %w", err)
		}
		return AuthPayload{}, &MFARequiredError{ChallengeID: challenge.ID}
	}

	profile, err := s.profiles.FindByUserID(ctx, user.ID)
	if err != nil {
		return AuthPayload{}, fmt.Errorf("find profile: %w", err)
	}

	return s.buildAuthPayload(user, profile)
}

func (s *AuthService) CompleteMFASignIn(ctx context.Context, input CompleteMFASignInInput) (AuthPayload, error) {
	challenge, err := s.signInChallenges.FindByID(ctx, input.ChallengeID)
	if err != nil {
		return AuthPayload{}, ErrMFASignInChallengeInvalid
	}
	if challenge.ConsumedAt != nil {
		return AuthPayload{}, ErrMFASignInChallengeInvalid
	}
	if time.Now().After(challenge.ExpiresAt) {
		return AuthPayload{}, ErrMFASignInChallengeExpired
	}

	user, err := s.users.FindByID(ctx, challenge.UserID)
	if err != nil {
		return AuthPayload{}, fmt.Errorf("find challenge user: %w", err)
	}
	if !user.MFAEnabled || user.MFATOTPSecretEncrypted == nil || *user.MFATOTPSecretEncrypted == "" {
		return AuthPayload{}, ErrMFANotEnabled
	}

	switch {
	case strings.TrimSpace(input.RecoveryCode) != "":
		if err := s.consumeRecoveryCode(ctx, user.ID, input.RecoveryCode); err != nil {
			return AuthPayload{}, err
		}
	case strings.TrimSpace(input.Code) != "":
		secret, err := s.totp.DecryptSecret(*user.MFATOTPSecretEncrypted)
		if err != nil {
			return AuthPayload{}, fmt.Errorf("decrypt active mfa secret: %w", err)
		}
		if !s.totp.VerifyCode(secret, input.Code, time.Now()) {
			return AuthPayload{}, ErrMFACodeInvalid
		}
	default:
		return AuthPayload{}, ErrMFACodeInvalid
	}

	if err := s.signInChallenges.Consume(ctx, challenge.ID, user.ID); err != nil {
		return AuthPayload{}, fmt.Errorf("consume sign-in challenge: %w", err)
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

	go func() {
		if err := s.issueVerificationEmail(context.Background(), user, profile); err != nil {
			log.Printf("auth: resend verification email failed for %s: %v", user.Email, err)
		}
	}()

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

func (s *AuthService) startMFASignInChallenge(ctx context.Context, userID string) (domain.AuthSignInChallenge, error) {
	if err := s.signInChallenges.RevokeActiveByUserID(ctx, userID); err != nil {
		return domain.AuthSignInChallenge{}, fmt.Errorf("revoke active mfa challenges: %w", err)
	}

	challenge := domain.AuthSignInChallenge{
		ID:        uuid.NewString(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.cfg.MFASignInChallengeTTL),
	}

	created, err := s.signInChallenges.Create(ctx, challenge)
	if err != nil {
		return domain.AuthSignInChallenge{}, fmt.Errorf("create mfa challenge: %w", err)
	}

	return created, nil
}

func (s *AuthService) generateRecoveryCodes(userID string) ([]string, []domain.MFARecoveryCode, error) {
	plain := make([]string, 0, 8)
	records := make([]domain.MFARecoveryCode, 0, 8)

	for i := 0; i < 8; i++ {
		code, err := generateRecoveryCode()
		if err != nil {
			return nil, nil, err
		}

		plain = append(plain, code)
		records = append(records, domain.MFARecoveryCode{
			ID:       uuid.NewString(),
			UserID:   userID,
			CodeHash: hashRecoveryCode(code),
		})
	}

	return plain, records, nil
}

func (s *AuthService) consumeRecoveryCode(ctx context.Context, userID string, plain string) error {
	record, err := s.recoveryCodes.FindActiveByCodeHash(ctx, userID, hashRecoveryCode(plain))
	if err != nil {
		return ErrMFARecoveryCodeInvalid
	}
	if err := s.recoveryCodes.Consume(ctx, record.ID, userID); err != nil {
		return fmt.Errorf("consume recovery code: %w", err)
	}
	return nil
}

func generateRecoveryCode() (string, error) {
	plain, _, err := newVerificationToken()
	if err != nil {
		return "", fmt.Errorf("generate recovery code: %w", err)
	}

	normalized := strings.ToUpper(strings.TrimRight(base32.StdEncoding.EncodeToString([]byte(plain)), "="))
	if len(normalized) < 12 {
		return normalized, nil
	}

	return normalized[:4] + "-" + normalized[4:8] + "-" + normalized[8:12], nil
}

func hashRecoveryCode(plain string) string {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(plain), "-", ""))
	sum := sha256.Sum256([]byte(normalized))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
