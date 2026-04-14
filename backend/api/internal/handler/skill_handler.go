package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/backend/api/internal/httpx"
	"github.com/sergehall/lavoval/backend/api/internal/service"
)

type SkillHandler struct {
	validate *validator.Validate
	service  *service.SkillService
}

func NewSkillHandler(validate *validator.Validate, service *service.SkillService) *SkillHandler {
	return &SkillHandler{validate: validate, service: service}
}

func (h *SkillHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	skills, err := h.service.ListPublic(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "skills_load_failed", "Could not list skills")
		return
	}
	httpx.JSON(w, http.StatusOK, skills)
}

func (h *SkillHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	skill, err := h.service.FindByID(r.Context(), chi.URLParam(r, "skillID"))
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "skill_not_found", "Skill not found")
		return
	}
	if skill.Status != "published" || skill.Visibility != "public" {
		httpx.Error(w, http.StatusNotFound, "skill_not_found", "Skill not found")
		return
	}
	httpx.JSON(w, http.StatusOK, skill)
}
