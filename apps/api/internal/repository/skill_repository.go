package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

// skillListCols are the SELECT columns for list queries (includes COUNT alias).
const skillListCols = `
	s.id, s.slug, s.title, s.summary, s.description, s.provider, s.entrypoint, s.config_json,
	s.status, s.visibility, s.created_by,
	u.email, p.first_name, p.last_name,
	s.price_cents, s.currency, s.access_type,
	s.is_featured, s.is_verified,
	s.moderation_reason, s.moderated_by, s.moderated_at,
	s.created_at, s.updated_at, COUNT(m.id) AS modules_count`

// skillDetailCols are the SELECT columns for single-skill queries (no COUNT).
const skillDetailCols = `
	s.id, s.slug, s.title, s.summary, s.description, s.provider, s.entrypoint, s.config_json,
	s.status, s.visibility, s.created_by,
	u.email, p.first_name, p.last_name,
	s.price_cents, s.currency, s.access_type,
	s.is_featured, s.is_verified,
	s.moderation_reason, s.moderated_by, s.moderated_at,
	s.created_at, s.updated_at`

type SkillRepository struct {
	pool *pgxpool.Pool
}

func NewSkillRepository(pool *pgxpool.Pool) *SkillRepository {
	return &SkillRepository{pool: pool}
}

func (r *SkillRepository) ListPublished(ctx context.Context) ([]domain.Skill, error) {
	return r.list(ctx, `WHERE s.deleted_at IS NULL AND s.status = 'published'`)
}

func (r *SkillRepository) ListAll(ctx context.Context) ([]domain.Skill, error) {
	return r.list(ctx, `WHERE s.deleted_at IS NULL`)
}

func (r *SkillRepository) ListByCreatorID(ctx context.Context, creatorID string) ([]domain.Skill, error) {
	return r.list(ctx, `WHERE s.deleted_at IS NULL AND s.created_by = $1`, creatorID)
}

