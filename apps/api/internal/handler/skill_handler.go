package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/httpx"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type SkillHandler struct {
	validate *validator.Validate
	service  *service.SkillService
}

func NewSkillHandler(validate *validator.Validate, service *service.SkillService) *SkillHandler {
	return &SkillHandler{validate: validate, service: service}
}

func (h *SkillHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := domain.SkillFilter{
		Query:         strings.TrimSpace(q.Get("q")),
		CategoryID:    q.Get("category"),
		SubcategoryID: q.Get("subcategory"),
		Difficulty:    q.Get("difficulty"),
		SkillType:     q.Get("skillType"),
		Sort:          q.Get("sort"),
	}

	if tags := q.Get("tags"); tags != "" {
		for _, t := range strings.Split(tags, ",") {
			if t = strings.TrimSpace(t); t != "" {
				filter.TagSlugs = append(filter.TagSlugs, t)
			}
		}
	}

	if agentReady := q.Get("agentReady"); agentReady == "true" {
		t := true
		filter.IsAgentReady = &t
	} else if agentReady == "false" {
		f := false
		filter.IsAgentReady = &f
	}

	skills, err := h.service.ListPublic(r.Context(), filter)
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
