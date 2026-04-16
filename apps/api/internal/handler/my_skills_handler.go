package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/apps/api/internal/httpx"
	appmiddleware "github.com/sergehall/lavoval/apps/api/internal/middleware"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type MySkillsHandler struct {
	validate *validator.Validate
	service  *service.SkillService
}

func NewMySkillsHandler(validate *validator.Validate, service *service.SkillService) *MySkillsHandler {
	return &MySkillsHandler{validate: validate, service: service}
}

func (h *MySkillsHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())
	skills, err := h.service.ListByCreator(r.Context(), claims.UserID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "skills_load_failed", "Could not list your skills")
		return
	}
	httpx.JSON(w, http.StatusOK, skills)
}

func (h *MySkillsHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())
	skill, err := h.service.FindOwnedByCreator(r.Context(), chi.URLParam(r, "skillID"), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrSkillForbidden) {
			httpx.Error(w, http.StatusForbidden, "forbidden", "You do not have access to this skill")
			return
		}
		httpx.Error(w, http.StatusNotFound, "skill_not_found", "Skill not found")
		return
	}
	httpx.JSON(w, http.StatusOK, skill)
}

func (h *MySkillsHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())
	var input service.SkillMutationInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	skill, err := h.service.Create(r.Context(), claims.UserID, input)
	if err != nil {
		if errors.Is(err, service.ErrAccountBlocked) || errors.Is(err, service.ErrAccountSuspended) {
			httpx.Error(w, http.StatusForbidden, "skill_mutation_forbidden", "Your account cannot change skills right now")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "skill_create_failed", "Could not create your skill")
		return
	}
	httpx.JSON(w, http.StatusCreated, skill)
}

func (h *MySkillsHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())
	var input service.SkillMutationInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	skill, err := h.service.UpdateOwnedByCreator(r.Context(), chi.URLParam(r, "skillID"), claims.UserID, input)
	if err != nil {
		if errors.Is(err, service.ErrSkillForbidden) {
			httpx.Error(w, http.StatusForbidden, "forbidden", "You do not have access to this skill")
			return
		}
		if errors.Is(err, service.ErrAccountBlocked) || errors.Is(err, service.ErrAccountSuspended) {
			httpx.Error(w, http.StatusForbidden, "skill_mutation_forbidden", "Your account cannot change skills right now")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "skill_update_failed", "Could not update your skill")
		return
	}
	httpx.JSON(w, http.StatusOK, skill)
}

func (h *MySkillsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())
	if err := h.service.ArchiveOwnedByCreator(r.Context(), chi.URLParam(r, "skillID"), claims.UserID); err != nil {
		if errors.Is(err, service.ErrSkillForbidden) {
			httpx.Error(w, http.StatusForbidden, "forbidden", "You do not have access to this skill")
			return
		}
		if errors.Is(err, service.ErrAccountBlocked) || errors.Is(err, service.ErrAccountSuspended) {
			httpx.Error(w, http.StatusForbidden, "skill_mutation_forbidden", "Your account cannot change skills right now")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "skill_delete_failed", "Could not archive your skill")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"success": true})
}
