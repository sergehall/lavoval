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

// UpdateProfileInput carries user-supplied values for profile update.
// Validation tags are enforced by go-playground/validator before the data
// reaches the repository.  All SQL parameters are positional ($N) in the
// repository, so there is no SQL-injection risk even without these checks —
// the tags are a second, application-layer defence.
type UpdateProfileInput struct {
	// Core identity
	FirstName string  `json:"firstName" validate:"required,min=2,max=100"`
	LastName  string  `json:"lastName"  validate:"required,min=2,max=100"`
	Bio       *string `json:"bio"       validate:"omitempty,max=500"`
	Timezone  string  `json:"timezone"  validate:"required,min=2,max=100"`

	// Public handle: letters, digits, dash, underscore, 3-30 chars.
	// The DB also enforces the character set via a CHECK constraint.
	Username *string `json:"username" validate:"omitempty,min=3,max=30"`

	// URLs: go-playground/validator's "url" tag rejects anything that is
	// not a well-formed absolute URL, preventing e.g. javascript: URIs.
	AvatarURL   *string `json:"avatarUrl"   validate:"omitempty,url,max=2048"`
	WebsiteURL  *string `json:"websiteUrl"  validate:"omitempty,url,max=2048"`
	LinkedInURL *string `json:"linkedinUrl" validate:"omitempty,url,max=2048"`
	GitHubURL   *string `json:"githubUrl"   validate:"omitempty,url,max=2048"`
	TwitterURL  *string `json:"twitterUrl"  validate:"omitempty,url,max=2048"`

	// Free-text location: 2-100 chars.
	Location *string `json:"location" validate:"omitempty,min=2,max=100"`

	// Array fields.  "max" validates the slice length; "dive" runs the
	// subsequent tags on every element.
	Skills    []string `json:"skills"    validate:"omitempty,max=20,dive,min=1,max=50"`
	Languages []string `json:"languages" validate:"omitempty,max=10,dive,min=2,max=10"`

	// Marketplace availability: closed enum.
	AvailabilityStatus domain.AvailabilityStatus `json:"availabilityStatus" validate:"omitempty,oneof=open limited closed"`

	// Profile visibility in public catalogue.
	IsPublicProfile bool `json:"isPublicProfile"`
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
	availStatus := input.AvailabilityStatus
	if availStatus == "" {
		availStatus = domain.AvailabilityOpen
	}

	profile := domain.Profile{
		UserID:             userID,
		FirstName:          input.FirstName,
		LastName:           input.LastName,
		Bio:                input.Bio,
		Timezone:           input.Timezone,
		Username:           input.Username,
		AvatarURL:          input.AvatarURL,
		Location:           input.Location,
		Skills:             input.Skills,
		Languages:          input.Languages,
		WebsiteURL:         input.WebsiteURL,
		LinkedInURL:        input.LinkedInURL,
		GitHubURL:          input.GitHubURL,
		TwitterURL:         input.TwitterURL,
		AvailabilityStatus: availStatus,
		IsPublicProfile:    input.IsPublicProfile,
	}

	updatedProfile, err := s.profiles.Update(ctx, profile)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("update profile: %w", err)
	}
	return updatedProfile, nil
}
