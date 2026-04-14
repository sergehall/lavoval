package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

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
		SELECT s.id, s.slug, s.title, s.summary, s.description, s.provider, s.entrypoint, s.config_json, s.status, s.visibility, s.created_by,
		       u.email, p.first_name, p.last_name,
		       s.created_at, s.updated_at, COUNT(m.id) AS modules_count
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
		var skill domain.Skill
		var configRaw []byte
		if err := rows.Scan(
			&skill.ID,
			&skill.Slug,
			&skill.Title,
			&skill.Summary,
			&skill.Description,
			&skill.Provider,
			&skill.Entrypoint,
			&configRaw,
			&skill.Status,
			&skill.Visibility,
			&skill.CreatedBy,
			&skill.Creator.Email,
			&skill.Creator.FirstName,
			&skill.Creator.LastName,
			&skill.CreatedAt,
			&skill.UpdatedAt,
			&skill.ModulesCount,
		); err != nil {
			return nil, fmt.Errorf("scan skill: %w", err)
		}
		config, err := decodeSkillConfig(configRaw)
		if err != nil {
			return nil, fmt.Errorf("decode skill config: %w", err)
		}
		skill.Config = config
		skill.Creator.ID = skill.CreatedBy
		skills = append(skills, skill)
	}
	return skills, rows.Err()
}

func (r *SkillRepository) FindByID(ctx context.Context, id string) (domain.Skill, error) {
	query := `
		SELECT s.id, s.slug, s.title, s.summary, s.description, s.provider, s.entrypoint, s.config_json, s.status, s.visibility, s.created_by,
		       u.email, p.first_name, p.last_name,
		       s.created_at, s.updated_at
		FROM skills s
		INNER JOIN users u ON u.id = s.created_by
		INNER JOIN profiles p ON p.user_id = u.id AND p.deleted_at IS NULL
		WHERE s.id = $1 AND s.deleted_at IS NULL`

	var skill domain.Skill
	var configRaw []byte
	if err := r.pool.QueryRow(ctx, query, id).Scan(
		&skill.ID,
		&skill.Slug,
		&skill.Title,
		&skill.Summary,
		&skill.Description,
		&skill.Provider,
		&skill.Entrypoint,
		&configRaw,
		&skill.Status,
		&skill.Visibility,
		&skill.CreatedBy,
		&skill.Creator.Email,
		&skill.Creator.FirstName,
		&skill.Creator.LastName,
		&skill.CreatedAt,
		&skill.UpdatedAt,
	); err != nil {
		return domain.Skill{}, fmt.Errorf("find skill: %w", err)
	}
	config, err := decodeSkillConfig(configRaw)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("decode skill config: %w", err)
	}
	skill.Config = config
	skill.Creator.ID = skill.CreatedBy

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
		INSERT INTO skills (id, slug, title, summary, description, provider, entrypoint, config_json, status, visibility, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`
	if err := r.pool.QueryRow(ctx, query, skill.ID, skill.Slug, skill.Title, skill.Summary, skill.Description, skill.Provider, skill.Entrypoint, skill.Config, skill.Status, skill.Visibility, skill.CreatedBy).Scan(&skill.CreatedAt, &skill.UpdatedAt); err != nil {
		return domain.Skill{}, fmt.Errorf("create skill: %w", err)
	}
	return skill, nil
}

func (r *SkillRepository) Update(ctx context.Context, skill domain.Skill) (domain.Skill, error) {
	query := `
		UPDATE skills
		SET slug = $2, title = $3, summary = $4, description = $5, provider = $6, entrypoint = $7, config_json = $8, status = $9, visibility = $10, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING created_by, created_at, updated_at`
	if err := r.pool.QueryRow(ctx, query, skill.ID, skill.Slug, skill.Title, skill.Summary, skill.Description, skill.Provider, skill.Entrypoint, skill.Config, skill.Status, skill.Visibility).Scan(&skill.CreatedBy, &skill.CreatedAt, &skill.UpdatedAt); err != nil {
		return domain.Skill{}, fmt.Errorf("update skill: %w", err)
	}
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