func (r *SkillRepository) list(ctx context.Context, clause string, args ...any) ([]domain.Skill, error) {
	query := `
		SELECT ` + skillListCols + `
		FROM skills s
		INNER JOIN users u ON u.id = s.created_by
		INNER JOIN profiles p ON p.user_id = u.id AND p.deleted_at IS NULL
		LEFT JOIN skill_modules m ON m.skill_id = s.id AND m.deleted_at IS NULL
		` + clause + `
		GROUP BY s.id, u.email, p.first_name, p.last_name
		ORDER BY s.updated_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	defer rows.Close()

	skills := make([]domain.Skill, 0)
	for rows.Next() {
		skill, err := scanSkillList(rows)
		if err != nil {
			return nil, fmt.Errorf("scan skill: %w", err)
		}
		skills = append(skills, skill)
	}
	return skills, rows.Err()
}

func (r *SkillRepository) FindByID(ctx context.Context, id string) (domain.Skill, error) {
	query := `
		SELECT ` + skillDetailCols + `
		FROM skills s
		INNER JOIN users u ON u.id = s.created_by
		INNER JOIN profiles p ON p.user_id = u.id AND p.deleted_at IS NULL
		WHERE s.id = $1 AND s.deleted_at IS NULL`

	skill, err := scanSkillDetail(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		return domain.Skill{}, fmt.Errorf("find skill: %w", err)
	}

	modules, err := r.findModules(ctx, id)
	if err != nil {
		return domain.Skill{}, err
	}
	skill.Modules = modules
	skill.ModulesCount = len(modules)
	return skill, nil
}

func (r *SkillRepository) findModules(ctx context.Context, skillID string) ([]domain.Module, error) {
	query := `
		SELECT id, skill_id, slug, title, summary, content, position, status, created_at, updated_at
		FROM skill_modules
		WHERE skill_id = $1 AND deleted_at IS NULL
		ORDER BY position ASC`

	rows, err := r.pool.Query(ctx, query, skillID)
	if err != nil {
		return nil, fmt.Errorf("query modules: %w", err)
	}
	defer rows.Close()

	modules := make([]domain.Module, 0)
	for rows.Next() {
		var module domain.Module
		if err := rows.Scan(&module.ID, &module.SkillID, &module.Slug, &module.Title, &module.Summary, &module.Content, &module.Position, &module.Status, &module.CreatedAt, &module.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan module: %w", err)
		}
		modules = append(modules, module)
	}
	return modules, rows.Err()
}

func (r *SkillRepository) Create(ctx context.Context, skill domain.Skill) (domain.Skill, error) {
	query := `
		INSERT INTO skills (id, slug, title, summary, description, provider, entrypoint, config_json, status, visibility, created_by, price_cents, currency, access_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING created_at, updated_at`
	if err := r.pool.QueryRow(ctx, query,
		skill.ID, skill.Slug, skill.Title, skill.Summary, skill.Description,
		skill.Provider, skill.Entrypoint, skill.Config, skill.Status, skill.Visibility,
		skill.CreatedBy, skill.PriceCents, skill.Currency, skill.AccessType,
	).Scan(&skill.CreatedAt, &skill.UpdatedAt); err != nil {
		return domain.Skill{}, fmt.Errorf("create skill: %w", err)
	}
	return skill, nil
}

func (r *SkillRepository) Update(ctx context.Context, skill domain.Skill) (domain.Skill, error) {
	query := `
		UPDATE skills
		SET slug = $2, title = $3, summary = $4, description = $5, provider = $6,
		    entrypoint = $7, config_json = $8, status = $9, visibility = $10,
		    price_cents = $11, currency = $12, access_type = $13,
		    updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING created_by, created_at, updated_at`
	if err := r.pool.QueryRow(ctx, query,
		skill.ID, skill.Slug, skill.Title, skill.Summary, skill.Description,
		skill.Provider, skill.Entrypoint, skill.Config, skill.Status, skill.Visibility,
		skill.PriceCents, skill.Currency, skill.AccessType,
	).Scan(&skill.CreatedBy, &skill.CreatedAt, &skill.UpdatedAt); err != nil {
		return domain.Skill{}, fmt.Errorf("update skill: %w", err)
	}
	return skill, nil
}

// UpdateGovernance sets the moderation status, reason, and featured/verified flags.
func (r *SkillRepository) UpdateGovernance(ctx context.Context, id, actorID string, status domain.SkillStatus, reason *string, featured, verified bool) (domain.Skill, error) {
	query := `
		UPDATE skills
		SET status            = $2,
		    moderation_reason = $3,
		    moderated_by      = $4::uuid,
		    moderated_at      = NOW(),
		    is_featured        = $5,
		    is_verified        = $6,
		    updated_at         = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING created_at, updated_at, created_by, slug, title, summary, description,
		          provider, entrypoint, config_json, visibility,
		          price_cents, currency, access_type, is_featured, is_verified,
		          moderation_reason, moderated_by, moderated_at`

	var skill domain.Skill
	var configRaw []byte
	skill.ID = id
	skill.Status = status
	if err := r.pool.QueryRow(ctx, query, id, status, reason, actorID, featured, verified).Scan(
		&skill.CreatedAt, &skill.UpdatedAt, &skill.CreatedBy, &skill.Slug, &skill.Title,
		&skill.Summary, &skill.Description, &skill.Provider, &skill.Entrypoint, &configRaw,
		&skill.Visibility, &skill.PriceCents, &skill.Currency, &skill.AccessType,
		&skill.IsFeatured, &skill.IsVerified, &skill.ModerationReason, &skill.ModeratedBy, &skill.ModeratedAt,
	); err != nil {
		return domain.Skill{}, fmt.Errorf("update skill governance: %w", err)
	}
	config, err := decodeSkillConfig(configRaw)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("decode skill config: %w", err)
	}
	skill.Config = config
	fullSkill, err := r.FindByID(ctx, skill.ID)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("hydrate governed skill: %w", err)
	}
	return fullSkill, nil
}

// UpdatePricing updates price_cents, currency, and access_type for a skill.
func (r *SkillRepository) UpdatePricing(ctx context.Context, id string, priceCents int, currency string, accessType domain.SkillAccessType) (domain.Skill, error) {
	query := `
		UPDATE skills
		SET price_cents = $2, currency = $3, access_type = $4, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING created_at, updated_at, created_by`

	var skill domain.Skill
	skill.ID = id
	skill.PriceCents = priceCents
	skill.Currency = currency
	skill.AccessType = accessType
	if err := r.pool.QueryRow(ctx, query, id, priceCents, currency, accessType).Scan(
		&skill.CreatedAt, &skill.UpdatedAt, &skill.CreatedBy,
	); err != nil {
		return domain.Skill{}, fmt.Errorf("update skill pricing: %w", err)
	}
	fullSkill, err := r.FindByID(ctx, skill.ID)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("hydrate priced skill: %w", err)
	}
	return fullSkill, nil
}

// GetStats returns aggregate skill counts for the admin dashboard.
func (r *SkillRepository) GetStats(ctx context.Context) (domain.AdminSkillStats, error) {
	query := `
		SELECT
			COUNT(*)                                                          AS total,
			COUNT(*) FILTER (WHERE status = 'published')                     AS published,
			COUNT(*) FILTER (WHERE status = 'pending_review')                AS pending_review,
			COUNT(*) FILTER (WHERE status = 'hidden')                        AS hidden,
			COUNT(*) FILTER (WHERE access_type = 'free')                     AS free,
			COUNT(*) FILTER (WHERE access_type = 'paid')                     AS paid,
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '7 days')  AS new_last_7d
		FROM skills
		WHERE deleted_at IS NULL`

	var s domain.AdminSkillStats
	if err := r.pool.QueryRow(ctx, query).Scan(
		&s.Total, &s.Published, &s.PendingReview, &s.Hidden, &s.Free, &s.Paid, &s.NewLast7d,
	); err != nil {
		return domain.AdminSkillStats{}, fmt.Errorf("skill stats: %w", err)
	}
	return s, nil
}

func (r *SkillRepository) SoftDelete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE skills SET deleted_at = NOW(), status = 'archived', updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("soft delete skill: %w", err)
	}
	_, err = r.pool.Exec(ctx, `UPDATE skill_modules SET deleted_at = NOW(), updated_at = NOW(), status = 'archived' WHERE skill_id = $1`, id)
	if err != nil {
		return fmt.Errorf("soft delete modules: %w", err)
	}
	return nil
}

// ── scan helpers ─────────────────────────────────────────────────────────────

func scanSkillList(row scanner) (domain.Skill, error) {
	var skill domain.Skill
	var configRaw []byte
	if err := row.Scan(
		&skill.ID, &skill.Slug, &skill.Title, &skill.Summary, &skill.Description,
		&skill.Provider, &skill.Entrypoint, &configRaw,
		&skill.Status, &skill.Visibility, &skill.CreatedBy,
		&skill.Creator.Email, &skill.Creator.FirstName, &skill.Creator.LastName,
		&skill.PriceCents, &skill.Currency, &skill.AccessType,
		&skill.IsFeatured, &skill.IsVerified,
		&skill.ModerationReason, &skill.ModeratedBy, &skill.ModeratedAt,
		&skill.CreatedAt, &skill.UpdatedAt, &skill.ModulesCount,
	); err != nil {
		return domain.Skill{}, err
	}
	config, err := decodeSkillConfig(configRaw)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("decode skill config: %w", err)
	}
	skill.Config = config
	skill.Creator.ID = skill.CreatedBy
	return skill, nil
}

func scanSkillDetail(row scanner) (domain.Skill, error) {
	var skill domain.Skill
	var configRaw []byte
	if err := row.Scan(
		&skill.ID, &skill.Slug, &skill.Title, &skill.Summary, &skill.Description,
		&skill.Provider, &skill.Entrypoint, &configRaw,
		&skill.Status, &skill.Visibility, &skill.CreatedBy,
		&skill.Creator.Email, &skill.Creator.FirstName, &skill.Creator.LastName,
		&skill.PriceCents, &skill.Currency, &skill.AccessType,
		&skill.IsFeatured, &skill.IsVerified,
		&skill.ModerationReason, &skill.ModeratedBy, &skill.ModeratedAt,
		&skill.CreatedAt, &skill.UpdatedAt,
	); err != nil {
		return domain.Skill{}, err
	}
	config, err := decodeSkillConfig(configRaw)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("decode skill config: %w", err)
	}
	skill.Config = config
	skill.Creator.ID = skill.CreatedBy
	return skill, nil
}

func decodeSkillConfig(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}
	if decoded == nil {
		return map[string]any{}, nil
	}
	return decoded, nil
}
