package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type MailJobRepository struct {
	pool *pgxpool.Pool
}

func NewMailJobRepository(pool *pgxpool.Pool) *MailJobRepository {
	return &MailJobRepository{pool: pool}
}

func (r *MailJobRepository) Enqueue(ctx context.Context, job domain.MailJob) (domain.MailJob, bool, error) {
	query := `
		WITH inserted AS (
			INSERT INTO mail_jobs (
				id, message_type, recipient_email, idempotency_key, payload, status, attempts, max_attempts, next_attempt_at
			)
			VALUES ($1, $2, $3, $9, $4, $5, $6, $7, $8)
			ON CONFLICT DO NOTHING
			RETURNING
				id,
				message_type,
				recipient_email,
				idempotency_key,
				payload,
				status,
				attempts,
				max_attempts,
				next_attempt_at,
				leased_until,
				last_error,
				last_error_code,
				provider,
				provider_message_id,
				sent_at,
				dead_lettered_at,
				created_at,
				updated_at,
				FALSE AS deduplicated
		),
		existing AS (
			SELECT
				id,
				message_type,
				recipient_email,
				idempotency_key,
				payload,
				status,
				attempts,
				max_attempts,
				next_attempt_at,
				leased_until,
				last_error,
				last_error_code,
				provider,
				provider_message_id,
				sent_at,
				dead_lettered_at,
				created_at,
				updated_at,
				TRUE AS deduplicated
			FROM mail_jobs
			WHERE idempotency_key = $9
			  AND $9 IS NOT NULL
			  AND status IN ('queued', 'retrying', 'processing')
			ORDER BY created_at ASC
			LIMIT 1
		)
		SELECT * FROM inserted
		UNION ALL
		SELECT * FROM existing
		WHERE NOT EXISTS (SELECT 1 FROM inserted)`

	var deduplicated bool
	if err := r.pool.QueryRow(
		ctx,
		query,
		job.ID,
		job.MessageType,
		job.RecipientEmail,
		job.Payload,
		job.Status,
		job.Attempts,
		job.MaxAttempts,
		job.NextAttemptAt,
		job.IdempotencyKey,
	).Scan(
		&job.ID,
		&job.MessageType,
		&job.RecipientEmail,
		&job.IdempotencyKey,
		&job.Payload,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.NextAttemptAt,
		&job.LeasedUntil,
		&job.LastError,
		&job.LastErrorCode,
		&job.Provider,
		&job.ProviderMessageID,
		&job.SentAt,
		&job.DeadLetteredAt,
		&job.CreatedAt,
		&job.UpdatedAt,
		&deduplicated,
	); err != nil {
		return domain.MailJob{}, false, fmt.Errorf("insert mail job: %w", err)
	}

	return job, deduplicated, nil
}

func (r *MailJobRepository) ClaimNext(ctx context.Context, leaseDuration time.Duration) (domain.MailJob, bool, error) {
	query := `
		WITH candidate AS (
			SELECT id
			FROM mail_jobs
			WHERE (
				status IN ('queued', 'retrying') AND next_attempt_at <= NOW()
			) OR (
				status = 'processing' AND leased_until IS NOT NULL AND leased_until <= NOW()
			)
			ORDER BY next_attempt_at ASC, created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE mail_jobs m
		SET
			status = 'processing',
			attempts = m.attempts + 1,
			leased_until = NOW() + $1::interval,
			updated_at = NOW()
		FROM candidate
		WHERE m.id = candidate.id
		RETURNING
			m.id,
			m.message_type,
			m.recipient_email,
			m.idempotency_key,
			m.payload,
			m.status,
			m.attempts,
			m.max_attempts,
			m.next_attempt_at,
			m.leased_until,
			m.last_error,
			m.last_error_code,
			m.provider,
			m.provider_message_id,
			m.sent_at,
			m.dead_lettered_at,
			m.created_at,
			m.updated_at`

	var job domain.MailJob
	err := r.pool.QueryRow(ctx, query, pgInterval(leaseDuration)).Scan(
		&job.ID,
		&job.MessageType,
		&job.RecipientEmail,
		&job.IdempotencyKey,
		&job.Payload,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.NextAttemptAt,
		&job.LeasedUntil,
		&job.LastError,
		&job.LastErrorCode,
		&job.Provider,
		&job.ProviderMessageID,
		&job.SentAt,
		&job.DeadLetteredAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.MailJob{}, false, nil
		}
		return domain.MailJob{}, false, fmt.Errorf("claim mail job: %w", err)
	}

	return job, true, nil
}

