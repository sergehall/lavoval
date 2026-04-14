package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/backend/api/internal/httpx"
	appmiddleware "github.com/sergehall/lavoval/backend/api/internal/middleware"
	"github.com/sergehall/lavoval/backend/api/internal/service"
)

type RuntimeHandler struct {
	validate *validator.Validate
	service  *service.RuntimeService
}

func NewRuntimeHandler(validate *validator.Validate, service *service.RuntimeService) *RuntimeHandler {
	return &RuntimeHandler{validate: validate, service: service}
}

func (h *RuntimeHandler) Run(w http.ResponseWriter, r *http.Request) {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())

	var input service.RuntimeRunInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	run, err := h.service.Run(r.Context(), claims.UserID, input)
	if err != nil {
		if errors.Is(err, service.ErrRuntimeEntrypointNotFound) {
			httpx.Error(w, http.StatusUnprocessableEntity, "runtime_entrypoint_not_found", "Skill runtime entrypoint is not available")
			return
		}

		httpx.Error(w, http.StatusInternalServerError, "skill_run_failed", "Could not execute skill")
		return
	}

	httpx.JSON(w, http.StatusCreated, run)
}

func (h *RuntimeHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())

	runs, err := h.service.ListByUser(r.Context(), claims.UserID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "skill_runs_load_failed", "Could not load skill runs")
		return
	}

	httpx.JSON(w, http.StatusOK, runs)
}

func (h *RuntimeHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())

	run, err := h.service.FindByIDForUser(r.Context(), chi.URLParam(r, "runID"), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrSkillRunForbidden) {
			httpx.Error(w, http.StatusForbidden, "forbidden", "You do not have access to this skill run")
			return
		}

		httpx.Error(w, http.StatusNotFound, "skill_run_not_found", "Skill run not found")
		return
	}

	httpx.JSON(w, http.StatusOK, run)
}
