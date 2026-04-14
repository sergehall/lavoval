package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/backend/api/internal/httpx"
	appmiddleware "github.com/sergehall/lavoval/backend/api/internal/middleware"
	"github.com/sergehall/lavoval/backend/api/internal/service"
)

type AdminHandler struct {
	validate     *validator.Validate
	adminService *service.AdminService
	skillService *service.SkillService
}

func NewAdminHandler(validate *validator.Validate, adminService *service.AdminService, skillService *service.SkillService) *AdminHandler {
	return &AdminHandler{validate: validate, adminService: adminService, skillService: skillService}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.adminService.ListUsers(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "users_load_failed", "Could not list users")
		return
	}
	httpx.JSON(w, http.StatusOK, users)
}

func (h *AdminHandler) ListSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := h.adminService.ListSkills(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "skills_load_failed", "Could not list skills")
		return
	}
	httpx.JSON(w, http.StatusOK, skills)
}

func (h *AdminHandler) GetSkill(w http.ResponseWriter, r *http.Request) {
	skill, err := h.skillService.FindByID(r.Context(), chi.URLParam(r, "skillID"))
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "skill_not_found", "Skill not found")
		return
	}
	httpx.JSON(w, http.StatusOK, skill)
}

func (h *AdminHandler) CreateSkill(w http.ResponseWriter, r *http.Request) {
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

	skill, err := h.skillService.Create(r.Context(), claims.UserID, input)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "skill_create_failed", "Could not create skill")
		return
	}
	httpx.JSON(w, http.StatusCreated, skill)
}

func (h *AdminHandler) UpdateSkill(w http.ResponseWriter, r *http.Request) {
	var input service.SkillMutationInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	skill, err := h.skillService.Update(r.Context(), chi.URLParam(r, "skillID"), input)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "skill_update_failed", "Could not update skill")
		return
	}
	httpx.JSON(w, http.StatusOK, skill)
}

func (h *AdminHandler) DeleteSkill(w http.ResponseWriter, r *http.Request) {
	if err := h.skillService.Archive(r.Context(), chi.URLParam(r, "skillID")); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "skill_delete_failed", "Could not archive skill")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"success": true})
}
