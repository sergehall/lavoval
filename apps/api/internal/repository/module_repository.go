package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type ModuleRepository struct {
	pool *pgxpool.Pool
}

func NewModuleRepository(pool *pgxpool.Pool) *ModuleRepository {
	return &ModuleRepository{pool: pool}
}

const moduleColumns = `id, skill_id, slug, title, summary, content, position, status, created_at, updated_at`

func scanModule(row interface {
	Scan(...any) error
}) (domain.Module, error) {
	var m domain.Module
	if err := row.Scan(
		&m.ID, &m.SkillID, &m.Slug, &m.Title,
		&m.Summary, &m.Content, &m.Position, &m.Status,
		&m.CreatedAt, &m.UpdatedAt,
	); err != nil {
		return domain.Module{}, err
	}
	return m, nil
}

// ListBySkillID returns all non-deleted modules for a skill, ordered by position.
func (r *ModuleRepository) ListBySkillID(ctx context.Context, skillID string) ([]domain.Module, error) {
	query := `
		SELECT ` + moduleColumns + `
		FROM skill_modules
		WHERE skill_id = $1 AND deleted_at IS NULL
		ORDER BY position ASC`

	rows, err := r.pool.Query(ctx, query, skillID)
	if err != nil {
		return nil, fmt.Errorf("list modules: %w", err)
	}
	defer rows.Close()

	modules := make([]domain.Module, 0)
	for rows.Next() {
		m, err := scanModule(rows)
		if err != nil {
			return nil, fmt.Errorf("scan module: %w", err)
		}
		modules = append(modules, m)
	}
	return modules, rows.Err()
}

// FindByID returns a single module by its ID.
func (r *ModuleRepository) FindByID(ctx context.Context, id string) (domain.Module, error) {
	query := `
		SELECT ` + moduleColumns + `
		FROM skill_modules
		WHERE id = $1 AND deleted_at IS NULL`

	m, err := scanModule(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		return domain.Module{}, fmt.Errorf("find module: %w", err)
	}
	return m, nil
}

// Create inserts a new module.  When position is not set (zero), it is
// auto-assigned as max(position)+1 for the skill to append to the end.
// The UNIQUE(skill_id, slug) constraint prevents duplicate slugs per skill.
func (r *ModuleRepository) Create(ctx context.Context, module domain.Module) (domain.Module, error) {
	query := `
		INSERT INTO skill_modules (id, skill_id, slug, title, summary, content, position, status)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			CASE WHEN $7::int >= 0
				THEN $7::int
				ELSE COALESCE(
					(SELECT MAX(position) + 1 FROM skill_modules WHERE skill_id = $2 AND deleted_at IS NULL),
					0
				)
			END,
			$8
		)
		RETURNING ` + moduleColumns

	m, err := scanModule(r.pool.QueryRow(ctx, query,
		module.ID, module.SkillID, module.Slug, module.Title,
		module.Summary, module.Content, module.Position, module.Status,
	))
	if err != nil {
		return domain.Module{}, fmt.Errorf("create module: %w", err)
	}
	return m, nil
}

// Update saves all editable module fields.  All values are bound parameters.
func (r *ModuleRepository) Update(ctx context.Context, module domain.Module) (domain.Module, error) {
	query := `
		UPDATE skill_modules
		SET slug       = $2,
		    title      = $3,
		    summary    = $4,
		    content    = $5,
		    position   = $6,
		    status     = $7,
		    updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + moduleColumns

	m, err := scanModule(r.pool.QueryRow(ctx, query,
		module.ID, module.Slug, module.Title,
		module.Summary, module.Content, module.Position, module.Status,
	))
	if err != nil {
		return domain.Module{}, fmt.Errorf("update module: %w", err)
	}
	return m, nil
}

// SoftDelete marks a module as deleted without removing the row.
func (r *ModuleRepository) SoftDelete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE skill_modules SET deleted_at = NOW(), status = 'archived', updated_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("soft delete module: %w", err)
	}
	return nil
}
