package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type MailCleanupRunRepository struct {
	pool *pgxpool.Pool
}

func NewMailCleanupRunRepository(pool *pgxpool.Pool) *MailCleanupRunRepository {
	return &MailCleanupRunRepository{pool: pool}
}

func (r *MailCleanupRunRepository) Create(ctx context.Context, item domain.MailCleanupRun) (domain.MailCleanupRun, error) {
	if err := r.pool.QueryRow(ctx, `
		INSERT INTO lavoval_mail_cleanup_runs (
			id, mode, status, dry_run, candidate_jobs, candidate_events,
			deleted_jobs, deleted_events, error_message, duration_ms
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at
	`,
		item.ID,
		item.Mode,
		item.Status,
		item.DryRun,
		item.CandidateJobs,
		item.CandidateEvents,
		item.DeletedJobs,
		item.DeletedEvents,
		item.ErrorMessage,
		item.DurationMs,
	).Scan(&item.CreatedAt); err != nil {
		return domain.MailCleanupRun{}, fmt.Errorf("insert mail cleanup run: %w", err)
	}

	return item, nil
}

func (r *MailCleanupRunRepository) ListRecent(ctx context.Context, limit int) ([]domain.MailCleanupRun, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, mode, status, dry_run, candidate_jobs, candidate_events,
		       deleted_jobs, deleted_events, error_message, duration_ms, created_at
		FROM lavoval_mail_cleanup_runs
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list mail cleanup runs: %w", err)
	}
	defer rows.Close()

	items := make([]domain.MailCleanupRun, 0, limit)
	for rows.Next() {
		var item domain.MailCleanupRun
		if err := rows.Scan(
			&item.ID,
			&item.Mode,
			&item.Status,
			&item.DryRun,
			&item.CandidateJobs,
			&item.CandidateEvents,
			&item.DeletedJobs,
			&item.DeletedEvents,
			&item.ErrorMessage,
			&item.DurationMs,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan mail cleanup run: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mail cleanup runs: %w", err)
	}

	return items, nil
}
