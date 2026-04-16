package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

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

func (r *MailEventRepository) ListRecent(ctx context.Context, filter domain.MailEventFilter) ([]domain.MailEvent, error) {
	query, args := buildMailEventListQuery("", filter)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list recent mail events: %w", err)
	}
	defer rows.Close()

	return scanMailEvents(rows)
}

func (r *MailEventRepository) ListByJobID(ctx context.Context, jobID string, filter domain.MailEventFilter) ([]domain.MailEvent, error) {
	filter.JobID = jobID
	query, args := buildMailEventListQuery("job_id = $1", filter)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list mail events by job: %w", err)
	}
	defer rows.Close()

	return scanMailEvents(rows)
}

func (r *MailEventRepository) CountBefore(ctx context.Context, before time.Time) (int64, error) {
	var count int64
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM mail_events
		WHERE created_at < $1
	`, before).Scan(&count); err != nil {
		return 0, fmt.Errorf("count purgeable mail events: %w", err)
	}
	return count, nil
}

func (r *MailEventRepository) DeleteBefore(ctx context.Context, before time.Time, limit int) (int64, error) {
	if limit <= 0 {
		limit = 500
	}

	tag, err := r.pool.Exec(ctx, `
		WITH doomed AS (
			SELECT id
			FROM mail_events
			WHERE created_at < $1
			ORDER BY created_at ASC
			LIMIT $2
		)
		DELETE FROM mail_events
		WHERE id IN (SELECT id FROM doomed)
	`, before, limit)
	if err != nil {
		return 0, fmt.Errorf("delete purgeable mail events: %w", err)
	}

	return tag.RowsAffected(), nil
}

func buildMailEventListQuery(baseCondition string, filter domain.MailEventFilter) (string, []any) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, job_id, event_type, message_type, provider, provider_message_id,
		       recipient_email, error_code, attempt, metadata, created_at
		FROM mail_events
	`

	conditions := make([]string, 0, 6)
	args := make([]any, 0, 7)
	nextArg := 1

	if baseCondition != "" {
		conditions = append(conditions, baseCondition)
		args = append(args, filter.JobID)
		nextArg++
	}

	if queryText := strings.TrimSpace(filter.Query); queryText != "" {
		conditions = append(conditions, fmt.Sprintf("(recipient_email::text ILIKE $%d OR job_id::text ILIKE $%d OR COALESCE(provider_message_id, '') ILIKE $%d)", nextArg, nextArg, nextArg))
		args = append(args, "%"+queryText+"%")
		nextArg++
	}
	if value := strings.TrimSpace(filter.EventType); value != "" {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", nextArg))
		args = append(args, value)
		nextArg++
	}
	if value := strings.TrimSpace(filter.MessageType); value != "" {
		conditions = append(conditions, fmt.Sprintf("message_type = $%d", nextArg))
		args = append(args, value)
		nextArg++
	}
	if value := strings.TrimSpace(filter.Provider); value != "" {
		conditions = append(conditions, fmt.Sprintf("provider = $%d", nextArg))
		args = append(args, value)
		nextArg++
	}
	if value := strings.TrimSpace(filter.ErrorCode); value != "" {
		conditions = append(conditions, fmt.Sprintf("error_code = $%d", nextArg))
		args = append(args, value)
		nextArg++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", nextArg)
	args = append(args, limit)

	return query, args
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
