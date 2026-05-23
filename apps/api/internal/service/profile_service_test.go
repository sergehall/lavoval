package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type profileRepoStub struct {
	profile       domain.Profile
	updateErr     error
	updatedCalled bool
	updatedInput  domain.Profile
}

func (s profileRepoStub) Create(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	return profile, nil
}

func (s profileRepoStub) Update(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	if s.updateErr != nil {
		return domain.Profile{}, s.updateErr
	}
	return profile, nil
}

func (s profileRepoStub) FindByUserID(_ context.Context, _ string) (domain.Profile, error) {
	return s.profile, nil
}

func (s profileRepoStub) SoftDeleteByUserID(_ context.Context, _ string) error {
	return nil
}

func TestProfileServiceFindByUserIDReturnsProfile(t *testing.T) {
	bio := "Go developer"
	avatarURL := "https://user:pass@avatars.githubusercontent.com/u/60080971"
	svc := NewProfileService(profileRepoStub{
		profile: domain.Profile{
			UserID:    "user-1",
			FirstName: "Ada",
			LastName:  "Lovelace",
			Bio:       &bio,
			Timezone:  "UTC",
			AvatarURL: &avatarURL,
		},
	})

	profile, err := svc.FindByUserID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if profile.FirstName != "Ada" {
		t.Fatalf("expected FirstName Ada, got %s", profile.FirstName)
	}
	if profile.Bio == nil || *profile.Bio != "Go developer" {
		t.Fatalf("expected bio to be preserved, got %v", profile.Bio)
	}
	if profile.AvatarURL != nil {
		t.Fatalf("expected unsafe stored avatar URL to be cleared, got %v", profile.AvatarURL)
	}
}

