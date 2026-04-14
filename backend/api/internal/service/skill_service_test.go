package service

import (
	"context"
	"testing"
	"time"

	"github.com/sergehall/lavoval/backend/api/internal/domain"
)

type skillRepoStub struct {
	created domain.Skill
}

func (s skillRepoStub) ListPublished(context.Context) ([]domain.Skill, error) {
	return []domain.Skill{}, nil
}
func (s skillRepoStub) ListAll(context.Context) ([]domain.Skill, error) { return []domain.Skill{}, nil }
func (s skillRepoStub) ListByCreatorID(context.Context, string) ([]domain.Skill, error) {
	return []domain.Skill{}, nil
}
func (s skillRepoStub) FindByID(context.Context, string) (domain.Skill, error) { return s.created, nil }
func (s skillRepoStub) Create(_ context.Context, skill domain.Skill) (domain.Skill, error) {
	skill.CreatedAt = time.Now()
	skill.UpdatedAt = skill.CreatedAt
	return skill, nil
}
func (s skillRepoStub) Update(_ context.Context, skill domain.Skill) (domain.Skill, error) {
	return skill, nil
}
func (s skillRepoStub) SoftDelete(context.Context, string) error { return nil }

type enrollmentRepoStub struct{}

func (enrollmentRepoStub) ListByUserID(context.Context, string) ([]domain.Enrollment, error) {
	return []domain.Enrollment{}, nil
}

func TestSkillServiceCreateAssignsActorID(t *testing.T) {
	service := NewSkillService(skillRepoStub{}, enrollmentRepoStub{})
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
