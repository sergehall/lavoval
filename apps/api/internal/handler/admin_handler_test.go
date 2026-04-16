package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/apps/api/internal/auth"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
	appmiddleware "github.com/sergehall/lavoval/apps/api/internal/middleware"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

// ── store stubs ───────────────────────────────────────────────────────────────

type hAdminUserStub struct {
	users map[string]domain.User
}

func (hAdminUserStub) Create(_ context.Context, u domain.User) (domain.User, error) { return u, nil }
func (hAdminUserStub) FindByEmail(_ context.Context, _ string) (domain.User, error) {
	return domain.User{}, nil
}
func (s hAdminUserStub) FindByID(_ context.Context, id string) (domain.User, error) {
	if s.users != nil {
		return s.users[id], nil
	}
	return domain.User{}, nil
}
func (hAdminUserStub) MarkEmailVerified(_ context.Context, _ string) (domain.User, error) {
	return domain.User{}, nil
}
func (hAdminUserStub) UpdatePasswordHash(_ context.Context, _, _ string) (domain.User, error) {
	return domain.User{}, nil
}
func (hAdminUserStub) UpdateRoleAndStatus(_ context.Context, _ string, _ domain.Role, _ domain.AccountStatus) (domain.User, error) {
	return domain.User{}, nil
}
func (hAdminUserStub) UpdateRoleStatusModeration(_ context.Context, _, _ string, _ domain.Role, _ domain.AccountStatus, _ *string) (domain.User, error) {
	return domain.User{}, nil
}
func (s hAdminUserStub) BumpSessionVersion(_ context.Context, id string) (domain.User, error) {
	if s.users != nil {
		user := s.users[id]
		user.SessionVersion++
		return user, nil
	}
	return domain.User{}, nil
}
func (hAdminUserStub) GetStats(_ context.Context) (domain.AdminUserStats, error) {
	return domain.AdminUserStats{}, nil
}
func (hAdminUserStub) StartTOTPEnrollment(_ context.Context, _ string, _ string) (domain.User, error) {
	return domain.User{}, nil
}
func (hAdminUserStub) CancelTOTPEnrollment(_ context.Context, _ string) (domain.User, error) {
	return domain.User{}, nil
}
func (hAdminUserStub) EnableTOTP(_ context.Context, _ string, _ string) (domain.User, error) {
	return domain.User{}, nil
}
func (hAdminUserStub) DisableTOTP(_ context.Context, _ string) (domain.User, error) {
	return domain.User{}, nil
}
func (hAdminUserStub) SoftDelete(_ context.Context, _ string) error { return nil }
func (s hAdminUserStub) List(_ context.Context) ([]domain.User, error) {
	if s.users == nil {
		return nil, nil
	}
	items := make([]domain.User, 0, len(s.users))
	for _, user := range s.users {
		items = append(items, user)
	}
	return items, nil
}

type hAdminProfileStub struct{}

func (hAdminProfileStub) Create(_ context.Context, p domain.Profile) (domain.Profile, error) {
	return p, nil
}
func (hAdminProfileStub) Update(_ context.Context, p domain.Profile) (domain.Profile, error) {
	return p, nil
}
func (hAdminProfileStub) FindByUserID(_ context.Context, _ string) (domain.Profile, error) {
	return domain.Profile{}, nil
}
func (hAdminProfileStub) SoftDeleteByUserID(_ context.Context, _ string) error { return nil }

type hAdminSkillStub struct{}

func (hAdminSkillStub) ListPublished(_ context.Context) ([]domain.Skill, error) { return nil, nil }
func (hAdminSkillStub) ListAll(_ context.Context) ([]domain.Skill, error)       { return nil, nil }
func (hAdminSkillStub) ListByCreatorID(_ context.Context, _ string) ([]domain.Skill, error) {
	return nil, nil
}
func (hAdminSkillStub) FindByID(_ context.Context, _ string) (domain.Skill, error) {
	return domain.Skill{}, nil
}
func (hAdminSkillStub) Create(_ context.Context, s domain.Skill) (domain.Skill, error) {
	return s, nil
}
func (hAdminSkillStub) Update(_ context.Context, s domain.Skill) (domain.Skill, error) {
	return s, nil
}
func (hAdminSkillStub) SoftDelete(_ context.Context, _ string) error { return nil }
func (hAdminSkillStub) UpdateGovernance(_ context.Context, _, _ string, _ domain.SkillStatus, _ *string, _, _ bool) (domain.Skill, error) {
	return domain.Skill{}, nil
}
func (hAdminSkillStub) UpdatePricing(_ context.Context, _ string, _ int, _ string, _ domain.SkillAccessType) (domain.Skill, error) {
	return domain.Skill{}, nil
}
func (hAdminSkillStub) GetStats(_ context.Context) (domain.AdminSkillStats, error) {
	return domain.AdminSkillStats{}, nil
}

