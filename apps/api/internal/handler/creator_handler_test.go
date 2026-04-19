package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/httpx"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type creatorProfileHandlerRepoStub struct {
	profile domain.Profile
	err     error
}

func (s creatorProfileHandlerRepoStub) Create(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	return profile, nil
}

func (s creatorProfileHandlerRepoStub) Update(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	return profile, nil
}

func (s creatorProfileHandlerRepoStub) FindByUserID(_ context.Context, _ string) (domain.Profile, error) {
	if s.err != nil {
		return domain.Profile{}, s.err
	}
	return s.profile, nil
}

func (s creatorProfileHandlerRepoStub) SoftDeleteByUserID(_ context.Context, _ string) error {
	return nil
}

type creatorSkillHandlerRepoStub struct {
	skills []domain.Skill
	err    error
}

func (s creatorSkillHandlerRepoStub) ListPublished(context.Context, domain.SkillFilter) ([]domain.Skill, error) {
	return nil, nil
}
func (s creatorSkillHandlerRepoStub) ListAll(context.Context) ([]domain.Skill, error) {
	return nil, nil
}
func (s creatorSkillHandlerRepoStub) ListByCreatorID(_ context.Context, _ string) ([]domain.Skill, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.skills, nil
}
func (s creatorSkillHandlerRepoStub) FindByID(context.Context, string) (domain.Skill, error) {
	return domain.Skill{}, nil
}
func (s creatorSkillHandlerRepoStub) Create(context.Context, domain.Skill) (domain.Skill, error) {
	return domain.Skill{}, nil
}
func (s creatorSkillHandlerRepoStub) Update(context.Context, domain.Skill) (domain.Skill, error) {
	return domain.Skill{}, nil
}
func (s creatorSkillHandlerRepoStub) UpdateGovernance(context.Context, string, string, domain.SkillStatus, *string, bool, bool) (domain.Skill, error) {
	return domain.Skill{}, nil
}
func (s creatorSkillHandlerRepoStub) UpdatePricing(context.Context, string, int, string, domain.SkillAccessType) (domain.Skill, error) {
	return domain.Skill{}, nil
}
func (s creatorSkillHandlerRepoStub) GetStats(context.Context) (domain.AdminSkillStats, error) {
	return domain.AdminSkillStats{}, nil
}
func (s creatorSkillHandlerRepoStub) SoftDelete(context.Context, string) error {
	return nil
}

func TestCreatorHandlerGetPublicProfileReturnsProfile(t *testing.T) {
	h := NewCreatorHandler(service.NewCreatorService(
		creatorProfileHandlerRepoStub{
			profile: domain.Profile{
				UserID:             "user-1",
				FirstName:          "Serge",
				LastName:           "Hall",
				IsPublicProfile:    true,
				ShowBio:            true,
				ShowAvatar:         true,
				ShowSkills:         true,
				ShowAvailability:   true,
				ShowWebsiteURL:     true,
				ShowLinkedInURL:    true,
				ShowGitHubURL:      true,
				ShowTwitterURL:     true,
				ShowLocation:       true,
				ShowLanguages:      true,
				AvailabilityStatus: domain.AvailabilityOpen,
			},
		},
		creatorSkillHandlerRepoStub{},
	))

	r := httptest.NewRequest(http.MethodGet, "/api/v1/creators/user-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("creatorID", "user-1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.GetPublicProfile(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var envelope httpx.Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("expected valid JSON envelope, got %v", err)
	}
	if envelope.Error != nil {
		t.Fatalf("expected no error, got %+v", envelope.Error)
	}
}

func TestCreatorHandlerGetPublicProfileReturnsNotFoundForPrivateProfile(t *testing.T) {
	h := NewCreatorHandler(service.NewCreatorService(
		creatorProfileHandlerRepoStub{profile: domain.Profile{IsPublicProfile: false}},
		creatorSkillHandlerRepoStub{},
	))

	r := httptest.NewRequest(http.MethodGet, "/api/v1/creators/user-2", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("creatorID", "user-2")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.GetPublicProfile(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCreatorHandlerGetPublicProfileReturnsInternalErrorOnLoadFailure(t *testing.T) {
	h := NewCreatorHandler(service.NewCreatorService(
		creatorProfileHandlerRepoStub{err: errors.New("db unavailable")},
		creatorSkillHandlerRepoStub{},
	))

	r := httptest.NewRequest(http.MethodGet, "/api/v1/creators/user-3", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("creatorID", "user-3")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.GetPublicProfile(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
