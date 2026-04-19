package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/apps/api/internal/httpx"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type CatalogHandler struct {
	validate *validator.Validate
	catalog  *service.CatalogService
	agents   *service.AgentService
}

func NewCatalogHandler(validate *validator.Validate, catalog *service.CatalogService, agents *service.AgentService) *CatalogHandler {
	return &CatalogHandler{validate: validate, catalog: catalog, agents: agents}
}

func (h *CatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.catalog.ListCategories(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "categories_load_failed", "Could not load categories")
		return
	}
	httpx.JSON(w, http.StatusOK, cats)
}

func (h *CatalogHandler) ListSubcategories(w http.ResponseWriter, r *http.Request) {
	subs, err := h.catalog.ListSubcategories(r.Context(), chi.URLParam(r, "categoryID"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "subcategories_load_failed", "Could not load subcategories")
		return
	}
	httpx.JSON(w, http.StatusOK, subs)
}

func (h *CatalogHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.catalog.ListTags(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "tags_load_failed", "Could not load tags")
		return
	}
	httpx.JSON(w, http.StatusOK, tags)
}

func (h *CatalogHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := h.agents.List(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "agents_load_failed", "Could not load agents")
		return
	}
	httpx.JSON(w, http.StatusOK, agents)
}

func (h *CatalogHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	agent, err := h.agents.FindBySlug(r.Context(), chi.URLParam(r, "agentSlug"))
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "agent_not_found", "Agent not found")
		return
	}
	httpx.JSON(w, http.StatusOK, agent)
}

func (h *CatalogHandler) RecommendedAgents(w http.ResponseWriter, r *http.Request) {
	items, err := h.agents.RecommendedForSkill(r.Context(), chi.URLParam(r, "skillID"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "agents_load_failed", "Could not load recommended agents")
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}