type hAdminEnrollmentStub struct{}

func (hAdminEnrollmentStub) ListByUserID(_ context.Context, _ string) ([]domain.Enrollment, error) {
	return nil, nil
}
func (hAdminEnrollmentStub) ListAll(_ context.Context) ([]domain.EnrollmentDetail, error) {
	return nil, nil
}
func (hAdminEnrollmentStub) Create(_ context.Context, _, _ string) (domain.Enrollment, error) {
	return domain.Enrollment{}, nil
}
func (hAdminEnrollmentStub) UpdateStatus(_ context.Context, _ string, _ domain.EnrollmentStatus, _ int) (domain.Enrollment, error) {
	return domain.Enrollment{}, nil
}

type hAdminModuleStub struct{}

func (hAdminModuleStub) ListBySkillID(_ context.Context, _ string) ([]domain.Module, error) {
	return nil, nil
}
func (hAdminModuleStub) FindByID(_ context.Context, _ string) (domain.Module, error) {
	return domain.Module{}, nil
}
func (hAdminModuleStub) Create(_ context.Context, m domain.Module) (domain.Module, error) {
	return m, nil
}
func (hAdminModuleStub) Update(_ context.Context, m domain.Module) (domain.Module, error) {
	return m, nil
}
func (hAdminModuleStub) SoftDelete(_ context.Context, _ string) error { return nil }

// ── constructor helpers ───────────────────────────────────────────────────────

func newTestAdminHandler() *AdminHandler {
	now := time.Now().UTC()
	adminSvc := service.NewAdminService(
		hAdminUserStub{users: map[string]domain.User{
			"actor-1": {
				ID:              "actor-1",
				Role:            domain.RoleRootOwner,
				Status:          domain.AccountStatusActive,
				EmailVerifiedAt: &now,
				MFAEnabled:      true,
			},
			"u1": {
				ID:              "u1",
				Role:            domain.RoleUser,
				Status:          domain.AccountStatusActive,
				EmailVerifiedAt: &now,
			},
		}},
		hAdminProfileStub{},
		hAdminSkillStub{},
		hAdminEnrollmentStub{},
		hAdminModuleStub{},
		nil,
		nil,
		nil,
		nil,
		service.MailRetentionPolicy{},
	)
	skillSvc := service.NewSkillService(
		hAdminSkillStub{},
		hAdminEnrollmentStub{},
		nil,
	)
	v := validator.New(validator.WithRequiredStructEnabled())
	return NewAdminHandler(v, adminSvc, skillSvc)
}

func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func withClaims(r *http.Request, userID string, role domain.Role) *http.Request {
	claims := &auth.Claims{UserID: userID, Role: role}
	return r.WithContext(appmiddleware.WithClaims(r.Context(), claims))
}

// Satisfy repository interfaces so we can pass the stubs (compile check).
var _ repository.UserStore = hAdminUserStub{}
var _ repository.ProfileStore = hAdminProfileStub{}
var _ repository.SkillStore = hAdminSkillStub{}
var _ repository.EnrollmentStore = hAdminEnrollmentStub{}
var _ repository.ModuleStore = hAdminModuleStub{}

// ── UpdateUser validation ─────────────────────────────────────────────────────

func TestAdminHandlerUpdateUserRejectsInvalidRole(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`{"role":"superuser","status":"active"}`)
	r := httptest.NewRequest(http.MethodPatch, "/admin/users/u1", body)
	r = withChiParam(r, "userID", "u1")
	r = withClaims(r, "actor-1", domain.RoleRootOwner)
	w := httptest.NewRecorder()

	h.UpdateUser(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid role, got %d", w.Code)
	}
}

func TestAdminHandlerUpdateUserRejectsInvalidStatus(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`{"role":"user","status":"banned"}`)
	r := httptest.NewRequest(http.MethodPatch, "/admin/users/u1", body)
	r = withChiParam(r, "userID", "u1")
	r = withClaims(r, "actor-1", domain.RoleRootOwner)
	w := httptest.NewRecorder()

	h.UpdateUser(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid status, got %d", w.Code)
	}
}

func TestAdminHandlerUpdateUserRejectsMalformedJSON(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`not-json`)
	r := httptest.NewRequest(http.MethodPatch, "/admin/users/u1", body)
	r = withChiParam(r, "userID", "u1")
	r = withClaims(r, "actor-1", domain.RoleRootOwner)
	w := httptest.NewRecorder()

	h.UpdateUser(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed JSON, got %d", w.Code)
	}
}

