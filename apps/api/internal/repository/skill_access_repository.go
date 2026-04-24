package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type SkillAccessRepository struct {
	pool *pgxpool.Pool
}

func NewSkillAccessRepository(pool *pgxpool.Pool) *SkillAccessRepository {
	return &SkillAccessRepository{pool: pool}
}

func (r *SkillAccessRepository) Create(ctx context.Context, access domain.SkillAccess) (domain.SkillAccess, error) {
	query := `
		INSERT INTO lavoval_skill_access (id, skill_id, user_id, access_type, granted_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (skill_id, user_id) DO UPDATE
		  SET access_type = EXCLUDED.access_type,
		      granted_by  = EXCLUDED.granted_by,
		      expires_at  = EXCLUDED.expires_at
		RETURNING created_at`
	if err := r.pool.QueryRow(ctx, query,
		access.ID, access.SkillID, access.UserID, access.AccessType,
		access.GrantedBy, access.ExpiresAt,
	).Scan(&access.CreatedAt); err != nil {
		return domain.SkillAccess{}, fmt.Errorf("upsert skill access: %w", err)
	}
	return access, nil
}

func (r *SkillAccessRepository) FindBySkillAndUser(ctx context.Context, skillID, userID string) (*domain.SkillAccess, error) {
	query := `
		SELECT id, skill_id, user_id, access_type, granted_by, expires_at, created_at
		FROM lavoval_skill_access
		WHERE skill_id = $1 AND user_id = $2`

	var a domain.SkillAccess
	err := r.pool.QueryRow(ctx, query, skillID, userID).Scan(
		&a.ID, &a.SkillID, &a.UserID, &a.AccessType, &a.GrantedBy, &a.ExpiresAt, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find skill access: %w", err)
	}
	return &a, nil
}

func (r *SkillAccessRepository) ListBySkillID(ctx context.Context, skillID string) ([]domain.SkillAccess, error) {
	return r.list(ctx, `WHERE skill_id = $1`, skillID)
}

func (r *SkillAccessRepository) ListByUserID(ctx context.Context, userID string) ([]domain.SkillAccess, error) {
	return r.list(ctx, `WHERE user_id = $1`, userID)
}

func (r *SkillAccessRepository) list(ctx context.Context, where string, arg any) ([]domain.SkillAccess, error) {
	query := `
		SELECT id, skill_id, user_id, access_type, granted_by, expires_at, created_at
		FROM lavoval_skill_access ` + where + ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("list skill access: %w", err)
	}
	defer rows.Close()

	items := make([]domain.SkillAccess, 0)
	for rows.Next() {
		var a domain.SkillAccess
		if err := rows.Scan(&a.ID, &a.SkillID, &a.UserID, &a.AccessType, &a.GrantedBy, &a.ExpiresAt, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan skill access: %w", err)
		}
		items = append(items, a)
	}
	return items, rows.Err()
}
