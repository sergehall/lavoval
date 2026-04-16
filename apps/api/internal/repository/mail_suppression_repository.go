package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type MailSuppressionRepository struct {
	pool *pgxpool.Pool
}

func NewMailSuppressionRepository(pool *pgxpool.Pool) *MailSuppressionRepository {
	return &MailSuppressionRepository{pool: pool}
}

func (r *MailSuppressionRepository) Create(ctx context.Context, item domain.MailSuppression) (domain.MailSuppression, error) {
	if err := r.pool.QueryRow(ctx, `
		INSERT INTO mail_suppressions (id, kind, value, reason)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`, item.ID, item.Kind, normalizeSuppressionValue(item.Kind, item.Value), item.Reason).Scan(&item.CreatedAt); err != nil {
		return domain.MailSuppression{}, fmt.Errorf("insert mail suppression: %w", err)
	}
	item.Value = normalizeSuppressionValue(item.Kind, item.Value)
	return item, nil
}

func (r *MailSuppressionRepository) Delete(ctx context.Context, id string) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM mail_suppressions WHERE id = $1`, id); err != nil {
		return fmt.Errorf("delete mail suppression: %w", err)
	}
	return nil
}

func (r *MailSuppressionRepository) List(ctx context.Context) ([]domain.MailSuppression, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, kind, value, reason, created_at
		FROM mail_suppressions
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list mail suppressions: %w", err)
	}
	defer rows.Close()

	items := make([]domain.MailSuppression, 0)
	for rows.Next() {
		var item domain.MailSuppression
		if err := rows.Scan(&item.ID, &item.Kind, &item.Value, &item.Reason, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan mail suppression: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mail suppressions: %w", err)
	}
	return items, nil
}

func (r *MailSuppressionRepository) FindMatch(ctx context.Context, recipient string) (*domain.MailSuppression, error) {
	recipient = strings.ToLower(strings.TrimSpace(recipient))
	domainPart := recipient
	if at := strings.LastIndex(recipient, "@"); at >= 0 && at < len(recipient)-1 {
		domainPart = recipient[at+1:]
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, kind, value, reason, created_at
		FROM mail_suppressions
		WHERE (kind = 'email' AND value = $1)
		   OR (kind = 'domain' AND value = $2)
		ORDER BY created_at DESC
		LIMIT 1
	`, recipient, domainPart)

	var item domain.MailSuppression
	if err := row.Scan(&item.ID, &item.Kind, &item.Value, &item.Reason, &item.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find mail suppression: %w", err)
	}
	return &item, nil
}

func normalizeSuppressionValue(kind domain.MailSuppressionKind, value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if kind == domain.MailSuppressionKindDomain && strings.HasPrefix(normalized, "@") {
		return strings.TrimPrefix(normalized, "@")
	}
	return normalized
}
