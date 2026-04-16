package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

var ErrSkillForbidden = errors.New("skill access forbidden")

type SkillService struct {
	skills      repository.SkillStore
	enrollments repository.EnrollmentStore
	users       repository.UserStore
}

type SkillMutationInput struct {
	Slug        string             `json:"slug" validate:"required,min=2"`
	Title       string             `json:"title" validate:"required,min=3"`
	Summary     string             `json:"summary" validate:"required,min=10"`
	Description string             `json:"description" validate:"required,min=20"`
	Provider    string             `json:"provider" validate:"required,min=2"`
	Entrypoint  string             `json:"entrypoint" validate:"required,min=2"`
	Config      map[string]any     `json:"config"`
	Status      domain.SkillStatus `json:"status" validate:"required,oneof=draft published archived"`
	Visibility  domain.Visibility  `json:"visibility" validate:"required,oneof=public private"`
}

func NewSkillService(
	skills repository.SkillStore,
	enrollments repository.EnrollmentStore,
	users repository.UserStore,
) *SkillService {
	return &SkillService{skills: skills, enrollments: enrollments, users: users}
}

func (s *SkillService) ListPublic(ctx context.Context) ([]domain.Skill, error) {
	items, err := s.skills.ListPublished(ctx)
	if err != nil {
		return nil, fmt.Errorf("list public skills: %w", err)
	}
	return items, nil
}

func (s *SkillService) FindByID(ctx context.Context, id string) (domain.Skill, error) {
	skill, err := s.skills.FindByID(ctx, id)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("find skill: %w", err)
	}
	return skill, nil
}

func (s *SkillService) ListByCreator(ctx context.Context, creatorID string) ([]domain.Skill, error) {
	items, err := s.skills.ListByCreatorID(ctx, creatorID)
	if err != nil {
		return nil, fmt.Errorf("list creator skills: %w", err)
	}
	return items, nil
}

func (s *SkillService) FindOwnedByCreator(ctx context.Context, id string, creatorID string) (domain.Skill, error) {
	skill, err := s.FindByID(ctx, id)
	if err != nil {
		return domain.Skill{}, err
	}
	if skill.CreatedBy != creatorID {
		return domain.Skill{}, ErrSkillForbidden
	}
	return skill, nil
}

func (s *SkillService) Create(ctx context.Context, actorID string, input SkillMutationInput) (domain.Skill, error) {
	if err := s.ensureCreatorCanMutate(ctx, actorID); err != nil {
		return domain.Skill{}, err
	}

	skill := domain.Skill{
		ID:          uuid.NewString(),
		Slug:        input.Slug,
		Title:       input.Title,
		Summary:     input.Summary,
		Description: input.Description,
		Provider:    input.Provider,
		Entrypoint:  input.Entrypoint,
		Config:      input.Config,
		Status:      input.Status,
		Visibility:  input.Visibility,
		CreatedBy:   actorID,
	}

	createdSkill, err := s.skills.Create(ctx, skill)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("create skill: %w", err)
	}
	return createdSkill, nil
}

func (s *SkillService) Update(ctx context.Context, id string, input SkillMutationInput) (domain.Skill, error) {
	skill := domain.Skill{
		ID:          id,
		Slug:        input.Slug,
		Title:       input.Title,
		Summary:     input.Summary,
		Description: input.Description,
		Provider:    input.Provider,
		Entrypoint:  input.Entrypoint,
		Config:      input.Config,
		Status:      input.Status,
		Visibility:  input.Visibility,
	}

	updatedSkill, err := s.skills.Update(ctx, skill)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("update skill: %w", err)
	}
	return updatedSkill, nil
}

func (s *SkillService) UpdateOwnedByCreator(ctx context.Context, id string, creatorID string, input SkillMutationInput) (domain.Skill, error) {
	if err := s.ensureCreatorCanMutate(ctx, creatorID); err != nil {
		return domain.Skill{}, err
	}

	skill, err := s.FindOwnedByCreator(ctx, id, creatorID)
	if err != nil {
		return domain.Skill{}, err
	}

	updatedSkill, err := s.skills.Update(ctx, domain.Skill{
		ID:          skill.ID,
		Slug:        input.Slug,
		Title:       input.Title,
		Summary:     input.Summary,
		Description: input.Description,
		Provider:    input.Provider,
		Entrypoint:  input.Entrypoint,
		Config:      input.Config,
		Status:      input.Status,
		Visibility:  input.Visibility,
	})
	if err != nil {
		return domain.Skill{}, fmt.Errorf("update owned skill: %w", err)
	}
	return updatedSkill, nil
}

func (s *SkillService) Archive(ctx context.Context, id string) error {
	if err := s.skills.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("archive skill: %w", err)
	}
	return nil
}

func (s *SkillService) ArchiveOwnedByCreator(ctx context.Context, id string, creatorID string) error {
	if err := s.ensureCreatorCanMutate(ctx, creatorID); err != nil {
		return err
	}
	if _, err := s.FindOwnedByCreator(ctx, id, creatorID); err != nil {
		return err
	}
	if err := s.skills.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("archive owned skill: %w", err)
	}
	return nil
}

func (s *SkillService) ensureCreatorCanMutate(ctx context.Context, creatorID string) error {
	if s.users == nil {
		return nil
	}

	user, err := s.users.FindByID(ctx, creatorID)
	if err != nil {
		return fmt.Errorf("find creator account: %w", err)
	}

	switch user.Status {
	case domain.AccountStatusBlocked:
		return ErrAccountBlocked
	case domain.AccountStatusSuspended:
		return ErrAccountSuspended
	default:
		return nil
	}
}
