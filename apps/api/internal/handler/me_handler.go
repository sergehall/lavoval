package handler

import (
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/apps/api/internal/httpx"
	appmiddleware "github.com/sergehall/lavoval/apps/api/internal/middleware"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type MeHandler struct {
	validate *validator.Validate
	service  *service.ProfileService
	security *service.AccountSecurityService
}

func NewMeHandler(
	validate *validator.Validate,
	service *service.ProfileService,
	security *service.AccountSecurityService,
) *MeHandler {
	return &MeHandler{validate: validate, service: service, security: security}
}

func (h *MeHandler) Profile(w http.ResponseWriter, r *http.Request) {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())
	profile, err := h.service.FindByUserID(r.Context(), claims.UserID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "profile_load_failed", "Could not load profile")
		return
	}
	httpx.JSON(w, http.StatusOK, profile)
}

func (h *MeHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())
	var input service.UpdateProfileInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	profile, err := h.service.Update(r.Context(), claims.UserID, input)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "profile_update_failed", "Could not update profile")
		return
	}
	httpx.JSON(w, http.StatusOK, profile)
}

func (h *MeHandler) Security(w http.ResponseWriter, r *http.Request) {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())
	summary, err := h.security.Summary(r.Context(), claims.UserID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "security_load_failed", "Could not load security settings")
		return
	}
	httpx.JSON(w, http.StatusOK, summary)
}
