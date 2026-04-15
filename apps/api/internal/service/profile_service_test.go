package service

import (
	"context"
	"testing"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type profileRepoStub struct {
	profile domain.Profile
}

func (s profileRepoStub) Create(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	return profile, nil
}

func (s profileRepoStub) Update(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	return profile, nil
}

func (s profileRepoStub) FindByUserID(_ context.Context, _ string) (domain.Profile, error) {
	return s.profile, nil
}

func TestProfileServiceFindByUserIDReturnsProfile(t *testing.T) {
	bio := "Go developer"
	svc := NewProfileService(profileRepoStub{
		profile: domain.Profile{
			UserID:    "user-1",
			FirstName: "Ada",
			LastName:  "Lovelace",
			Bio:       &bio,
			Timezone:  "UTC",
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
