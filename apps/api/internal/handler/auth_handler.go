package handler

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/apps/api/internal/httpx"
	appmiddleware "github.com/sergehall/lavoval/apps/api/internal/middleware"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type AuthHandler struct {
	validate *validator.Validate
	service  *service.AuthService
}

func NewAuthHandler(validate *validator.Validate, service *service.AuthService) *AuthHandler {
	return &AuthHandler{validate: validate, service: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	payload, err := h.service.Register(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			httpx.Error(w, http.StatusConflict, "email_in_use", "An account with this email already exists")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "register_failed", "Could not register account")
		return
	}

	httpx.JSON(w, http.StatusCreated, payload)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	payload, err := h.service.Login(r.Context(), input)
	if err != nil {
		var mfaRequired *service.MFARequiredError
		if errors.As(err, &mfaRequired) {
			httpx.ErrorWithMeta(w, http.StatusUnauthorized, "mfa_required", "Multi-factor authentication is required to finish signing in", map[string]any{
				"challengeId": mfaRequired.ChallengeID,
			})
			return
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			httpx.Error(w, http.StatusUnauthorized, "invalid_credentials", "Email or password is incorrect")
			return
		}
		if errors.Is(err, service.ErrEmailNotVerified) {
			httpx.Error(w, http.StatusForbidden, "email_not_verified", "Please confirm your email before signing in")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "login_failed", "Could not sign in")
		return
	}

	httpx.JSON(w, http.StatusOK, payload)
}

func (h *AuthHandler) CompleteMFASignIn(w http.ResponseWriter, r *http.Request) {
	var input service.CompleteMFASignInInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.StructPartial(input, "ChallengeID"); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	payload, err := h.service.CompleteMFASignIn(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFASignInChallengeExpired):
			httpx.Error(w, http.StatusGone, "mfa_challenge_expired", "This sign-in challenge expired. Sign in again to get a fresh code prompt.")
		case errors.Is(err, service.ErrMFASignInChallengeInvalid):
			httpx.Error(w, http.StatusBadRequest, "mfa_challenge_invalid", "This sign-in challenge is invalid.")
		case errors.Is(err, service.ErrMFARecoveryCodeInvalid):
			httpx.Error(w, http.StatusBadRequest, "mfa_recovery_code_invalid", "That recovery code is invalid.")
		case errors.Is(err, service.ErrMFACodeInvalid):
			httpx.Error(w, http.StatusBadRequest, "mfa_code_invalid", "The authenticator code is invalid.")
		default:
			httpx.Error(w, http.StatusInternalServerError, "mfa_complete_sign_in_failed", "Could not complete sign in.")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, payload)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var input service.VerifyEmailInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	payload, err := h.service.VerifyEmail(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrVerificationTokenExpired):
			httpx.Error(w, http.StatusGone, "verification_token_expired", "This confirmation link has expired")
		case errors.Is(err, service.ErrVerificationTokenInvalid):
			httpx.Error(w, http.StatusBadRequest, "verification_token_invalid", "This confirmation link is invalid")
		default:
			httpx.Error(w, http.StatusInternalServerError, "verify_email_failed", "Could not confirm email")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, payload)
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var input service.ResendVerificationInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	payload, err := h.service.ResendVerification(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			httpx.Error(w, http.StatusNotFound, "email_not_found", "No account found for this email")
		case errors.Is(err, service.ErrEmailAlreadyVerified):
			httpx.Error(w, http.StatusConflict, "email_already_verified", "This email is already confirmed")
		default:
			httpx.Error(w, http.StatusInternalServerError, "resend_verification_failed", "Could not resend confirmation email")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, payload)
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var input service.ForgotPasswordInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	payload, err := h.service.ForgotPassword(r.Context(), input)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "forgot_password_failed", "Could not start password recovery")
		return
	}

	httpx.JSON(w, http.StatusOK, payload)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var input service.ResetPasswordInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	payload, err := h.service.ResetPassword(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPasswordResetTokenExpired):
			httpx.Error(w, http.StatusGone, "password_reset_token_expired", "This password reset link has expired")
		case errors.Is(err, service.ErrPasswordResetTokenInvalid):
			httpx.Error(w, http.StatusBadRequest, "password_reset_token_invalid", "This password reset link is invalid")
		default:
			httpx.Error(w, http.StatusInternalServerError, "reset_password_failed", "Could not reset password")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, payload)
}

