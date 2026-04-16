package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type handlerSkillRepoStub struct {
	skill domain.Skill
	err   error
}

func (s handlerSkillRepoStub) ListPublished(_ context.Context) ([]domain.Skill, error) {
	return []domain.Skill{s.skill}, nil
}
func (s handlerSkillRepoStub) ListAll(_ context.Context) ([]domain.Skill, error) {
	return []domain.Skill{s.skill}, nil
}
func (s handlerSkillRepoStub) ListByCreatorID(_ context.Context, _ string) ([]domain.Skill, error) {
	return []domain.Skill{}, nil
}
func (s handlerSkillRepoStub) FindByID(_ context.Context, _ string) (domain.Skill, error) {
	return s.skill, s.err
}
func (s handlerSkillRepoStub) Create(_ context.Context, skill domain.Skill) (domain.Skill, error) {
	return skill, nil
}
func (s handlerSkillRepoStub) Update(_ context.Context, skill domain.Skill) (domain.Skill, error) {
	return skill, nil
}
func (s handlerSkillRepoStub) SoftDelete(_ context.Context, _ string) error { return nil }
func (handlerSkillRepoStub) UpdateGovernance(_ context.Context, _, _ string, _ domain.SkillStatus, _ *string, _, _ bool) (domain.Skill, error) {
	return domain.Skill{}, nil
}
func (handlerSkillRepoStub) UpdatePricing(_ context.Context, _ string, _ int, _ string, _ domain.SkillAccessType) (domain.Skill, error) {
	return domain.Skill{}, nil
}
func (handlerSkillRepoStub) GetStats(_ context.Context) (domain.AdminSkillStats, error) {
	return domain.AdminSkillStats{}, nil
}

type handlerEnrollmentRepoStub struct{}

func (handlerEnrollmentRepoStub) ListByUserID(_ context.Context, _ string) ([]domain.Enrollment, error) {
	return []domain.Enrollment{}, nil
}
func (handlerEnrollmentRepoStub) ListAll(_ context.Context) ([]domain.EnrollmentDetail, error) {
	return []domain.EnrollmentDetail{}, nil
}
func (handlerEnrollmentRepoStub) Create(_ context.Context, _, _ string) (domain.Enrollment, error) {
	return domain.Enrollment{}, nil
}
func (handlerEnrollmentRepoStub) UpdateStatus(_ context.Context, _ string, _ domain.EnrollmentStatus, _ int) (domain.Enrollment, error) {
	return domain.Enrollment{}, nil
}

func newSkillHandlerWithRepo(repo handlerSkillRepoStub) *SkillHandler {
	skillSvc := service.NewSkillService(repo, handlerEnrollmentRepoStub{}, nil)
	return NewSkillHandler(validator.New(validator.WithRequiredStructEnabled()), skillSvc)
}

func withChiSkillID(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("skillID", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestSkillHandlerFindByIDRejectsUnpublishedSkill(t *testing.T) {
	h := newSkillHandlerWithRepo(handlerSkillRepoStub{
		skill: domain.Skill{
			ID:         "skill-1",
			Status:     domain.SkillStatusDraft,
			Visibility: domain.VisibilityPublic,
		},
	})

	r := withChiSkillID(httptest.NewRequest(http.MethodGet, "/api/v1/skills/skill-1", nil), "skill-1")
	w := httptest.NewRecorder()
	h.FindByID(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for draft skill, got %d", w.Code)
	}
}

func TestSkillHandlerFindByIDRejectsPrivateSkill(t *testing.T) {
	h := newSkillHandlerWithRepo(handlerSkillRepoStub{
		skill: domain.Skill{
			ID:         "skill-2",
			Status:     domain.SkillStatusPublished,
			Visibility: domain.VisibilityPrivate,
		},
	})

	r := withChiSkillID(httptest.NewRequest(http.MethodGet, "/api/v1/skills/skill-2", nil), "skill-2")
	w := httptest.NewRecorder()
	h.FindByID(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for private skill, got %d", w.Code)
	}
}

func TestSkillHandlerFindByIDReturnsPublishedPublicSkill(t *testing.T) {
	h := newSkillHandlerWithRepo(handlerSkillRepoStub{
		skill: domain.Skill{
			ID:         "skill-3",
			Title:      "Clean Architecture",
			Status:     domain.SkillStatusPublished,
			Visibility: domain.VisibilityPublic,
		},
	})

	r := withChiSkillID(httptest.NewRequest(http.MethodGet, "/api/v1/skills/skill-3", nil), "skill-3")
	w := httptest.NewRecorder()
	h.FindByID(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for published public skill, got %d", w.Code)
	}
}

func TestSkillHandlerListPublicReturnsOK(t *testing.T) {
	h := newSkillHandlerWithRepo(handlerSkillRepoStub{
		skill: domain.Skill{ID: "skill-1", Status: domain.SkillStatusPublished, Visibility: domain.VisibilityPublic},
	})

	r := httptest.NewRequest(http.MethodGet, "/api/v1/skills", nil)
	w := httptest.NewRecorder()
	h.ListPublic(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
