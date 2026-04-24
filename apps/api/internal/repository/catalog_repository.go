package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type CatalogRepository struct {
	pool *pgxpool.Pool
}

func NewCatalogRepository(pool *pgxpool.Pool) *CatalogRepository {
	return &CatalogRepository{pool: pool}
}

func (r *CatalogRepository) ListCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, description, icon, sort_order, created_at
		FROM lavoval_categories
		ORDER BY sort_order ASC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list lavoval_categories: %w", err)
	}
	defer rows.Close()

	var out []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.Description, &c.Icon, &c.SortOrder, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CatalogRepository) ListSubcategories(ctx context.Context, categoryID string) ([]domain.Subcategory, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, category_id, slug, name, sort_order, created_at
		FROM lavoval_subcategories
		WHERE category_id = $1
		ORDER BY sort_order ASC, name ASC`, categoryID)
	if err != nil {
		return nil, fmt.Errorf("list lavoval_subcategories: %w", err)
	}
	defer rows.Close()

	var out []domain.Subcategory
	for rows.Next() {
		var s domain.Subcategory
		if err := rows.Scan(&s.ID, &s.CategoryID, &s.Slug, &s.Name, &s.SortOrder, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan subcategory: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *CatalogRepository) ListAllSubcategories(ctx context.Context) ([]domain.Subcategory, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, category_id, slug, name, sort_order, created_at
		FROM lavoval_subcategories
		ORDER BY sort_order ASC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list all lavoval_subcategories: %w", err)
	}
	defer rows.Close()

	var out []domain.Subcategory
	for rows.Next() {
		var s domain.Subcategory
		if err := rows.Scan(&s.ID, &s.CategoryID, &s.Slug, &s.Name, &s.SortOrder, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan subcategory: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *CatalogRepository) ListTags(ctx context.Context) ([]domain.Tag, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, kind, created_at
		FROM lavoval_tags
		ORDER BY kind ASC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("list lavoval_tags: %w", err)
	}
	defer rows.Close()

	var out []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.Slug, &t.Name, &t.Kind, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *CatalogRepository) ListTagsForSkill(ctx context.Context, skillID string) ([]domain.Tag, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.id, t.slug, t.name, t.kind, t.created_at
		FROM lavoval_tags t
		INNER JOIN lavoval_skill_tag_links stl ON stl.tag_id = t.id
		WHERE stl.skill_id = $1
		ORDER BY t.kind ASC, t.name ASC`, skillID)
	if err != nil {
		return nil, fmt.Errorf("list lavoval_tags for skill: %w", err)
	}
	defer rows.Close()

	var out []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.Slug, &t.Name, &t.Kind, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *CatalogRepository) SetSkillTags(ctx context.Context, skillID string, tagIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `DELETE FROM lavoval_skill_tag_links WHERE skill_id = $1`, skillID); err != nil {
		return fmt.Errorf("clear skill lavoval_tags: %w", err)
	}

	for _, tagID := range tagIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO lavoval_skill_tag_links (skill_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			skillID, tagID,
		); err != nil {
			return fmt.Errorf("insert skill tag: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *CatalogRepository) FindTagsBySlug(ctx context.Context, slugs []string) ([]domain.Tag, error) {
	if len(slugs) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, kind, created_at
		FROM lavoval_tags
		WHERE slug = ANY($1)`, slugs)
	if err != nil {
		return nil, fmt.Errorf("find lavoval_tags by slug: %w", err)
	}
	defer rows.Close()

	var out []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.Slug, &t.Name, &t.Kind, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
