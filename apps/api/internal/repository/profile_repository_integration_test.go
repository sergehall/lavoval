package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

func TestProfileRepositoryUpdatePersistsEditableFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := openIntegrationTestPool(t)
	repo := NewProfileRepository(pool)

	userID := uuid.NewString()
	createdAt := seedProfileFixture(t, ctx, pool, userID)

	username := "sergehall_integration"
	avatarURL := "https://avatars.githubusercontent.com/u/60080971?v=4"
	location := "Los Angeles"
	bio := "Full-stack engineer building scalable products."
	websiteURL := "https://sergioartg.com"
	linkedinURL := "https://www.linkedin.com/in/sergehall"
	githubURL := "https://github.com/sergehall"
	twitterURL := "https://x.com/sergehall"

	updated, err := repo.Update(ctx, domain.Profile{
		UserID:             userID,
		FirstName:          "Serge",
		LastName:           "Hall",
		Bio:                &bio,
		Timezone:           "America/Los_Angeles",
		Username:           &username,
		AvatarURL:          &avatarURL,
		Location:           &location,
		Skills:             []string{"TypeScript", "Go", "Next.js"},
		Languages:          []string{"en", "be"},
		WebsiteURL:         &websiteURL,
		LinkedInURL:        &linkedinURL,
		GitHubURL:          &githubURL,
		TwitterURL:         &twitterURL,
		AvailabilityStatus: domain.AvailabilityLimited,
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
	})
	if err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if !updated.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected createdAt to stay unchanged, got %v want %v", updated.CreatedAt, createdAt)
	}
	if !updated.UpdatedAt.After(createdAt) {
		t.Fatalf("expected updatedAt to move forward, got %v <= %v", updated.UpdatedAt, createdAt)
	}

	stored, err := repo.FindByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("expected stored profile to load, got %v", err)
	}

	if stored.FirstName != "Serge" || stored.LastName != "Hall" {
		t.Fatalf("expected stored name Serge Hall, got %s %s", stored.FirstName, stored.LastName)
	}
	if stored.Username == nil || *stored.Username != username {
		t.Fatalf("expected username %q, got %v", username, stored.Username)
	}
	if stored.Bio == nil || *stored.Bio != bio {
		t.Fatalf("expected bio %q, got %v", bio, stored.Bio)
	}
	if stored.Location == nil || *stored.Location != location {
		t.Fatalf("expected location %q, got %v", location, stored.Location)
	}
	if len(stored.Skills) != 3 || stored.Skills[2] != "Next.js" {
		t.Fatalf("expected skills to persist, got %v", stored.Skills)
	}
	if len(stored.Languages) != 2 || stored.Languages[1] != "be" {
		t.Fatalf("expected languages to persist, got %v", stored.Languages)
	}
	if stored.AvailabilityStatus != domain.AvailabilityLimited {
		t.Fatalf("expected availability limited, got %s", stored.AvailabilityStatus)
	}
	if !stored.IsPublicProfile {
		t.Fatalf("expected public profile flag to persist")
	}
	if stored.ShowLocation {
		t.Fatalf("expected show location to persist as false")
	}
	if stored.ShowLanguages {
		t.Fatalf("expected show languages to persist as false")
	}
	if !stored.ShowAvatar || !stored.ShowWebsiteURL || !stored.ShowGitHubURL {
		t.Fatalf("expected selected public visibility flags to persist, got %+v", stored)
	}
	if stored.WebsiteURL == nil || *stored.WebsiteURL != websiteURL {
		t.Fatalf("expected website URL %q, got %v", websiteURL, stored.WebsiteURL)
	}
	if stored.LinkedInURL == nil || *stored.LinkedInURL != linkedinURL {
		t.Fatalf("expected LinkedIn URL %q, got %v", linkedinURL, stored.LinkedInURL)
	}
	if stored.GitHubURL == nil || *stored.GitHubURL != githubURL {
		t.Fatalf("expected GitHub URL %q, got %v", githubURL, stored.GitHubURL)
	}
	if stored.TwitterURL == nil || *stored.TwitterURL != twitterURL {
		t.Fatalf("expected Twitter URL %q, got %v", twitterURL, stored.TwitterURL)
	}
}

func openIntegrationTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://codex:codex@localhost:5432/lavoval?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Skipf("skipping repository integration test: could not create pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skipping repository integration test: database unavailable: %v", err)
	}

	t.Cleanup(pool.Close)
	return pool
}

func seedProfileFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userID string) time.Time {
	t.Helper()

	_, err := pool.Exec(
		ctx,
		`INSERT INTO users (id, email, password_hash, role, status)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID,
		userID+"@example.com",
		"hashed-password",
		domain.RoleUser,
		domain.AccountStatusActive,
	)
	if err != nil {
		t.Fatalf("expected seed user insert to succeed, got %v", err)
	}

	var createdAt time.Time
	if err := pool.QueryRow(
		ctx,
		`INSERT INTO profiles (user_id, first_name, last_name, bio, timezone)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING created_at`,
		userID,
		"Initial",
		"Profile",
		nil,
		"UTC",
	).Scan(&createdAt); err != nil {
		t.Fatalf("expected seed profile insert to succeed, got %v", err)
	}

	t.Cleanup(func() {
		if _, cleanupErr := pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID); cleanupErr != nil {
			t.Fatalf("expected cleanup to delete seeded user, got %v", cleanupErr)
		}
	})

	return createdAt
}
