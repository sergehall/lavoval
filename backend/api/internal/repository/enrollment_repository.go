package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/backend/api/internal/domain"
)

type EnrollmentRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) *EnrollmentRepository {
	return &EnrollmentRepository{pool: pool}
}

func (r *EnrollmentRepository) ListByUserID(ctx context.Context, userID string) ([]domain.Enrollment, error) {
	query := `
		SELECT id, user_id, skill_id, status, progress_percent, assigned_at, completed_at
		FROM user_skill_enrollments
		WHERE user_id = $1
		ORDER BY assigned_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list enrollments: %w", err)
	}
	defer rows.Close()

	enrollments := make([]domain.Enrollment, 0)
	for rows.Next() {
		var enrollment domain.Enrollment
		if err := rows.Scan(&enrollment.ID, &enrollment.UserID, &enrollment.SkillID, &enrollment.Status, &enrollment.ProgressPercent, &enrollment.AssignedAt, &enrollment.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan enrollment: %w", err)
		}
		enrollments = append(enrollments, enrollment)
	}
	return enrollments, rows.Err()
}
