package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type MailEventRepository struct {
	pool *pgxpool.Pool
}

func NewMailEventRepository(pool *pgxpool.Pool) *MailEventRepository {
	return &MailEventRepository{pool: pool}
}

func (r *MailEventRepository) Append(ctx context.Context, event domain.MailEvent) (domain.MailEvent, error) {
	query := `
		INSERT INTO mail_events (
			id, job_id, event_type, message_type, provider, provider_message_id,
			recipient_email, error_code, attempt, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at`

	if err := r.pool.QueryRow(
		ctx,
		query,
		event.ID,
		event.JobID,
		event.EventType,
		event.MessageType,
		event.Provider,
		event.ProviderMessageID,
		event.RecipientEmail,
		event.ErrorCode,
		event.Attempt,
		event.Metadata,
	).Scan(&event.CreatedAt); err != nil {
		return domain.MailEvent{}, fmt.Errorf("insert mail event: %w", err)
	}

	return event, nil
}

func (r *MailEventRepository) ListRecent(ctx context.Context, limit int) ([]domain.MailEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, job_id, event_type, message_type, provider, provider_message_id,
		       recipient_email, error_code, attempt, metadata, created_at
		FROM mail_events
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent mail events: %w", err)
	}
	defer rows.Close()

	return scanMailEvents(rows)
}

func (r *MailEventRepository) ListByJobID(ctx context.Context, jobID string, limit int) ([]domain.MailEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, job_id, event_type, message_type, provider, provider_message_id,
		       recipient_email, error_code, attempt, metadata, created_at
		FROM mail_events
		WHERE job_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, jobID, limit)
	if err != nil {
		return nil, fmt.Errorf("list mail events by job: %w", err)
	}
	defer rows.Close()

	return scanMailEvents(rows)
}

func scanMailEvents(rows pgxRows) ([]domain.MailEvent, error) {
	events := make([]domain.MailEvent, 0)
	for rows.Next() {
		var event domain.MailEvent
		if err := rows.Scan(
			&event.ID,
			&event.JobID,
			&event.EventType,
			&event.MessageType,
			&event.Provider,
			&event.ProviderMessageID,
			&event.RecipientEmail,
			&event.ErrorCode,
			&event.Attempt,
			&event.Metadata,
			&event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan mail event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mail events: %w", err)
	}

	return events, nil
}

type pgxRows interface {
	Next() bool
	Scan(...any) error
	Err() error
}
