package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type ProfileRepository struct {
	pool *pgxpool.Pool
}

func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

// Create inserts a new profile row.  All values are bound via positional
// parameters ($N) — the pgx driver sends them as separate wire-protocol
// parameters, so no SQL injection is possible regardless of input content.
func (r *ProfileRepository) Create(ctx context.Context, profile domain.Profile) (domain.Profile, error) {
	query := `
		INSERT INTO profiles (user_id, first_name, last_name, bio, timezone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at, availability_status, is_public_profile,
			show_avatar, show_bio, show_location, show_skills, show_languages,
			show_availability_status, show_website_url, show_linkedin_url, show_github_url, show_twitter_url`

	if err := r.pool.QueryRow(
		ctx, query,
		profile.UserID,
		profile.FirstName,
		profile.LastName,
		profile.Bio,
		profile.Timezone,
	).Scan(
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.AvailabilityStatus,
		&profile.IsPublicProfile,
		&profile.ShowAvatar,
		&profile.ShowBio,
		&profile.ShowLocation,
		&profile.ShowSkills,
		&profile.ShowLanguages,
		&profile.ShowAvailability,
		&profile.ShowWebsiteURL,
		&profile.ShowLinkedInURL,
		&profile.ShowGitHubURL,
		&profile.ShowTwitterURL,
	); err != nil {
		return domain.Profile{}, fmt.Errorf("create profile: %w", err)
	}
	return profile, nil
}

// Update overwrites all editable profile fields.  Every column value arrives
// as a bound parameter ($N); the query text itself never contains user data.
func (r *ProfileRepository) Update(ctx context.Context, profile domain.Profile) (domain.Profile, error) {
	query := `
		UPDATE profiles
		SET
			first_name          = $2,
			last_name           = $3,
			bio                 = $4,
			timezone            = $5,
			username            = $6,
			avatar_url          = $7,
			location            = $8,
			skills              = $9,
			languages           = $10,
			website_url         = $11,
			linkedin_url        = $12,
			github_url          = $13,
			twitter_url         = $14,
			availability_status = $15,
			is_public_profile   = $16,
			show_avatar         = $17,
			show_bio            = $18,
			show_location       = $19,
			show_skills         = $20,
			show_languages      = $21,
			show_availability_status = $22,
			show_website_url    = $23,
			show_linkedin_url   = $24,
			show_github_url     = $25,
			show_twitter_url    = $26,
			updated_at          = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
		RETURNING created_at, updated_at`

	if err := r.pool.QueryRow(
		ctx, query,
		profile.UserID,             // $1
		profile.FirstName,          // $2
		profile.LastName,           // $3
		profile.Bio,                // $4
		profile.Timezone,           // $5
		profile.Username,           // $6
		profile.AvatarURL,          // $7
		profile.Location,           // $8
		profile.Skills,             // $9  — pgx encodes []string as TEXT[]
		profile.Languages,          // $10 — pgx encodes []string as TEXT[]
		profile.WebsiteURL,         // $11
		profile.LinkedInURL,        // $12
		profile.GitHubURL,          // $13
		profile.TwitterURL,         // $14
		profile.AvailabilityStatus, // $15
		profile.IsPublicProfile,    // $16
		profile.ShowAvatar,         // $17
		profile.ShowBio,            // $18
		profile.ShowLocation,       // $19
		profile.ShowSkills,         // $20
		profile.ShowLanguages,      // $21
		profile.ShowAvailability,   // $22
		profile.ShowWebsiteURL,     // $23
		profile.ShowLinkedInURL,    // $24
		profile.ShowGitHubURL,      // $25
		profile.ShowTwitterURL,     // $26
	).Scan(
		&profile.CreatedAt,
		&profile.UpdatedAt,
	); err != nil {
		return domain.Profile{}, fmt.Errorf("update profile: %w", err)
	}
	return profile, nil
}

// FindByUserID loads the full profile for a given user.  The user_id is
// passed as a bound parameter; the JOIN is on a trusted UUID column.
func (r *ProfileRepository) FindByUserID(ctx context.Context, userID string) (domain.Profile, error) {
	query := `
		SELECT
			p.user_id,
			u.role,
			p.first_name,
			p.last_name,
			p.bio,
			p.timezone,
			p.username,
			p.avatar_url,
			p.location,
			p.skills,
			p.languages,
			p.website_url,
			p.linkedin_url,
			p.github_url,
			p.twitter_url,
			p.availability_status,
			p.is_public_profile,
			p.show_avatar,
			p.show_bio,
			p.show_location,
			p.show_skills,
			p.show_languages,
			p.show_availability_status,
			p.show_website_url,
			p.show_linkedin_url,
			p.show_github_url,
			p.show_twitter_url,
			p.created_at,
			p.updated_at,
			p.deleted_at
		FROM profiles p
		JOIN users u ON u.id = p.user_id
		WHERE p.user_id = $1 AND p.deleted_at IS NULL`

	var profile domain.Profile
	if err := r.pool.QueryRow(ctx, query, userID).Scan(
		&profile.UserID,
		&profile.Role,
		&profile.FirstName,
		&profile.LastName,
		&profile.Bio,
		&profile.Timezone,
		&profile.Username,
		&profile.AvatarURL,
		&profile.Location,
		&profile.Skills,
		&profile.Languages,
		&profile.WebsiteURL,
		&profile.LinkedInURL,
		&profile.GitHubURL,
		&profile.TwitterURL,
		&profile.AvailabilityStatus,
		&profile.IsPublicProfile,
		&profile.ShowAvatar,
		&profile.ShowBio,
		&profile.ShowLocation,
		&profile.ShowSkills,
		&profile.ShowLanguages,
		&profile.ShowAvailability,
		&profile.ShowWebsiteURL,
		&profile.ShowLinkedInURL,
		&profile.ShowGitHubURL,
		&profile.ShowTwitterURL,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.DeletedAt,
	); err != nil {
		return domain.Profile{}, fmt.Errorf("find profile: %w", err)
	}
	return profile, nil
}

func (r *ProfileRepository) SoftDeleteByUserID(ctx context.Context, userID string) error {
	query := `
		UPDATE profiles
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL`

	if _, err := r.pool.Exec(ctx, query, userID); err != nil {
		return fmt.Errorf("soft delete profile: %w", err)
	}
	return nil
}
