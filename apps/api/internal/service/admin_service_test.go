package service

import (
	"context"
	"errors"
	"testing"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

// ── store stubs ───────────────────────────────────────────────────────────────

type adminUserStub struct {
	user domain.User
	err  error
}

func (s adminUserStub) Create(_ context.Context, u domain.User) (domain.User, error) {
	return u, s.err
}
func (s adminUserStub) FindByEmail(_ context.Context, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s adminUserStub) FindByID(_ context.Context, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s adminUserStub) MarkEmailVerified(_ context.Context, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s adminUserStub) UpdatePasswordHash(_ context.Context, _, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s adminUserStub) UpdateRoleAndStatus(_ context.Context, _ string, role domain.Role, status domain.AccountStatus) (domain.User, error) {
	u := s.user
	u.Role = role
	u.Status = status
	return u, s.err
}
func (s adminUserStub) StartTOTPEnrollment(_ context.Context, _ string, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s adminUserStub) CancelTOTPEnrollment(_ context.Context, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s adminUserStub) EnableTOTP(_ context.Context, _ string, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s adminUserStub) DisableTOTP(_ context.Context, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s adminUserStub) SoftDelete(_ context.Context, _ string) error {
	return s.err
}
func (s adminUserStub) List(_ context.Context) ([]domain.User, error) {
	return []domain.User{s.user}, s.err
}

type adminProfileStub struct {
	profile domain.Profile
	err     error
}

func (s adminProfileStub) Create(_ context.Context, p domain.Profile) (domain.Profile, error) {
	return p, s.err
}
func (s adminProfileStub) Update(_ context.Context, p domain.Profile) (domain.Profile, error) {
	return p, s.err
}
func (s adminProfileStub) FindByUserID(_ context.Context, _ string) (domain.Profile, error) {
	return s.profile, s.err
}
func (s adminProfileStub) SoftDeleteByUserID(_ context.Context, _ string) error {
	return s.err
}

type adminSkillStub struct{}

func (adminSkillStub) ListPublished(_ context.Context) ([]domain.Skill, error) {
	return []domain.Skill{}, nil
}
func (adminSkillStub) ListAll(_ context.Context) ([]domain.Skill, error) {
	return []domain.Skill{}, nil
}
func (adminSkillStub) ListByCreatorID(_ context.Context, _ string) ([]domain.Skill, error) {
	return []domain.Skill{}, nil
}
func (adminSkillStub) FindByID(_ context.Context, _ string) (domain.Skill, error) {
	return domain.Skill{}, nil
}
func (adminSkillStub) Create(_ context.Context, s domain.Skill) (domain.Skill, error) {
	return s, nil
}
func (adminSkillStub) Update(_ context.Context, s domain.Skill) (domain.Skill, error) {
	return s, nil
}
func (adminSkillStub) SoftDelete(_ context.Context, _ string) error { return nil }

type adminEnrollmentStub struct {
	enrollment domain.Enrollment
	details    []domain.EnrollmentDetail
	err        error
}

func (s adminEnrollmentStub) ListByUserID(_ context.Context, _ string) ([]domain.Enrollment, error) {
	return []domain.Enrollment{s.enrollment}, s.err
}
func (s adminEnrollmentStub) ListAll(_ context.Context) ([]domain.EnrollmentDetail, error) {
	return s.details, s.err
}
func (s adminEnrollmentStub) Create(_ context.Context, userID, skillID string) (domain.Enrollment, error) {
	e := s.enrollment
	e.UserID = userID
	e.SkillID = skillID
	return e, s.err
}
func (s adminEnrollmentStub) UpdateStatus(_ context.Context, id string, status domain.EnrollmentStatus, progress int) (domain.Enrollment, error) {
	e := s.enrollment
	e.ID = id
	e.Status = status
	e.ProgressPercent = progress
	return e, s.err
}

type adminModuleStub struct {
	module domain.Module
	err    error
}

func (s adminModuleStub) ListBySkillID(_ context.Context, _ string) ([]domain.Module, error) {
	return []domain.Module{s.module}, s.err
}
func (s adminModuleStub) FindByID(_ context.Context, _ string) (domain.Module, error) {
	return s.module, s.err
}
func (s adminModuleStub) Create(_ context.Context, m domain.Module) (domain.Module, error) {
	return m, s.err
}
func (s adminModuleStub) Update(_ context.Context, m domain.Module) (domain.Module, error) {
	return m, s.err
}
func (s adminModuleStub) SoftDelete(_ context.Context, _ string) error { return s.err }

// ── helpers ───────────────────────────────────────────────────────────────────

func newAdminSvc(
	users adminUserStub,
	profiles adminProfileStub,
	enrollments adminEnrollmentStub,
	modules adminModuleStub,
) *AdminService {
	return NewAdminService(users, profiles, adminSkillStub{}, enrollments, modules, nil, nil)
}

// ── user tests ────────────────────────────────────────────────────────────────

func TestAdminGetUserReturnsDetail(t *testing.T) {
	user := domain.User{ID: "u1", Email: "alice@example.com", Role: domain.RoleUser}
	profile := domain.Profile{UserID: "u1", FirstName: "Alice"}

	svc := newAdminSvc(
		adminUserStub{user: user},
		adminProfileStub{profile: profile},
		adminEnrollmentStub{},
		adminModuleStub{},
	)

	detail, err := svc.GetUser(context.Background(), "u1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if detail.User.ID != "u1" {
		t.Fatalf("expected user id u1, got %s", detail.User.ID)
	}
	if detail.Profile.FirstName != "Alice" {
		t.Fatalf("expected profile firstName Alice, got %s", detail.Profile.FirstName)
	}
}

func TestAdminGetUserPropagatesUserError(t *testing.T) {
	svc := newAdminSvc(
		adminUserStub{err: errors.New("not found")},
		adminProfileStub{},
		adminEnrollmentStub{},
		adminModuleStub{},
	)

	_, err := svc.GetUser(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAdminUpdateUserSetsRoleAndStatus(t *testing.T) {
	svc := newAdminSvc(
		adminUserStub{user: domain.User{ID: "u1"}},
		adminProfileStub{},
		adminEnrollmentStub{},
		adminModuleStub{},
	)

	input := UpdateUserInput{Role: domain.RoleAdmin, Status: domain.AccountStatusSuspended}
	user, err := svc.UpdateUser(context.Background(), "u1", input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Role != domain.RoleAdmin {
		t.Fatalf("expected role admin, got %s", user.Role)
	}
	if user.Status != domain.AccountStatusSuspended {
		t.Fatalf("expected status suspended, got %s", user.Status)
	}
}

// ── enrollment tests ──────────────────────────────────────────────────────────

func TestAdminListEnrollmentsReturnsDetails(t *testing.T) {
	details := []domain.EnrollmentDetail{
		{Enrollment: domain.Enrollment{ID: "e1"}, UserEmail: "bob@example.com", SkillTitle: "Go Basics"},
		{Enrollment: domain.Enrollment{ID: "e2"}, UserEmail: "eve@example.com", SkillTitle: "Clean Arch"},
	}

	svc := newAdminSvc(
		adminUserStub{},
		adminProfileStub{},
		adminEnrollmentStub{details: details},
		adminModuleStub{},
	)

	result, err := svc.ListEnrollments(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 enrollments, got %d", len(result))
	}
	if result[0].UserEmail != "bob@example.com" {
		t.Fatalf("expected userEmail bob@example.com, got %s", result[0].UserEmail)
	}
}

func TestAdminAssignSkillAttachesIDs(t *testing.T) {
	svc := newAdminSvc(
		adminUserStub{},
		adminProfileStub{},
		adminEnrollmentStub{},
		adminModuleStub{},
	)

	input := AssignSkillInput{
		UserID:  "00000000-0000-4000-8000-000000000001",
		SkillID: "00000000-0000-4000-8000-000000000002",
	}
	enrollment, err := svc.AssignSkill(context.Background(), input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if enrollment.UserID != input.UserID {
		t.Fatalf("expected userID %s, got %s", input.UserID, enrollment.UserID)
	}
	if enrollment.SkillID != input.SkillID {
		t.Fatalf("expected skillID %s, got %s", input.SkillID, enrollment.SkillID)
	}
}

func TestAdminUpdateEnrollmentSetsStatusAndProgress(t *testing.T) {
	svc := newAdminSvc(
		adminUserStub{},
		adminProfileStub{},
		adminEnrollmentStub{},
		adminModuleStub{},
	)

	input := UpdateEnrollmentInput{Status: domain.EnrollmentStatusInProgress, ProgressPercent: 42}
	enrollment, err := svc.UpdateEnrollment(context.Background(), "e1", input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if enrollment.Status != domain.EnrollmentStatusInProgress {
		t.Fatalf("expected status in_progress, got %s", enrollment.Status)
	}
	if enrollment.ProgressPercent != 42 {
		t.Fatalf("expected progress 42, got %d", enrollment.ProgressPercent)
	}
}

// ── module tests ──────────────────────────────────────────────────────────────

func TestAdminListModulesReturnsSkillModules(t *testing.T) {
	mod := domain.Module{ID: "m1", SkillID: "sk1", Title: "Intro"}
	svc := newAdminSvc(
		adminUserStub{},
		adminProfileStub{},
		adminEnrollmentStub{},
		adminModuleStub{module: mod},
	)

	modules, err := svc.ListModules(context.Background(), "sk1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(modules) != 1 || modules[0].Title != "Intro" {
		t.Fatalf("unexpected modules result: %+v", modules)
	}
}

func TestAdminCreateModulePreservesInput(t *testing.T) {
	svc := newAdminSvc(
		adminUserStub{},
		adminProfileStub{},
		adminEnrollmentStub{},
		adminModuleStub{},
	)

	input := ModuleMutationInput{
		Slug:    "intro-ts",
		Title:   "Intro to TypeScript",
		Summary: "Learn the basics of TypeScript.",
		Content: "TypeScript adds types to JavaScript.",
		Status:  domain.SkillStatusDraft,
	}
	mod, err := svc.CreateModule(context.Background(), "sk1", input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mod.Slug != "intro-ts" {
		t.Fatalf("expected slug intro-ts, got %s", mod.Slug)
	}
	if mod.SkillID != "sk1" {
		t.Fatalf("expected skillID sk1, got %s", mod.SkillID)
	}
	if mod.ID == "" {
		t.Fatal("expected non-empty module ID to be assigned")
	}
}

func TestAdminUpdateModulePreservesID(t *testing.T) {
	svc := newAdminSvc(
		adminUserStub{},
		adminProfileStub{},
		adminEnrollmentStub{},
		adminModuleStub{},
	)

	input := ModuleMutationInput{
		Slug:    "updated-slug",
		Title:   "Updated Title",
		Summary: "Updated summary text here.",
		Content: "Updated content that is long enough.",
		Status:  domain.SkillStatusPublished,
	}
	mod, err := svc.UpdateModule(context.Background(), "m-fixed-id", input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mod.ID != "m-fixed-id" {
		t.Fatalf("expected module ID m-fixed-id, got %s", mod.ID)
	}
	if mod.Status != domain.SkillStatusPublished {
		t.Fatalf("expected status published, got %s", mod.Status)
	}
}

func TestAdminDeleteModuleCallsSoftDelete(t *testing.T) {
	svc := newAdminSvc(
		adminUserStub{},
		adminProfileStub{},
		adminEnrollmentStub{},
		adminModuleStub{},
	)

	if err := svc.DeleteModule(context.Background(), "m1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAdminDeleteModulePropagatesError(t *testing.T) {
	svc := newAdminSvc(
		adminUserStub{},
		adminProfileStub{},
		adminEnrollmentStub{},
		adminModuleStub{err: errors.New("db failure")},
	)

	if err := svc.DeleteModule(context.Background(), "m1"); err == nil {
		t.Fatal("expected error, got nil")
	}
}
