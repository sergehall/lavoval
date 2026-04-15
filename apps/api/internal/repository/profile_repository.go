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

func (r *ProfileRepository) Create(ctx context.Context, profile domain.Profile) (domain.Profile, error) {
	query := `
		INSERT INTO profiles (user_id, first_name, last_name, bio, timezone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`
	if err := r.pool.QueryRow(ctx, query, profile.UserID, profile.FirstName, profile.LastName, profile.Bio, profile.Timezone).Scan(&profile.CreatedAt, &profile.UpdatedAt); err != nil {
		return domain.Profile{}, fmt.Errorf("create profile: %w", err)
	}
	return profile, nil
}

func (r *ProfileRepository) Update(ctx context.Context, profile domain.Profile) (domain.Profile, error) {
	query := `
		UPDATE profiles
		SET first_name = $2, last_name = $3, bio = $4, timezone = $5, updated_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
		RETURNING created_at, updated_at`
	if err := r.pool.QueryRow(ctx, query, profile.UserID, profile.FirstName, profile.LastName, profile.Bio, profile.Timezone).Scan(&profile.CreatedAt, &profile.UpdatedAt); err != nil {
		return domain.Profile{}, fmt.Errorf("update profile: %w", err)
	}
	return profile, nil
}

func (r *ProfileRepository) FindByUserID(ctx context.Context, userID string) (domain.Profile, error) {
	query := `
		SELECT p.user_id, u.role, p.first_name, p.last_name, p.bio, p.timezone, p.created_at, p.updated_at, p.deleted_at
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
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.DeletedAt,
	); err != nil {
		return domain.Profile{}, fmt.Errorf("find profile: %w", err)
	}
	return profile, nil
}
