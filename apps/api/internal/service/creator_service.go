package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

var ErrPublicProfileNotFound = errors.New("public profile not found")

type CreatorService struct {
	profiles repository.ProfileStore
	skills   repository.SkillStore
}

func NewCreatorService(profiles repository.ProfileStore, skills repository.SkillStore) *CreatorService {
	return &CreatorService{profiles: profiles, skills: skills}
}

func (s *CreatorService) FindPublicByUserID(ctx context.Context, userID string) (domain.PublicProfile, error) {
	profile, err := s.profiles.FindByUserID(ctx, userID)
	if err != nil {
		return domain.PublicProfile{}, fmt.Errorf("load profile: %w", err)
	}
	if !profile.IsPublicProfile {
		return domain.PublicProfile{}, ErrPublicProfileNotFound
	}

	skills, err := s.skills.ListByCreatorID(ctx, userID)
	if err != nil {
		return domain.PublicProfile{}, fmt.Errorf("load creator skills: %w", err)
	}

	publicSkills := make([]domain.PublicProfileSkill, 0, len(skills))
	for _, skill := range skills {
		if skill.Status != domain.SkillStatusPublished || skill.Visibility != domain.VisibilityPublic {
			continue
		}

		publicSkills = append(publicSkills, domain.PublicProfileSkill{
			ID:         skill.ID,
			Slug:       skill.Slug,
			Title:      skill.Title,
			Summary:    skill.Summary,
			SkillType:  skill.SkillType,
			Difficulty: skill.Difficulty,
			AvgRating:  skill.AvgRating,
			RunsCount:  skill.RunsCount,
			UpdatedAt:  skill.UpdatedAt,
		})
	}

	return domain.PublicProfile{
		UserID:             profile.UserID,
		FirstName:          profile.FirstName,
		LastName:           profile.LastName,
		FullName:           profile.FirstName + " " + profile.LastName,
		Username:           profile.Username,
		AvatarURL:          visibleString(safeAvatarURL(profile.AvatarURL), profile.ShowAvatar),
		Bio:                visibleString(profile.Bio, profile.ShowBio),
		Location:           visibleString(profile.Location, profile.ShowLocation),
		Skills:             visibleStrings(profile.Skills, profile.ShowSkills),
		Languages:          visibleStrings(profile.Languages, profile.ShowLanguages),
		WebsiteURL:         visibleString(profile.WebsiteURL, profile.ShowWebsiteURL),
		LinkedInURL:        visibleString(profile.LinkedInURL, profile.ShowLinkedInURL),
		GitHubURL:          visibleString(profile.GitHubURL, profile.ShowGitHubURL),
		TwitterURL:         visibleString(profile.TwitterURL, profile.ShowTwitterURL),
		AvailabilityStatus: visibleAvailability(profile.AvailabilityStatus, profile.ShowAvailability),
		PublicSkills:       publicSkills,
	}, nil
}

func visibleString(value *string, visible bool) *string {
	if !visible {
		return nil
	}
	return value
}

func visibleStrings(values []string, visible bool) []string {
	if !visible || len(values) == 0 {
		return nil
	}
	return values
}

func visibleAvailability(value domain.AvailabilityStatus, visible bool) *domain.AvailabilityStatus {
	if !visible {
		return nil
	}
	status := value
	return &status
}