func TestAdminHandlerUpdateUserAcceptsValidPayload(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`{"role":"admin","status":"active","reason":"Promoting trusted operator"}`)
	r := httptest.NewRequest(http.MethodPatch, "/admin/users/u1", body)
	r = withChiParam(r, "userID", "u1")
	r = withClaims(r, "actor-1", domain.RoleRootOwner)
	w := httptest.NewRecorder()

	h.UpdateUser(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid payload, got %d", w.Code)
	}
}

// ── AssignSkill validation ────────────────────────────────────────────────────

func TestAdminHandlerAssignSkillRejectsNonUUID(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`{"userId":"not-a-uuid","skillId":"not-a-uuid"}`)
	r := httptest.NewRequest(http.MethodPost, "/admin/enrollments", body)
	w := httptest.NewRecorder()

	h.AssignSkill(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-UUID IDs, got %d", w.Code)
	}
}

func TestAdminHandlerAssignSkillRejectsMissingFields(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`{}`)
	r := httptest.NewRequest(http.MethodPost, "/admin/enrollments", body)
	w := httptest.NewRecorder()

	h.AssignSkill(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing fields, got %d", w.Code)
	}
}

// ── UpdateEnrollment validation ───────────────────────────────────────────────

func TestAdminHandlerUpdateEnrollmentRejectsInvalidStatus(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`{"status":"unknown","progressPercent":50}`)
	r := httptest.NewRequest(http.MethodPatch, "/admin/enrollments/e1", body)
	r = withChiParam(r, "enrollmentID", "e1")
	w := httptest.NewRecorder()

	h.UpdateEnrollment(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid enrollment status, got %d", w.Code)
	}
}

func TestAdminHandlerUpdateEnrollmentRejectsProgressOver100(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`{"status":"in_progress","progressPercent":150}`)
	r := httptest.NewRequest(http.MethodPatch, "/admin/enrollments/e1", body)
	r = withChiParam(r, "enrollmentID", "e1")
	w := httptest.NewRecorder()

	h.UpdateEnrollment(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for progress > 100, got %d", w.Code)
	}
}

func TestAdminHandlerUpdateEnrollmentAcceptsValidPayload(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`{"status":"completed","progressPercent":100}`)
	r := httptest.NewRequest(http.MethodPatch, "/admin/enrollments/e1", body)
	r = withChiParam(r, "enrollmentID", "e1")
	w := httptest.NewRecorder()

	h.UpdateEnrollment(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid enrollment update, got %d", w.Code)
	}
}

// ── CreateModule validation ───────────────────────────────────────────────────

func TestAdminHandlerCreateModuleRejectsMissingSlug(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`{"title":"Intro","summary":"A summary here.","content":"Some longer content here.","status":"draft"}`)
	r := httptest.NewRequest(http.MethodPost, "/admin/skills/sk1/modules", body)
	r = withChiParam(r, "skillID", "sk1")
	w := httptest.NewRecorder()

	h.CreateModule(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing slug, got %d", w.Code)
	}
}

func TestAdminHandlerCreateModuleRejectsShortContent(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`{"slug":"intro","title":"Intro","summary":"A summary here.","content":"short","status":"draft"}`)
	r := httptest.NewRequest(http.MethodPost, "/admin/skills/sk1/modules", body)
	r = withChiParam(r, "skillID", "sk1")
	w := httptest.NewRecorder()

	h.CreateModule(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for content too short, got %d", w.Code)
	}
}

func TestAdminHandlerCreateModuleRejectsInvalidStatus(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`{"slug":"intro","title":"Intro","summary":"A summary here.","content":"Some longer content here.","status":"unknown"}`)
	r := httptest.NewRequest(http.MethodPost, "/admin/skills/sk1/modules", body)
	r = withChiParam(r, "skillID", "sk1")
	w := httptest.NewRecorder()

	h.CreateModule(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid module status, got %d", w.Code)
	}
}

func TestAdminHandlerCreateModuleAcceptsValidPayload(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`{"slug":"intro-ts","title":"Intro to TypeScript","summary":"Learn the fundamentals.","content":"TypeScript adds static types to JavaScript.","status":"draft"}`)
	r := httptest.NewRequest(http.MethodPost, "/admin/skills/sk1/modules", body)
	r = withChiParam(r, "skillID", "sk1")
	w := httptest.NewRecorder()

	h.CreateModule(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 for valid module, got %d", w.Code)
	}
}

// ── UpdateModule validation ───────────────────────────────────────────────────

func TestAdminHandlerUpdateModuleRejectsMalformedJSON(t *testing.T) {
	h := newTestAdminHandler()

	body := bytes.NewBufferString(`{bad json}`)
	r := httptest.NewRequest(http.MethodPatch, "/admin/skills/sk1/modules/m1", body)
	r = withChiParam(r, "moduleID", "m1")
	w := httptest.NewRecorder()

	h.UpdateModule(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed JSON, got %d", w.Code)
	}
}
