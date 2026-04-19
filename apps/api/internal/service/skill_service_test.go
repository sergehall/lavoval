package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type skillRepoStub struct {
	created domain.Skill
}

func (s *skillRepoStub) ListPublished(context.Context, domain.SkillFilter) ([]domain.Skill, error) {
	return []domain.Skill{}, nil
}
func (s *skillRepoStub) ListAll(context.Context) ([]domain.Skill, error) {
	return []domain.Skill{}, nil
}
func (s *skillRepoStub) ListByCreatorID(context.Context, string) ([]domain.Skill, error) {
	return []domain.Skill{}, nil
}
func (s *skillRepoStub) FindByID(context.Context, string) (domain.Skill, error) {
	return s.created, nil
}
func (s *skillRepoStub) Create(_ context.Context, skill domain.Skill) (domain.Skill, error) {
	skill.CreatedAt = time.Now()
	skill.UpdatedAt = skill.CreatedAt
	s.created = skill
	return skill, nil
}
func (s *skillRepoStub) Update(_ context.Context, skill domain.Skill) (domain.Skill, error) {
	s.created = skill
	return skill, nil
}
func (s *skillRepoStub) SoftDelete(context.Context, string) error { return nil }
func (*skillRepoStub) UpdateGovernance(_ context.Context, _, _ string, _ domain.SkillStatus, _ *string, _, _ bool) (domain.Skill, error) {
	return domain.Skill{}, nil
}
func (*skillRepoStub) UpdatePricing(_ context.Context, _ string, _ int, _ string, _ domain.SkillAccessType) (domain.Skill, error) {
	return domain.Skill{}, nil
}
func (*skillRepoStub) GetStats(_ context.Context) (domain.AdminSkillStats, error) {
	return domain.AdminSkillStats{}, nil
}

type skillVersionRepoStub struct {
	version *domain.SkillVersion
}

func (s skillVersionRepoStub) FindCurrentBySkillID(_ context.Context, _ string) (*domain.SkillVersion, error) {
	return s.version, nil
}

func (s skillVersionRepoStub) SaveCurrent(_ context.Context, version domain.SkillVersion) (domain.SkillVersion, error) {
	version.VersionNo = 1
	version.IsCurrent = true
	return version, nil
}

type enrollmentRepoStub struct{}

func (enrollmentRepoStub) ListByUserID(context.Context, string) ([]domain.Enrollment, error) {
	return []domain.Enrollment{}, nil
}
func (enrollmentRepoStub) ListAll(context.Context) ([]domain.EnrollmentDetail, error) {
	return []domain.EnrollmentDetail{}, nil
}
func (enrollmentRepoStub) Create(_ context.Context, _, _ string) (domain.Enrollment, error) {
	return domain.Enrollment{}, nil
}
func (enrollmentRepoStub) UpdateStatus(_ context.Context, _ string, _ domain.EnrollmentStatus, _ int) (domain.Enrollment, error) {
	return domain.Enrollment{}, nil
}

type skillUserRepoStub struct {
	user domain.User
	err  error
}

func (s skillUserRepoStub) Create(_ context.Context, user domain.User) (domain.User, error) {
	return user, s.err
}
func (s skillUserRepoStub) FindByEmail(_ context.Context, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s skillUserRepoStub) FindByID(_ context.Context, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s skillUserRepoStub) MarkEmailVerified(_ context.Context, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s skillUserRepoStub) UpdatePasswordHash(_ context.Context, _, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s skillUserRepoStub) UpdateRoleAndStatus(_ context.Context, _ string, _ domain.Role, _ domain.AccountStatus) (domain.User, error) {
	return s.user, s.err
}
func (s skillUserRepoStub) UpdateRoleStatusModeration(_ context.Context, _, _ string, _ domain.Role, _ domain.AccountStatus, _ *string) (domain.User, error) {
	return s.user, s.err
}
func (s skillUserRepoStub) BumpSessionVersion(_ context.Context, _ string) (domain.User, error) {
	u := s.user
	u.SessionVersion++
	return u, s.err
}
func (s skillUserRepoStub) GetStats(_ context.Context) (domain.AdminUserStats, error) {
	return domain.AdminUserStats{}, s.err
}
func (s skillUserRepoStub) StartTOTPEnrollment(_ context.Context, _ string, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s skillUserRepoStub) CancelTOTPEnrollment(_ context.Context, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s skillUserRepoStub) EnableTOTP(_ context.Context, _ string, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s skillUserRepoStub) DisableTOTP(_ context.Context, _ string) (domain.User, error) {
	return s.user, s.err
}
func (s skillUserRepoStub) SoftDelete(_ context.Context, _ string) error { return s.err }
func (s skillUserRepoStub) List(_ context.Context) ([]domain.User, error) {
	return []domain.User{s.user}, s.err
}

