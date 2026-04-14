package service

import (
	"context"
	"errors"
	"testing"

	"github.com/sergehall/lavoval/backend/api/internal/domain"
	appRuntime "github.com/sergehall/lavoval/backend/api/internal/runtime"
)

type skillLookupStub struct {
	skill domain.Skill
	err   error
}

func (s skillLookupStub) FindByID(context.Context, string) (domain.Skill, error) {
	if s.err != nil {
		return domain.Skill{}, s.err
	}
	return s.skill, nil
}

type skillRunStoreStub struct {
	created []domain.SkillRun
	updated []domain.SkillRun
}

func (s *skillRunStoreStub) Create(_ context.Context, run domain.SkillRun) (domain.SkillRun, error) {
	run.CreatedAt = now()
	s.created = append(s.created, run)
	return run, nil
}

func (s *skillRunStoreStub) Update(_ context.Context, run domain.SkillRun) (domain.SkillRun, error) {
	s.updated = append(s.updated, run)
	return run, nil
}

func (s *skillRunStoreStub) FindByID(context.Context, string) (domain.SkillRun, error) {
	if len(s.updated) > 0 {
		run := s.updated[len(s.updated)-1]
		run.Skill = domain.SkillRunSkill{
			ID:         run.SkillID,
			Slug:       "echo-skill",
			Title:      "Echo Skill",
			Entrypoint: "echo",
		}
		run.Meta.HasOutput = len(run.Output) > 0
		run.Meta.HasError = run.ErrorMessage != nil
		run.Meta.InputKeysCount = len(run.Input)
		run.Meta.OutputKeysCount = len(run.Output)
		return run, nil
	}
	if len(s.created) > 0 {
		run := s.created[len(s.created)-1]
		run.Skill = domain.SkillRunSkill{
			ID:         run.SkillID,
			Slug:       "echo-skill",
			Title:      "Echo Skill",
			Entrypoint: "echo",
		}
		run.Meta.InputKeysCount = len(run.Input)
		return run, nil
	}
	return domain.SkillRun{}, nil
}

func (s *skillRunStoreStub) ListByUserID(context.Context, string) ([]domain.SkillRun, error) {
	return []domain.SkillRun{}, nil
}

func (s *skillRunStoreStub) ListAll(context.Context) ([]domain.SkillRun, error) {
	return []domain.SkillRun{}, nil
}

type registryStub struct {
	executor appRuntime.Executor
	err      error
}

func (s registryStub) Find(string) (appRuntime.Executor, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.executor, nil
}

type executorStub struct {
	output map[string]any
	err    error
}

func (e executorStub) Run(context.Context, map[string]any, map[string]any) (map[string]any, error) {
	if e.err != nil {
		return nil, e.err
	}
	return e.output, nil
}

func TestRuntimeServiceRunCompletesAndPersistsOutput(t *testing.T) {
	runs := &skillRunStoreStub{}
	service := NewRuntimeService(
		skillLookupStub{skill: domain.Skill{ID: "skill-1", Entrypoint: "echo", Config: map[string]any{"mode": "safe"}}},
		runs,
		registryStub{executor: executorStub{output: map[string]any{"result": "ok"}}},
	)

	run, err := service.Run(context.Background(), "user-1", RuntimeRunInput{
		SkillID: "skill-1",
		Input:   map[string]any{"text": "hello"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if run.Status != domain.SkillRunStatusCompleted {
		t.Fatalf("expected completed status, got %s", run.Status)
	}
	if run.Output["result"] != "ok" {
		t.Fatalf("expected output to be persisted, got %#v", run.Output)
	}
	if len(runs.created) != 1 || len(runs.updated) != 1 {
		t.Fatalf("expected create and update to be called once, got create=%d update=%d", len(runs.created), len(runs.updated))
	}
	if run.StartedAt == nil || run.FinishedAt == nil {
		t.Fatalf("expected timestamps to be present")
	}
}

func TestRuntimeServiceRunMarksFailures(t *testing.T) {
	runs := &skillRunStoreStub{}
	service := NewRuntimeService(
		skillLookupStub{skill: domain.Skill{ID: "skill-1", Entrypoint: "echo"}},
		runs,
		registryStub{executor: executorStub{err: errors.New("boom")}},
	)

	run, err := service.Run(context.Background(), "user-1", RuntimeRunInput{
		SkillID: "skill-1",
	})
	if err != nil {
		t.Fatalf("expected no top-level error, got %v", err)
	}
	if run.Status != domain.SkillRunStatusFailed {
		t.Fatalf("expected failed status, got %s", run.Status)
	}
	if run.ErrorMessage == nil || *run.ErrorMessage != "boom" {
		t.Fatalf("expected error message to be persisted, got %#v", run.ErrorMessage)
	}
}

func TestRuntimeServiceReturnsEntrypointError(t *testing.T) {
	service := NewRuntimeService(
		skillLookupStub{skill: domain.Skill{ID: "skill-1", Entrypoint: "missing"}},
		&skillRunStoreStub{},
		registryStub{err: errors.New("not found")},
	)

	_, err := service.Run(context.Background(), "user-1", RuntimeRunInput{SkillID: "skill-1"})
	if !errors.Is(err, ErrRuntimeEntrypointNotFound) {
		t.Fatalf("expected ErrRuntimeEntrypointNotFound, got %v", err)
	}
}

func TestNormalizeMapReturnsEmptyMapForNil(t *testing.T) {
	value := normalizeMap(nil)
	if len(value) != 0 {
		t.Fatalf("expected empty map, got %#v", value)
	}
}
