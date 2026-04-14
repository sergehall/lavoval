package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/sergehall/lavoval/backend/api/internal/domain"
	"github.com/sergehall/lavoval/backend/api/internal/repository"
	appRuntime "github.com/sergehall/lavoval/backend/api/internal/runtime"
)

var (
	ErrRuntimeEntrypointNotFound = errors.New("runtime entrypoint not found")
	ErrSkillRunForbidden         = errors.New("skill run access forbidden")
)

type RuntimeService struct {
	skills   SkillLookupStore
	runs     repository.SkillRunStore
	registry executorRegistry
}

type SkillLookupStore interface {
	FindByID(context.Context, string) (domain.Skill, error)
}

type executorRegistry interface {
	Find(entrypoint string) (appRuntime.Executor, error)
}

type RuntimeRunInput struct {
	SkillID string         `json:"skillId" validate:"required,uuid"`
	Input   map[string]any `json:"input"`
}

func NewRuntimeService(skills SkillLookupStore, runs repository.SkillRunStore, registry executorRegistry) *RuntimeService {
	return &RuntimeService{
		skills:   skills,
		runs:     runs,
		registry: registry,
	}
}

func (s *RuntimeService) Run(ctx context.Context, userID string, input RuntimeRunInput) (domain.SkillRun, error) {
	skill, err := s.skills.FindByID(ctx, input.SkillID)
	if err != nil {
		return domain.SkillRun{}, fmt.Errorf("find skill for runtime: %w", err)
	}

	executor, err := s.registry.Find(skill.Entrypoint)
	if err != nil {
		return domain.SkillRun{}, ErrRuntimeEntrypointNotFound
	}

	runID := uuid.NewString()
	createdRun, err := s.runs.Create(ctx, domain.SkillRun{
		ID:      runID,
		SkillID: skill.ID,
		UserID:  userID,
		Status:  domain.SkillRunStatusRunning,
		Input:   normalizeMap(input.Input),
	})
	if err != nil {
		return domain.SkillRun{}, fmt.Errorf("create skill run: %w", err)
	}

	startedAt := createdRun.CreatedAt
	createdRun.StartedAt = &startedAt

	output, execErr := executor.Run(ctx, createdRun.Input, normalizeMap(skill.Config))
	finishedAt := time.Now().UTC()
	createdRun.FinishedAt = &finishedAt

	if execErr != nil {
		createdRun.Status = domain.SkillRunStatusFailed
		message := execErr.Error()
		createdRun.ErrorMessage = &message

		updatedRun, updateErr := s.runs.Update(ctx, createdRun)
		if updateErr != nil {
			return domain.SkillRun{}, fmt.Errorf("update failed skill run: %w", updateErr)
		}
		runWithDetails, findErr := s.runs.FindByID(ctx, updatedRun.ID)
		if findErr != nil {
			return domain.SkillRun{}, fmt.Errorf("reload failed skill run: %w", findErr)
		}
		return runWithDetails, nil
	}

	createdRun.Status = domain.SkillRunStatusCompleted
	createdRun.Output = normalizeMap(output)

	updatedRun, err := s.runs.Update(ctx, createdRun)
	if err != nil {
		return domain.SkillRun{}, fmt.Errorf("update completed skill run: %w", err)
	}

	runWithDetails, err := s.runs.FindByID(ctx, updatedRun.ID)
	if err != nil {
		return domain.SkillRun{}, fmt.Errorf("reload completed skill run: %w", err)
	}

	return runWithDetails, nil
}

func (s *RuntimeService) ListByUser(ctx context.Context, userID string) ([]domain.SkillRun, error) {
	runs, err := s.runs.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list skill runs: %w", err)
	}

	return runs, nil
}

func (s *RuntimeService) FindByIDForUser(ctx context.Context, id string, userID string) (domain.SkillRun, error) {
	run, err := s.runs.FindByID(ctx, id)
	if err != nil {
		return domain.SkillRun{}, fmt.Errorf("find skill run: %w", err)
	}
	if run.UserID != userID {
		return domain.SkillRun{}, ErrSkillRunForbidden
	}

	return run, nil
}

func (s *RuntimeService) ListAll(ctx context.Context) ([]domain.SkillRun, error) {
	runs, err := s.runs.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all skill runs: %w", err)
	}

	return runs, nil
}

func (s *RuntimeService) FindByID(ctx context.Context, id string) (domain.SkillRun, error) {
	run, err := s.runs.FindByID(ctx, id)
	if err != nil {
		return domain.SkillRun{}, fmt.Errorf("find skill run: %w", err)
	}

	return run, nil
}

func normalizeMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	return input
}
