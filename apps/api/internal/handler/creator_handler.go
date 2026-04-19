package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/sergehall/lavoval/apps/api/internal/httpx"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type CreatorHandler struct {
	service *service.CreatorService
}

func NewCreatorHandler(service *service.CreatorService) *CreatorHandler {
	return &CreatorHandler{service: service}
}

func (h *CreatorHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := h.service.FindPublicByUserID(r.Context(), chi.URLParam(r, "creatorID"))
	if err != nil {
		if errors.Is(err, service.ErrPublicProfileNotFound) {
			httpx.Error(w, http.StatusNotFound, "creator_not_found", "Public creator profile not found")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "creator_load_failed", "Could not load creator profile")
		return
	}

	httpx.JSON(w, http.StatusOK, profile)
}
