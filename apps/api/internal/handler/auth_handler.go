package handler

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/apps/api/internal/httpx"
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
		if errors.Is(err, service.ErrInvalidCredentials) {
			httpx.Error(w, http.StatusUnauthorized, "invalid_credentials", "Email or password is incorrect")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "login_failed", "Could not sign in")
		return
	}

	httpx.JSON(w, http.StatusOK, payload)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]bool{"success": true})
}
