package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type creatorProfileRepoStub struct {
	profile domain.Profile
	err     error
}

func (s creatorProfileRepoStub) Create(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	return profile, nil
}

func (s creatorProfileRepoStub) Update(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	return profile, nil
}

func (s creatorProfileRepoStub) FindByUserID(_ context.Context, _ string) (domain.Profile, error) {
	if s.err != nil {
		return domain.Profile{}, s.err
	}
	return s.profile, nil
}

func (s creatorProfileRepoStub) SoftDeleteByUserID(_ context.Context, _ string) error {
	return nil
}

type creatorSkillRepoStub struct {
	skills []domain.Skill
	err    error
}

func (s creatorSkillRepoStub) ListPublished(context.Context, domain.SkillFilter) ([]domain.Skill, error) {
	return nil, nil
}

func (s creatorSkillRepoStub) ListAll(context.Context) ([]domain.Skill, error) {
	return nil, nil
}

func (s creatorSkillRepoStub) ListByCreatorID(_ context.Context, _ string) ([]domain.Skill, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.skills, nil
}

func (s creatorSkillRepoStub) FindByID(context.Context, string) (domain.Skill, error) {
	return domain.Skill{}, nil
}

func (s creatorSkillRepoStub) Create(context.Context, domain.Skill) (domain.Skill, error) {
	return domain.Skill{}, nil
}

func (s creatorSkillRepoStub) Update(context.Context, domain.Skill) (domain.Skill, error) {
	return domain.Skill{}, nil
}

func (s creatorSkillRepoStub) UpdateGovernance(context.Context, string, string, domain.SkillStatus, *string, bool, bool) (domain.Skill, error) {
	return domain.Skill{}, nil
}

func (s creatorSkillRepoStub) UpdatePricing(context.Context, string, int, string, domain.SkillAccessType) (domain.Skill, error) {
	return domain.Skill{}, nil
}

func (s creatorSkillRepoStub) GetStats(context.Context) (domain.AdminSkillStats, error) {
	return domain.AdminSkillStats{}, nil
}

func (s creatorSkillRepoStub) SoftDelete(context.Context, string) error {
	return nil
}

func TestCreatorServiceFindPublicByUserIDReturnsVisibleFieldsAndPublicSkills(t *testing.T) {
	now := time.Now().UTC()
	username := "sergehall"
	avatarURL := "https://example.com/avatar.png"
	bio := "Builds systems."
	location := "Los Angeles"
	websiteURL := "https://sergioartg.com"
	githubURL := "https://github.com/sergehall"

	svc := NewCreatorService(
		creatorProfileRepoStub{
			profile: domain.Profile{
				UserID:             "user-1",
				FirstName:          "Serge",
				LastName:           "Hall",
				Username:           &username,
				AvatarURL:          &avatarURL,
				Bio:                &bio,
				Location:           &location,
				Skills:             []string{"TypeScript", "Go"},
				Languages:          []string{"en", "be"},
				WebsiteURL:         &websiteURL,
				GitHubURL:          &githubURL,
				AvailabilityStatus: domain.AvailabilityOpen,
				IsPublicProfile:    true,
				ShowAvatar:         true,
				ShowBio:            true,
				ShowLocation:       false,
				ShowSkills:         true,
				ShowLanguages:      false,
				ShowAvailability:   true,
				ShowWebsiteURL:     true,
				ShowLinkedInURL:    false,
				ShowGitHubURL:      true,
				ShowTwitterURL:     false,
			},
		},
		creatorSkillRepoStub{
			skills: []domain.Skill{
				{
					ID:         "skill-1",
					Slug:       "typescript-contracts",
					Title:      "TypeScript Contracts",
					Summary:    "Contract-safe TypeScript delivery.",
					Status:     domain.SkillStatusPublished,
					Visibility: domain.VisibilityPublic,
					SkillType:  "workflow",
					Difficulty: "senior",
					AvgRating:  4.8,
					RunsCount:  12,
					UpdatedAt:  now,
				},
				{
					ID:         "skill-2",
					Slug:       "private-draft",
					Title:      "Private Draft",
					Summary:    "Should stay hidden.",
					Status:     domain.SkillStatusDraft,
					Visibility: domain.VisibilityPrivate,
					UpdatedAt:  now,
				},
			},
		},
	)

	profile, err := svc.FindPublicByUserID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if profile.FullName != "Serge Hall" {
		t.Fatalf("expected full name Serge Hall, got %s", profile.FullName)
	}
	if profile.Location != nil {
		t.Fatalf("expected hidden location to be nil, got %v", profile.Location)
	}
	if profile.Languages != nil {
		t.Fatalf("expected hidden languages to be nil, got %v", profile.Languages)
	}
	if profile.AvailabilityStatus == nil || *profile.AvailabilityStatus != domain.AvailabilityOpen {
		t.Fatalf("expected visible availability status, got %v", profile.AvailabilityStatus)
	}
	if len(profile.PublicSkills) != 1 {
		t.Fatalf("expected exactly one public skill, got %d", len(profile.PublicSkills))
	}
	if profile.PublicSkills[0].Slug != "typescript-contracts" {
		t.Fatalf("expected public skill slug typescript-contracts, got %s", profile.PublicSkills[0].Slug)
	}
}

func TestCreatorServiceFindPublicByUserIDRejectsPrivateProfile(t *testing.T) {
	svc := NewCreatorService(
		creatorProfileRepoStub{profile: domain.Profile{IsPublicProfile: false}},
		creatorSkillRepoStub{},
	)

	_, err := svc.FindPublicByUserID(context.Background(), "user-2")
	if !errors.Is(err, ErrPublicProfileNotFound) {
		t.Fatalf("expected ErrPublicProfileNotFound, got %v", err)
	}
}