func (h *AuthHandler) MFAStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "Missing authenticated session")
		return
	}

	payload, err := h.service.MFAStatus(r.Context(), claims.UserID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "mfa_status_failed", "Could not load MFA status")
		return
	}

	httpx.JSON(w, http.StatusOK, payload)
}

func (h *AuthHandler) EnrollMFA(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "Missing authenticated session")
		return
	}

	payload, err := h.service.EnrollMFA(r.Context(), claims.UserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFAAlreadyEnabled):
			httpx.Error(w, http.StatusConflict, "mfa_already_enabled", "MFA is already enabled for this account")
		default:
			httpx.Error(w, http.StatusInternalServerError, "mfa_enroll_failed", "Could not start MFA enrollment")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, payload)
}

func (h *AuthHandler) VerifyMFAEnrollment(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "Missing authenticated session")
		return
	}

	var input service.MFAVerifyEnrollmentInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	payload, err := h.service.VerifyMFAEnrollment(r.Context(), claims.UserID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFAAlreadyEnabled):
			httpx.Error(w, http.StatusConflict, "mfa_already_enabled", "MFA is already enabled for this account")
		case errors.Is(err, service.ErrMFAPendingEnrollmentMissing):
			httpx.Error(w, http.StatusConflict, "mfa_enrollment_missing", "Start MFA enrollment before verifying it")
		case errors.Is(err, service.ErrMFACodeInvalid):
			httpx.Error(w, http.StatusBadRequest, "mfa_code_invalid", "The authenticator code is invalid")
		default:
			httpx.Error(w, http.StatusInternalServerError, "mfa_verify_failed", "Could not verify MFA enrollment")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, payload)
}

func (h *AuthHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "Missing authenticated session")
		return
	}

	var input service.MFADisableInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	payload, err := h.service.DisableMFA(r.Context(), claims.UserID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFANotEnabled):
			httpx.Error(w, http.StatusConflict, "mfa_not_enabled", "MFA is not enabled for this account")
		case errors.Is(err, service.ErrInvalidCredentials):
			httpx.Error(w, http.StatusUnauthorized, "invalid_credentials", "Current password is incorrect")
		case errors.Is(err, service.ErrMFACodeInvalid):
			httpx.Error(w, http.StatusBadRequest, "mfa_code_invalid", "The authenticator code is invalid")
		default:
			httpx.Error(w, http.StatusInternalServerError, "mfa_disable_failed", "Could not disable MFA")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, payload)
}

func (h *AuthHandler) RegenerateMFARecoveryCodes(w http.ResponseWriter, r *http.Request) {
	claims, ok := appmiddleware.ClaimsFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "Missing authenticated session")
		return
	}

	var input service.MFARegenerateRecoveryCodesInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	payload, err := h.service.RegenerateMFARecoveryCodes(r.Context(), claims.UserID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFANotEnabled):
			httpx.Error(w, http.StatusConflict, "mfa_not_enabled", "MFA is not enabled for this account")
		case errors.Is(err, service.ErrInvalidCredentials):
			httpx.Error(w, http.StatusUnauthorized, "invalid_credentials", "Current password is incorrect")
		case errors.Is(err, service.ErrMFACodeInvalid):
			httpx.Error(w, http.StatusBadRequest, "mfa_code_invalid", "The authenticator code is invalid")
		default:
			httpx.Error(w, http.StatusInternalServerError, "mfa_recovery_codes_regenerate_failed", "Could not regenerate recovery codes")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, payload)
}