func (r *MailJobRepository) MarkSent(ctx context.Context, jobID string, provider string, providerMessageID string) error {
	query := `
		UPDATE mail_jobs
		SET
			status = 'sent',
			provider = $2,
			provider_message_id = NULLIF($3, ''),
			leased_until = NULL,
			last_error = NULL,
			last_error_code = NULL,
			sent_at = NOW(),
			updated_at = NOW()
		WHERE id = $1`

	if _, err := r.pool.Exec(ctx, query, jobID, provider, providerMessageID); err != nil {
		return fmt.Errorf("mark mail job sent: %w", err)
	}
	return nil
}

func (r *MailJobRepository) MarkRetry(ctx context.Context, jobID string, lastError string, errorCode string, nextAttemptAt time.Time) error {
	query := `
		UPDATE mail_jobs
		SET
			status = 'retrying',
			leased_until = NULL,
			last_error = $2,
			last_error_code = $3,
			next_attempt_at = $4,
			updated_at = NOW()
		WHERE id = $1`

	if _, err := r.pool.Exec(ctx, query, jobID, lastError, errorCode, nextAttemptAt); err != nil {
		return fmt.Errorf("mark mail job retry: %w", err)
	}
	return nil
}

func (r *MailJobRepository) MarkDeadLetter(ctx context.Context, jobID string, lastError string, errorCode string) error {
	query := `
		UPDATE mail_jobs
		SET
			status = 'dead_letter',
			leased_until = NULL,
			last_error = $2,
			last_error_code = $3,
			dead_lettered_at = NOW(),
			updated_at = NOW()
		WHERE id = $1`

	if _, err := r.pool.Exec(ctx, query, jobID, lastError, errorCode); err != nil {
		return fmt.Errorf("mark mail job dead letter: %w", err)
	}
	return nil
}

func (r *MailJobRepository) CountByStatus(ctx context.Context) (map[domain.MailJobStatus]int64, error) {
	rows, err := r.pool.Query(ctx, `SELECT status, COUNT(*) FROM mail_jobs GROUP BY status`)
	if err != nil {
		return nil, fmt.Errorf("count mail jobs by status: %w", err)
	}
	defer rows.Close()

	counts := map[domain.MailJobStatus]int64{
		domain.MailJobStatusQueued:     0,
		domain.MailJobStatusRetrying:   0,
		domain.MailJobStatusProcessing: 0,
		domain.MailJobStatusSent:       0,
		domain.MailJobStatusDeadLetter: 0,
	}

	for rows.Next() {
		var status domain.MailJobStatus
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan mail job counts: %w", err)
		}
		counts[status] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mail job counts: %w", err)
	}

	return counts, nil
}

