package service

import (
	"context"
	"fmt"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

type ProfileService struct {
	profiles repository.ProfileStore
}

type UpdateProfileInput struct {
	FirstName string  `json:"firstName" validate:"required,min=2"`
	LastName  string  `json:"lastName" validate:"required,min=2"`
	Bio       *string `json:"bio"`
	Timezone  string  `json:"timezone" validate:"required,min=2"`
}

func NewProfileService(profiles repository.ProfileStore) *ProfileService {
	return &ProfileService{profiles: profiles}
}

func (s *ProfileService) FindByUserID(ctx context.Context, userID string) (domain.Profile, error) {
	profile, err := s.profiles.FindByUserID(ctx, userID)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("load profile: %w", err)
	}
	return profile, nil
}

func (s *ProfileService) Update(ctx context.Context, userID string, input UpdateProfileInput) (domain.Profile, error) {
	profile := domain.Profile{
		UserID:    userID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Bio:       input.Bio,
		Timezone:  input.Timezone,
	}

	updatedProfile, err := s.profiles.Update(ctx, profile)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("update profile: %w", err)
	}
	return updatedProfile, nil
}
