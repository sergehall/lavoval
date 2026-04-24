package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type AdminAuditLogRepository struct {
	pool *pgxpool.Pool
}

func NewAdminAuditLogRepository(pool *pgxpool.Pool) *AdminAuditLogRepository {
	return &AdminAuditLogRepository{pool: pool}
}

func (r *AdminAuditLogRepository) Create(ctx context.Context, entry domain.AdminAuditLog) (domain.AdminAuditLog, error) {
	query := `
		INSERT INTO lavoval_admin_audit_logs (id, entity_type, entity_id, action, old_value_json, new_value_json, reason, actor_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`
	if err := r.pool.QueryRow(ctx, query,
		entry.ID, entry.EntityType, entry.EntityID, entry.Action,
		entry.OldValueJSON, entry.NewValueJSON, entry.Reason, entry.ActorID,
	).Scan(&entry.CreatedAt); err != nil {
		return domain.AdminAuditLog{}, fmt.Errorf("insert audit log: %w", err)
	}
	return entry, nil
}

func (r *AdminAuditLogRepository) ListByEntity(ctx context.Context, entityType, entityID string, limit int) ([]domain.AdminAuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, entity_type, entity_id, action, old_value_json, new_value_json, reason, actor_id, created_at
		FROM lavoval_admin_audit_logs
		WHERE entity_type = $1 AND entity_id = $2::uuid
		ORDER BY created_at DESC
		LIMIT $3`

	rows, err := r.pool.Query(ctx, query, entityType, entityID, limit)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	entries := make([]domain.AdminAuditLog, 0)
	for rows.Next() {
		var e domain.AdminAuditLog
		if err := rows.Scan(
			&e.ID, &e.EntityType, &e.EntityID, &e.Action,
			&e.OldValueJSON, &e.NewValueJSON, &e.Reason, &e.ActorID, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