func TestSkillServiceCreateAssignsActorID(t *testing.T) {
	skillRepo := &skillRepoStub{}
	service := NewSkillService(skillRepo, skillVersionRepoStub{}, enrollmentRepoStub{}, skillUserRepoStub{
		user: domain.User{ID: "admin-1", Status: domain.AccountStatusActive},
	})
	result, err := service.Create(context.Background(), "admin-1", SkillMutationInput{
		Slug:        "clean-architecture",
		Title:       "Clean Architecture",
		Summary:     "Use clear boundaries for sustainable product growth.",
		Description: "Learn how to separate application layers and keep product systems maintainable over time.",
		Provider:    "internal",
		Entrypoint:  "echo",
		Config:      map[string]any{"mode": "test"},
		Status:      domain.SkillStatusDraft,
		Visibility:  domain.VisibilityPrivate,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.CreatedBy != "admin-1" {
		t.Fatalf("expected actor to be attached, got %s", result.CreatedBy)
	}
	if result.Provider != "internal" {
		t.Fatalf("expected provider to be preserved, got %s", result.Provider)
	}
}

func TestSkillServiceCreateRejectsSuspendedCreator(t *testing.T) {
	service := NewSkillService(&skillRepoStub{}, skillVersionRepoStub{}, enrollmentRepoStub{}, skillUserRepoStub{
		user: domain.User{ID: "creator-1", Status: domain.AccountStatusSuspended},
	})

	_, err := service.Create(context.Background(), "creator-1", SkillMutationInput{
		Slug:        "clean-architecture",
		Title:       "Clean Architecture",
		Summary:     "Use clear boundaries for sustainable product growth.",
		Description: "Learn how to separate application layers and keep product systems maintainable over time.",
		Provider:    "internal",
		Entrypoint:  "echo",
		Config:      map[string]any{"mode": "test"},
		Status:      domain.SkillStatusDraft,
		Visibility:  domain.VisibilityPrivate,
	})
	if !errors.Is(err, ErrAccountSuspended) {
		t.Fatalf("expected ErrAccountSuspended, got %v", err)
	}
}

func TestSkillServiceCreateRejectsPromptVariablesOutsideInputSchema(t *testing.T) {
	service := NewSkillService(&skillRepoStub{}, skillVersionRepoStub{}, enrollmentRepoStub{}, skillUserRepoStub{
		user: domain.User{ID: "creator-1", Status: domain.AccountStatusActive},
	})

	_, err := service.Create(context.Background(), "creator-1", SkillMutationInput{
		Slug:               "ai-brief",
		Title:              "AI Brief",
		Summary:            "Create a clear brief for a downstream agent run.",
		Description:        "Use this skill to generate a structured brief for downstream AI operators.",
		Provider:           "internal",
		Entrypoint:         "echo",
		Config:             map[string]any{"mode": "test"},
		Status:             domain.SkillStatusDraft,
		Visibility:         domain.VisibilityPrivate,
		InputSchema:        map[string]any{"type": "object", "properties": map[string]any{"topic": map[string]any{"type": "string"}}},
		PromptTemplate:     "Explain {{topic}} for {{audience}}",
		SystemInstructions: "Stay grounded in the provided schema.",
	})

	var contractErr *SkillContractValidationError
	if !errors.As(err, &contractErr) {
		t.Fatalf("expected skill contract validation error, got %v", err)
	}
}
