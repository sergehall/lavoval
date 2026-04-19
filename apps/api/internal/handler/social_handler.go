package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/apps/api/internal/httpx"
	appmiddleware "github.com/sergehall/lavoval/apps/api/internal/middleware"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type SocialHandler struct {
	validate *validator.Validate
	social   *service.SocialService
}

func NewSocialHandler(validate *validator.Validate, social *service.SocialService) *SocialHandler {
	return &SocialHandler{validate: validate, social: social}
}

func userIDFromCtx(r *http.Request) string {
	claims, _ := appmiddleware.ClaimsFromContext(r.Context())
	if claims == nil {
		return ""
	}
	return claims.UserID
}

// ── Reviews ───────────────────────────────────────────────────────────────────

func (h *SocialHandler) ListReviews(w http.ResponseWriter, r *http.Request) {
	items, err := h.social.ListReviews(r.Context(), chi.URLParam(r, "skillID"))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "reviews_load_failed", "Could not load reviews")
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *SocialHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromCtx(r)

	var input service.CreateReviewInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}

	rv, err := h.social.CreateReview(r.Context(), chi.URLParam(r, "skillID"), userID, input)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "review_create_failed", "Could not create review")
		return
	}
	httpx.JSON(w, http.StatusCreated, rv)
}

// ── Saves ─────────────────────────────────────────────────────────────────────

func (h *SocialHandler) SaveSkill(w http.ResponseWriter, r *http.Request) {
	if err := h.social.SaveSkill(r.Context(), userIDFromCtx(r), chi.URLParam(r, "skillID")); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "save_failed", "Could not save skill")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SocialHandler) UnsaveSkill(w http.ResponseWriter, r *http.Request) {
	if err := h.social.UnsaveSkill(r.Context(), userIDFromCtx(r), chi.URLParam(r, "skillID")); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "unsave_failed", "Could not unsave skill")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Collections ───────────────────────────────────────────────────────────────

func (h *SocialHandler) ListCollections(w http.ResponseWriter, r *http.Request) {
	items, err := h.social.ListCollections(r.Context(), userIDFromCtx(r))
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "collections_load_failed", "Could not load collections")
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *SocialHandler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	var input service.CollectionInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}

	c, err := h.social.CreateCollection(r.Context(), userIDFromCtx(r), input)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "collection_create_failed", "Could not create collection")
		return
	}
	httpx.JSON(w, http.StatusCreated, c)
}

func (h *SocialHandler) UpdateCollection(w http.ResponseWriter, r *http.Request) {
	var input service.CollectionInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}

	c, err := h.social.UpdateCollection(r.Context(), chi.URLParam(r, "collectionID"), userIDFromCtx(r), input)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "collection_update_failed", "Could not update collection")
		return
	}
	httpx.JSON(w, http.StatusOK, c)
}

func (h *SocialHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	if err := h.social.DeleteCollection(r.Context(), chi.URLParam(r, "collectionID"), userIDFromCtx(r)); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "collection_delete_failed", "Could not delete collection")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SocialHandler) AddCollectionItem(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SkillID string `json:"skillId" validate:"required"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if err := h.validate.Struct(body); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}
	if err := h.social.AddCollectionItem(r.Context(), chi.URLParam(r, "collectionID"), body.SkillID, userIDFromCtx(r)); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "collection_item_failed", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SocialHandler) RemoveCollectionItem(w http.ResponseWriter, r *http.Request) {
	if err := h.social.RemoveCollectionItem(r.Context(), chi.URLParam(r, "collectionID"), chi.URLParam(r, "skillID"), userIDFromCtx(r)); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "collection_item_failed", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Run feedback ─────────────────────────────────────────────────────────────

func (h *SocialHandler) CreateRunFeedback(w http.ResponseWriter, r *http.Request) {
	var input service.RunFeedbackInput
	if err := httpx.Decode(r, &input); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	if err := h.validate.Struct(input); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}

	if err := h.social.CreateRunFeedback(r.Context(), chi.URLParam(r, "runID"), userIDFromCtx(r), input); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "feedback_failed", "Could not save feedback")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
