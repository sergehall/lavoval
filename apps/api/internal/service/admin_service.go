package service

import (
	"context"
	"fmt"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

type AdminService struct {
	users  repository.UserStore
	skills repository.SkillStore
}

func NewAdminService(users repository.UserStore, skills repository.SkillStore) *AdminService {
	return &AdminService{users: users, skills: skills}
}

func (s *AdminService) ListUsers(ctx context.Context) ([]domain.User, error) {
	users, err := s.users.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (s *AdminService) ListSkills(ctx context.Context) ([]domain.Skill, error) {
	skills, err := s.skills.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list admin skills: %w", err)
	}
	return skills, nil
}
