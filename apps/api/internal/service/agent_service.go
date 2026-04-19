package service

import (
	"context"
	"fmt"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

type AgentService struct {
	agents *repository.AgentRepository
}

func NewAgentService(agents *repository.AgentRepository) *AgentService {
	return &AgentService{agents: agents}
}

func (s *AgentService) List(ctx context.Context) ([]domain.Agent, error) {
	items, err := s.agents.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}
	return items, nil
}

func (s *AgentService) FindBySlug(ctx context.Context, slug string) (domain.Agent, error) {
	a, err := s.agents.FindBySlug(ctx, slug)
	if err != nil {
		return domain.Agent{}, fmt.Errorf("find agent: %w", err)
	}
	return a, nil
}

func (s *AgentService) RecommendedForSkill(ctx context.Context, skillID string) ([]domain.SkillAgentCompatibility, error) {
	items, err := s.agents.ListRecommendedForSkill(ctx, skillID)
	if err != nil {
		return nil, fmt.Errorf("recommended agents: %w", err)
	}
	return items, nil
}