func TestProfileServiceUpdateBuildsProfileFromInput(t *testing.T) {
	svc := NewProfileService(profileRepoStub{})

	bio := "Mathematician"
	updated, err := svc.Update(context.Background(), "user-1", UpdateProfileInput{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Bio:       &bio,
		Timezone:  "Europe/London",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.UserID != "user-1" {
		t.Fatalf("expected userID user-1, got %s", updated.UserID)
	}
	if updated.FirstName != "Ada" {
		t.Fatalf("expected FirstName Ada, got %s", updated.FirstName)
	}
	if updated.Timezone != "Europe/London" {
		t.Fatalf("expected timezone Europe/London, got %s", updated.Timezone)
	}
	if updated.Bio == nil || *updated.Bio != "Mathematician" {
		t.Fatalf("expected bio Mathematician, got %v", updated.Bio)
	}
}

func TestProfileServiceUpdateNilBioIsAllowed(t *testing.T) {
	svc := NewProfileService(profileRepoStub{})

	updated, err := svc.Update(context.Background(), "user-2", UpdateProfileInput{
		FirstName: "Charles",
		LastName:  "Babbage",
		Bio:       nil,
		Timezone:  "UTC",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Bio != nil {
		t.Fatalf("expected nil bio, got %v", updated.Bio)
	}
}

type profileRepoCaptureStub struct {
	updateErr    error
	updatedInput domain.Profile
}

func (s *profileRepoCaptureStub) Create(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	return profile, nil
}

func (s *profileRepoCaptureStub) Update(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	s.updatedInput = profile
	if s.updateErr != nil {
		return domain.Profile{}, s.updateErr
	}
	return profile, nil
}

func (s *profileRepoCaptureStub) FindByUserID(_ context.Context, _ string) (domain.Profile, error) {
	return domain.Profile{}, nil
}

func (s *profileRepoCaptureStub) SoftDeleteByUserID(_ context.Context, _ string) error {
	return nil
}

func TestProfileServiceUpdatePersistsMarketplaceFields(t *testing.T) {
	username := "sergehall"
	avatarURL := "https://avatars.githubusercontent.com/u/60080971?v=4"
	location := "Los Angeles"
	websiteURL := "https://sergioartg.com"
	linkedinURL := "https://linkedin.com/in/sergehall"
	githubURL := "https://github.com/sergehall"
	twitterURL := "https://x.com/sergehall"

	repo := &profileRepoCaptureStub{}
	svc := NewProfileService(repo)

	updated, err := svc.Update(context.Background(), "user-3", UpdateProfileInput{
		FirstName:              "Serge",
		LastName:               "Hall",
		Timezone:               "America/Los_Angeles",
		Username:               &username,
		AvatarURL:              &avatarURL,
		Location:               &location,
		Skills:                 []string{"TypeScript", "Go"},
		Languages:              []string{"en", "be"},
		WebsiteURL:             &websiteURL,
		LinkedInURL:            &linkedinURL,
		GitHubURL:              &githubURL,
		TwitterURL:             &twitterURL,
		AvailabilityStatus:     domain.AvailabilityLimited,
		IsPublicProfile:        true,
		ShowAvatar:             true,
		ShowBio:                true,
		ShowLocation:           true,
		ShowSkills:             true,
		ShowLanguages:          true,
		ShowAvailabilityStatus: true,
		ShowWebsiteURL:         true,
		ShowLinkedInURL:        false,
		ShowGitHubURL:          true,
		ShowTwitterURL:         false,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.UserID != "user-3" {
		t.Fatalf("expected userID user-3, got %s", updated.UserID)
	}
	if repo.updatedInput.Username == nil || *repo.updatedInput.Username != username {
		t.Fatalf("expected username %q, got %v", username, repo.updatedInput.Username)
	}
	if len(repo.updatedInput.Skills) != 2 || repo.updatedInput.Skills[0] != "TypeScript" {
		t.Fatalf("expected skills to be forwarded, got %v", repo.updatedInput.Skills)
	}
	if len(repo.updatedInput.Languages) != 2 || repo.updatedInput.Languages[1] != "be" {
		t.Fatalf("expected languages to be forwarded, got %v", repo.updatedInput.Languages)
	}
	if repo.updatedInput.AvailabilityStatus != domain.AvailabilityLimited {
		t.Fatalf("expected availability status limited, got %s", repo.updatedInput.AvailabilityStatus)
	}
	if !repo.updatedInput.IsPublicProfile {
		t.Fatalf("expected public profile flag to be true")
	}
	if !repo.updatedInput.ShowAvatar || !repo.updatedInput.ShowSkills || !repo.updatedInput.ShowWebsiteURL {
		t.Fatalf("expected public visibility fields to be forwarded, got %+v", repo.updatedInput)
	}
	if repo.updatedInput.ShowLinkedInURL {
		t.Fatalf("expected showLinkedInUrl to remain false")
	}
}

func TestProfileServiceUpdateClearsUnsafeAvatarURL(t *testing.T) {
	unsafeAvatarURL := "https://user:pass@avatars.githubusercontent.com/u/60080971"
	repo := &profileRepoCaptureStub{}
	svc := NewProfileService(repo)

	updated, err := svc.Update(context.Background(), "user-4", UpdateProfileInput{
		FirstName: "Serge",
		LastName:  "Hall",
		Timezone:  "UTC",
		AvatarURL: &unsafeAvatarURL,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.AvatarURL != nil {
		t.Fatalf("expected unsafe avatar URL to be cleared, got %v", updated.AvatarURL)
	}
	if repo.updatedInput.AvatarURL != nil {
		t.Fatalf("expected repository input avatar URL to be cleared, got %v", repo.updatedInput.AvatarURL)
	}
}

func TestIsSafeAvatarURL(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "github avatar",
			value: "https://avatars.githubusercontent.com/u/60080971?v=4",
			want:  true,
		},
		{
			name:  "gravatar",
			value: "https://secure.gravatar.com/avatar/hash?s=96",
			want:  true,
		},
		{
			name:  "http rejected",
			value: "http://avatars.githubusercontent.com/u/60080971",
			want:  false,
		},
		{
			name:  "credentials rejected",
			value: "https://user:pass@avatars.githubusercontent.com/u/60080971",
			want:  false,
		},
		{
			name:  "custom port rejected",
			value: "https://avatars.githubusercontent.com:8443/u/60080971",
			want:  false,
		},
		{
			name:  "unlisted host rejected",
			value: "https://example.com/avatar.png",
			want:  false,
		},
		{
			name:  "host suffix rejected",
			value: "https://avatars.githubusercontent.com.evil.test/avatar.png",
			want:  false,
		},
		{
			name:  "data URL rejected",
			value: "data:image/png;base64,abcd",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSafeAvatarURL(tt.value); got != tt.want {
				t.Fatalf("IsSafeAvatarURL(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestProfileServiceUpdateWrapsRepositoryError(t *testing.T) {
	repo := &profileRepoCaptureStub{updateErr: errors.New("write failed")}
	svc := NewProfileService(repo)

	_, err := svc.Update(context.Background(), "user-4", UpdateProfileInput{
		FirstName: "Serge",
		LastName:  "Hall",
		Timezone:  "UTC",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "update profile: write failed") {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}
