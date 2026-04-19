package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type SocialRepository struct {
	pool *pgxpool.Pool
}

func NewSocialRepository(pool *pgxpool.Pool) *SocialRepository {
	return &SocialRepository{pool: pool}
}

// ── Reviews ───────────────────────────────────────────────────────────────────

func (r *SocialRepository) ListReviews(ctx context.Context, skillID string) ([]domain.SkillReview, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT sr.id, sr.skill_id, sr.user_id, sr.run_id, sr.rating, sr.review_text, sr.created_at,
		       u.id, u.email, p.first_name, p.last_name
		FROM skill_reviews sr
		INNER JOIN users u ON u.id = sr.user_id
		INNER JOIN profiles p ON p.user_id = u.id AND p.deleted_at IS NULL
		WHERE sr.skill_id = $1
		ORDER BY sr.created_at DESC`, skillID)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	var out []domain.SkillReview
	for rows.Next() {
		var rv domain.SkillReview
		if err := rows.Scan(
			&rv.ID, &rv.SkillID, &rv.UserID, &rv.RunID, &rv.Rating, &rv.ReviewText, &rv.CreatedAt,
			&rv.Reviewer.ID, &rv.Reviewer.Email, &rv.Reviewer.FirstName, &rv.Reviewer.LastName,
		); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		out = append(out, rv)
	}
	return out, rows.Err()
}

func (r *SocialRepository) CreateReview(ctx context.Context, rv domain.SkillReview) (domain.SkillReview, error) {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO skill_reviews (id, skill_id, user_id, run_id, rating, review_text)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (skill_id, user_id) DO UPDATE
		  SET rating = EXCLUDED.rating, review_text = EXCLUDED.review_text`,
		rv.ID, rv.SkillID, rv.UserID, rv.RunID, rv.Rating, rv.ReviewText)
	if err != nil {
		return domain.SkillReview{}, fmt.Errorf("upsert review: %w", err)
	}
	return rv, nil
}

// ── Saved skills ─────────────────────────────────────────────────────────────

func (r *SocialRepository) SaveSkill(ctx context.Context, userID, skillID string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO saved_skills (user_id, skill_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, skillID)
	return err
}

func (r *SocialRepository) UnsaveSkill(ctx context.Context, userID, skillID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM saved_skills WHERE user_id = $1 AND skill_id = $2`, userID, skillID)
	return err
}

func (r *SocialRepository) IsSaved(ctx context.Context, userID, skillID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM saved_skills WHERE user_id = $1 AND skill_id = $2)`,
		userID, skillID).Scan(&exists)
	return exists, err
}

func (r *SocialRepository) ListSavedSkillIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT skill_id FROM saved_skills WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list saved skill ids: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ── Collections ───────────────────────────────────────────────────────────────

func (r *SocialRepository) ListCollections(ctx context.Context, ownerID string) ([]domain.Collection, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, owner_id, title, description, visibility, created_at, updated_at
		FROM collections
		WHERE owner_id = $1
		ORDER BY updated_at DESC`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	defer rows.Close()

	var out []domain.Collection
	for rows.Next() {
		var c domain.Collection
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.Title, &c.Description, &c.Visibility, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan collection: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *SocialRepository) FindCollection(ctx context.Context, id string) (domain.Collection, error) {
	var c domain.Collection
	err := r.pool.QueryRow(ctx, `
		SELECT id, owner_id, title, description, visibility, created_at, updated_at
		FROM collections WHERE id = $1`, id).
		Scan(&c.ID, &c.OwnerID, &c.Title, &c.Description, &c.Visibility, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Collection{}, fmt.Errorf("find collection: %w", err)
	}
	return c, nil
}

func (r *SocialRepository) CreateCollection(ctx context.Context, c domain.Collection) (domain.Collection, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO collections (id, owner_id, title, description, visibility)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`,
		c.ID, c.OwnerID, c.Title, c.Description, c.Visibility,
	).Scan(&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Collection{}, fmt.Errorf("create collection: %w", err)
	}
	return c, nil
}

func (r *SocialRepository) UpdateCollection(ctx context.Context, c domain.Collection) (domain.Collection, error) {
	err := r.pool.QueryRow(ctx, `
		UPDATE collections SET title = $2, description = $3, visibility = $4, updated_at = NOW()
		WHERE id = $1 AND owner_id = $5
		RETURNING updated_at`,
		c.ID, c.Title, c.Description, c.Visibility, c.OwnerID,
	).Scan(&c.UpdatedAt)
	if err != nil {
		return domain.Collection{}, fmt.Errorf("update collection: %w", err)
	}
	return c, nil
}

func (r *SocialRepository) DeleteCollection(ctx context.Context, id, ownerID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM collections WHERE id = $1 AND owner_id = $2`, id, ownerID)
	return err
}

func (r *SocialRepository) AddCollectionItem(ctx context.Context, collectionID, skillID, ownerID string) error {
	// verify ownership first
	var exists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM collections WHERE id = $1 AND owner_id = $2)`, collectionID, ownerID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("collection not found")
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO collection_items (collection_id, skill_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		collectionID, skillID)
	return err
}

func (r *SocialRepository) RemoveCollectionItem(ctx context.Context, collectionID, skillID, ownerID string) error {
	var exists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM collections WHERE id = $1 AND owner_id = $2)`, collectionID, ownerID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("collection not found")
	}
	_, err := r.pool.Exec(ctx, `DELETE FROM collection_items WHERE collection_id = $1 AND skill_id = $2`, collectionID, skillID)
	return err
}

// ── Run feedback ─────────────────────────────────────────────────────────────

func (r *SocialRepository) CreateRunFeedback(ctx context.Context, fb domain.RunFeedback) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO run_feedback (run_id, user_id, rating, usefulness_score, would_use_again, comment)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (run_id) DO UPDATE
		  SET rating = EXCLUDED.rating,
		      usefulness_score = EXCLUDED.usefulness_score,
		      would_use_again = EXCLUDED.would_use_again,
		      comment = EXCLUDED.comment`,
		fb.RunID, fb.UserID, fb.Rating, fb.UsefulnessScore, fb.WouldUseAgain, fb.Comment)
	return err
}
