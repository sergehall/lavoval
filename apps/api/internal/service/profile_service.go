package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

var allowedAvatarURLHosts = map[string]struct{}{
	"avatars.githubusercontent.com": {},
	"secure.gravatar.com":           {},
	"www.gravatar.com":              {},
	"lh3.googleusercontent.com":     {},
}

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

	// AvatarURL is hardened separately from general profile links: only
	// HTTPS URLs on the avatar allowlist are accepted, and credentials are
	// never allowed in the URL authority.
	AvatarURL   *string `json:"avatarUrl"   validate:"omitempty,max=2048"`
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
	IsPublicProfile        bool `json:"isPublicProfile"`
	ShowAvatar             bool `json:"showAvatar"`
	ShowBio                bool `json:"showBio"`
	ShowLocation           bool `json:"showLocation"`
	ShowSkills             bool `json:"showSkills"`
	ShowLanguages          bool `json:"showLanguages"`
	ShowAvailabilityStatus bool `json:"showAvailabilityStatus"`
	ShowWebsiteURL         bool `json:"showWebsiteUrl"`
	ShowLinkedInURL        bool `json:"showLinkedinUrl"`
	ShowGitHubURL          bool `json:"showGithubUrl"`
	ShowTwitterURL         bool `json:"showTwitterUrl"`
}

func NewProfileService(profiles repository.ProfileStore) *ProfileService {
	return &ProfileService{profiles: profiles}
}

func IsSafeAvatarURL(value string) bool {
	if value == "" || len(value) > 2048 {
		return false
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	if parsed.Scheme != "https" || host == "" || parsed.Port() != "" {
		return false
	}

	if parsed.User != nil {
		return false
	}

	_, ok := allowedAvatarURLHosts[host]
	return ok
}

func safeAvatarURL(value *string) *string {
	if value == nil || !IsSafeAvatarURL(*value) {
		return nil
	}
	return value
}

func (s *ProfileService) FindByUserID(ctx context.Context, userID string) (domain.Profile, error) {
	profile, err := s.profiles.FindByUserID(ctx, userID)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("load profile: %w", err)
	}
	profile.AvatarURL = safeAvatarURL(profile.AvatarURL)
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
		AvatarURL:          safeAvatarURL(input.AvatarURL),
		Location:           input.Location,
		Skills:             input.Skills,
		Languages:          input.Languages,
		WebsiteURL:         input.WebsiteURL,
		LinkedInURL:        input.LinkedInURL,
		GitHubURL:          input.GitHubURL,
		TwitterURL:         input.TwitterURL,
		AvailabilityStatus: availStatus,
		IsPublicProfile:    input.IsPublicProfile,
		ShowAvatar:         input.ShowAvatar,
		ShowBio:            input.ShowBio,
		ShowLocation:       input.ShowLocation,
		ShowSkills:         input.ShowSkills,
		ShowLanguages:      input.ShowLanguages,
		ShowAvailability:   input.ShowAvailabilityStatus,
		ShowWebsiteURL:     input.ShowWebsiteURL,
		ShowLinkedInURL:    input.ShowLinkedInURL,
		ShowGitHubURL:      input.ShowGitHubURL,
		ShowTwitterURL:     input.ShowTwitterURL,
	}

	updatedProfile, err := s.profiles.Update(ctx, profile)
	if err != nil {
		return domain.Profile{}, fmt.Errorf("update profile: %w", err)
	}
	return updatedProfile, nil
}