func (r *MailJobRepository) OperationalSnapshot(ctx context.Context) (domain.MailOperationalSnapshot, error) {
	snapshot := domain.MailOperationalSnapshot{
		CountsByStatus: map[domain.MailJobStatus]int64{
			domain.MailJobStatusQueued:     0,
			domain.MailJobStatusRetrying:   0,
			domain.MailJobStatusProcessing: 0,
			domain.MailJobStatusSent:       0,
			domain.MailJobStatusDeadLetter: 0,
		},
		DeadLettersByErrorCode: map[string]int64{},
	}

	rows, err := r.pool.Query(ctx, `SELECT status, COUNT(*) FROM mail_jobs GROUP BY status`)
	if err != nil {
		return domain.MailOperationalSnapshot{}, fmt.Errorf("count mail jobs by status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status domain.MailJobStatus
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return domain.MailOperationalSnapshot{}, fmt.Errorf("scan mail job counts: %w", err)
		}
		snapshot.CountsByStatus[status] = count
	}
	if err := rows.Err(); err != nil {
		return domain.MailOperationalSnapshot{}, fmt.Errorf("iterate mail job counts: %w", err)
	}

	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(EXTRACT(EPOCH FROM (NOW() - next_attempt_at))), 0)
		FROM mail_jobs
		WHERE status IN ('queued', 'retrying') AND next_attempt_at <= NOW()
	`).Scan(&snapshot.OldestReadyAgeSeconds)
	if err != nil {
		return domain.MailOperationalSnapshot{}, fmt.Errorf("select oldest ready age: %w", err)
	}

	deadRows, err := r.pool.Query(ctx, `
		SELECT COALESCE(last_error_code, 'unknown') AS error_code, COUNT(*)
		FROM mail_jobs
		WHERE status = 'dead_letter'
		GROUP BY COALESCE(last_error_code, 'unknown')
	`)
	if err != nil {
		return domain.MailOperationalSnapshot{}, fmt.Errorf("count dead letters by error code: %w", err)
	}
	defer deadRows.Close()

	for deadRows.Next() {
		var errorCode string
		var count int64
		if err := deadRows.Scan(&errorCode, &count); err != nil {
			return domain.MailOperationalSnapshot{}, fmt.Errorf("scan dead letter counts: %w", err)
		}
		snapshot.DeadLettersByErrorCode[errorCode] = count
	}
	if err := deadRows.Err(); err != nil {
		return domain.MailOperationalSnapshot{}, fmt.Errorf("iterate dead letter counts: %w", err)
	}

	return snapshot, nil
}

func (r *MailJobRepository) ListDeadLetters(ctx context.Context, limit int) ([]domain.MailJob, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, message_type, recipient_email, idempotency_key, payload, status, attempts, max_attempts,
		       next_attempt_at, leased_until, last_error, last_error_code, provider, provider_message_id,
		       sent_at, dead_lettered_at, created_at, updated_at
		FROM mail_jobs
		WHERE status = 'dead_letter'
		ORDER BY dead_lettered_at DESC NULLS LAST, updated_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list dead letter jobs: %w", err)
	}
	defer rows.Close()

	return scanMailJobs(rows)
}

func (r *MailJobRepository) RequeueDeadLetter(ctx context.Context, jobID string) (domain.MailJob, error) {
	query := `
		UPDATE mail_jobs
		SET
			status = 'queued',
			leased_until = NULL,
			last_error = NULL,
			last_error_code = NULL,
			dead_lettered_at = NULL,
			next_attempt_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND status = 'dead_letter'
		RETURNING
			id, message_type, recipient_email, idempotency_key, payload, status, attempts, max_attempts,
			next_attempt_at, leased_until, last_error, last_error_code, provider, provider_message_id,
			sent_at, dead_lettered_at, created_at, updated_at`

	var job domain.MailJob
	if err := r.pool.QueryRow(ctx, query, jobID).Scan(
		&job.ID,
		&job.MessageType,
		&job.RecipientEmail,
		&job.IdempotencyKey,
		&job.Payload,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.NextAttemptAt,
		&job.LeasedUntil,
		&job.LastError,
		&job.LastErrorCode,
		&job.Provider,
		&job.ProviderMessageID,
		&job.SentAt,
		&job.DeadLetteredAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	); err != nil {
		return domain.MailJob{}, fmt.Errorf("requeue dead letter job: %w", err)
	}

	return job, nil
}

func (r *MailJobRepository) FindByID(ctx context.Context, jobID string) (domain.MailJob, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, message_type, recipient_email, idempotency_key, payload, status, attempts, max_attempts,
		       next_attempt_at, leased_until, last_error, last_error_code, provider, provider_message_id,
		       sent_at, dead_lettered_at, created_at, updated_at
		FROM mail_jobs
		WHERE id = $1
	`, jobID)

	var job domain.MailJob
	if err := row.Scan(
		&job.ID,
		&job.MessageType,
		&job.RecipientEmail,
		&job.IdempotencyKey,
		&job.Payload,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.NextAttemptAt,
		&job.LeasedUntil,
		&job.LastError,
		&job.LastErrorCode,
		&job.Provider,
		&job.ProviderMessageID,
		&job.SentAt,
		&job.DeadLetteredAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	); err != nil {
		return domain.MailJob{}, fmt.Errorf("find mail job by id: %w", err)
	}

	return job, nil
}

func pgInterval(d time.Duration) string {
	return fmt.Sprintf("%f seconds", d.Seconds())
}

func scanMailJobs(rows pgx.Rows) ([]domain.MailJob, error) {
	jobs := make([]domain.MailJob, 0)
	for rows.Next() {
		var job domain.MailJob
		if err := rows.Scan(
			&job.ID,
			&job.MessageType,
			&job.RecipientEmail,
			&job.IdempotencyKey,
			&job.Payload,
			&job.Status,
			&job.Attempts,
			&job.MaxAttempts,
			&job.NextAttemptAt,
			&job.LeasedUntil,
			&job.LastError,
			&job.LastErrorCode,
			&job.Provider,
			&job.ProviderMessageID,
			&job.SentAt,
			&job.DeadLetteredAt,
			&job.CreatedAt,
			&job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan mail job: %w", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mail jobs: %w", err)
	}

	return jobs, nil
}
