package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type SkillVersionRepository struct {
	pool *pgxpool.Pool
}

func NewSkillVersionRepository(pool *pgxpool.Pool) *SkillVersionRepository {
	return &SkillVersionRepository{pool: pool}
}

func (r *SkillVersionRepository) FindCurrentBySkillID(ctx context.Context, skillID string) (*domain.SkillVersion, error) {
	const query = `
		SELECT
			id, skill_id, version_no, is_current, changelog, content_md,
			prompt_template, system_instructions,
			input_schema_json, output_schema_json, error_schema_json,
			created_by, created_at
		FROM lavoval_skill_versions
		WHERE skill_id = $1 AND is_current = true
		ORDER BY version_no DESC
		LIMIT 1`

	version, err := scanSkillVersion(r.pool.QueryRow(ctx, query, skillID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find current skill version: %w", err)
	}
	return &version, nil
}

func (r *SkillVersionRepository) SaveCurrent(ctx context.Context, version domain.SkillVersion) (domain.SkillVersion, error) {
	if version.ID == "" {
		version.ID = uuid.NewString()
	}

	const query = `
		WITH current_version AS (
			SELECT COALESCE(MAX(version_no), 0) AS max_version
			FROM lavoval_skill_versions
			WHERE skill_id = $1
		),
		clear_current AS (
			UPDATE lavoval_skill_versions
			SET is_current = false
			WHERE skill_id = $1 AND is_current = true
		)
		INSERT INTO lavoval_skill_versions (
			id, skill_id, version_no, is_current, changelog, content_md,
			prompt_template, system_instructions,
			input_schema_json, output_schema_json, error_schema_json,
			created_by
		) VALUES (
			$2, $1, (SELECT max_version + 1 FROM current_version), true, $3, $4,
			$5, $6,
			$7, $8, $9,
			$10
		)
		RETURNING
			id, skill_id, version_no, is_current, changelog, content_md,
			prompt_template, system_instructions,
			input_schema_json, output_schema_json, error_schema_json,
			created_by, created_at`

	saved, err := scanSkillVersion(r.pool.QueryRow(
		ctx,
		query,
		version.SkillID,
		version.ID,
		version.Changelog,
		version.ContentMD,
		version.PromptTemplate,
		version.SystemInstructions,
		version.InputSchema,
		version.OutputSchema,
		version.ErrorSchema,
		version.CreatedBy,
	))
	if err != nil {
		return domain.SkillVersion{}, fmt.Errorf("save current skill version: %w", err)
	}
	return saved, nil
}

func scanSkillVersion(row scanner) (domain.SkillVersion, error) {
	var version domain.SkillVersion
	var inputSchemaRaw []byte
	var outputSchemaRaw []byte
	var errorSchemaRaw []byte

	if err := row.Scan(
		&version.ID,
		&version.SkillID,
		&version.VersionNo,
		&version.IsCurrent,
		&version.Changelog,
		&version.ContentMD,
		&version.PromptTemplate,
		&version.SystemInstructions,
		&inputSchemaRaw,
		&outputSchemaRaw,
		&errorSchemaRaw,
		&version.CreatedBy,
		&version.CreatedAt,
	); err != nil {
		return domain.SkillVersion{}, err
	}

	var err error
	version.InputSchema, err = decodeSkillVersionJSONMap(inputSchemaRaw)
	if err != nil {
		return domain.SkillVersion{}, fmt.Errorf("decode input schema: %w", err)
	}
	version.OutputSchema, err = decodeSkillVersionJSONMap(outputSchemaRaw)
	if err != nil {
		return domain.SkillVersion{}, fmt.Errorf("decode output schema: %w", err)
	}
	version.ErrorSchema, err = decodeSkillVersionJSONMap(errorSchemaRaw)
	if err != nil {
		return domain.SkillVersion{}, fmt.Errorf("decode error schema: %w", err)
	}

	return version, nil
}

func decodeSkillVersionJSONMap(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}
	if decoded == nil {
		return nil, nil
	}
	return decoded, nil
}
