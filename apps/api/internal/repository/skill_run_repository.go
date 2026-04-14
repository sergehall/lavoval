package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type SkillRunRepository struct {
	pool *pgxpool.Pool
}

func NewSkillRunRepository(pool *pgxpool.Pool) *SkillRunRepository {
	return &SkillRunRepository{pool: pool}
}

func (r *SkillRunRepository) Create(ctx context.Context, run domain.SkillRun) (domain.SkillRun, error) {
	query := `
		INSERT INTO skill_runs (
			id, skill_id, user_id, status, input_json, output_json, error_message, started_at, finished_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at`

	if err := r.pool.QueryRow(
		ctx,
		query,
		run.ID,
		run.SkillID,
		run.UserID,
		run.Status,
		run.Input,
		run.Output,
		run.ErrorMessage,
		run.StartedAt,
		run.FinishedAt,
	).Scan(&run.CreatedAt); err != nil {
		return domain.SkillRun{}, fmt.Errorf("create skill run: %w", err)
	}

	return run, nil
}

func (r *SkillRunRepository) Update(ctx context.Context, run domain.SkillRun) (domain.SkillRun, error) {
	query := `
		UPDATE skill_runs
		SET status = $2,
		    input_json = $3,
		    output_json = $4,
		    error_message = $5,
		    started_at = $6,
		    finished_at = $7
		WHERE id = $1
		RETURNING skill_id, user_id, created_at`

	if err := r.pool.QueryRow(
		ctx,
		query,
		run.ID,
		run.Status,
		run.Input,
		run.Output,
		run.ErrorMessage,
		run.StartedAt,
		run.FinishedAt,
	).Scan(&run.SkillID, &run.UserID, &run.CreatedAt); err != nil {
		return domain.SkillRun{}, fmt.Errorf("update skill run: %w", err)
	}

	return run, nil
}

func (r *SkillRunRepository) FindByID(ctx context.Context, id string) (domain.SkillRun, error) {
	query := `
		SELECT r.id, r.skill_id, r.user_id,
		       s.id, s.slug, s.title, s.entrypoint,
		       r.status, r.input_json, r.output_json, r.error_message, r.started_at, r.finished_at, r.created_at
		FROM skill_runs r
		INNER JOIN skills s ON s.id = r.skill_id
		WHERE r.id = $1`

	row := r.pool.QueryRow(ctx, query, id)
	run, err := scanSkillRun(row.Scan)
	if err != nil {
		return domain.SkillRun{}, fmt.Errorf("find skill run: %w", err)
	}

	return run, nil
}

func (r *SkillRunRepository) ListByUserID(ctx context.Context, userID string) ([]domain.SkillRun, error) {
	return r.list(ctx, `WHERE r.user_id = $1`, userID)
}

func (r *SkillRunRepository) ListAll(ctx context.Context) ([]domain.SkillRun, error) {
	return r.list(ctx, "")
}

type scannerFn func(dest ...any) error

func scanSkillRun(scan scannerFn) (domain.SkillRun, error) {
	var run domain.SkillRun
	var inputRaw []byte
	var outputRaw []byte

	if err := scan(
		&run.ID,
		&run.SkillID,
		&run.UserID,
		&run.Skill.ID,
		&run.Skill.Slug,
		&run.Skill.Title,
		&run.Skill.Entrypoint,
		&run.Status,
		&inputRaw,
		&outputRaw,
		&run.ErrorMessage,
		&run.StartedAt,
		&run.FinishedAt,
		&run.CreatedAt,
	); err != nil {
		return domain.SkillRun{}, err
	}

	input, err := decodeJSONMap(inputRaw)
	if err != nil {
		return domain.SkillRun{}, fmt.Errorf("decode input json: %w", err)
	}
	output, err := decodeJSONMap(outputRaw)
	if err != nil {
		return domain.SkillRun{}, fmt.Errorf("decode output json: %w", err)
	}

	run.Input = input
	run.Output = output
	finalizeSkillRun(&run)
	return run, nil
}

func decodeJSONMap(raw []byte) (map[string]any, error) {
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

func finalizeSkillRun(run *domain.SkillRun) {
	run.Meta.InputKeysCount = len(run.Input)
	run.Meta.OutputKeysCount = len(run.Output)
	run.Meta.HasOutput = len(run.Output) > 0
	run.Meta.HasError = run.ErrorMessage != nil && *run.ErrorMessage != ""

	if run.StartedAt != nil && run.FinishedAt != nil {
		duration := run.FinishedAt.Sub(*run.StartedAt).Milliseconds()
		run.Meta.DurationMs = &duration
	}
}

func (r *SkillRunRepository) list(ctx context.Context, clause string, args ...any) ([]domain.SkillRun, error) {
	query := `
		SELECT r.id, r.skill_id, r.user_id,
		       s.id, s.slug, s.title, s.entrypoint,
		       r.status, r.input_json, r.output_json, r.error_message, r.started_at, r.finished_at, r.created_at
		FROM skill_runs r
		INNER JOIN skills s ON s.id = r.skill_id
		` + clause + `
		ORDER BY r.created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list skill runs: %w", err)
	}
	defer rows.Close()

	runs := make([]domain.SkillRun, 0)
	for rows.Next() {
		run, scanErr := scanSkillRun(rows.Scan)
		if scanErr != nil {
			return nil, fmt.Errorf("scan skill run: %w", scanErr)
		}
		runs = append(runs, run)
	}

	return runs, rows.Err()
}
