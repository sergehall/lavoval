package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
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
		FROM lavoval_user_skill_enrollments
		WHERE user_id = $1
		ORDER BY assigned_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list enrollments: %w", err)
	}
	defer rows.Close()

	enrollments := make([]domain.Enrollment, 0)
	for rows.Next() {
		var e domain.Enrollment
		if err := rows.Scan(&e.ID, &e.UserID, &e.SkillID, &e.Status, &e.ProgressPercent, &e.AssignedAt, &e.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan enrollment: %w", err)
		}
		enrollments = append(enrollments, e)
	}
	return enrollments, rows.Err()
}

// ListAll returns every enrollment joined with user email and skill title/slug.
// All parameters are bound ($N) — no SQL injection risk.
func (r *EnrollmentRepository) ListAll(ctx context.Context) ([]domain.EnrollmentDetail, error) {
	query := `
		SELECT
			e.id, e.user_id, e.skill_id, e.status, e.progress_percent, e.assigned_at, e.completed_at,
			u.email    AS user_email,
			s.title    AS skill_title,
			s.slug     AS skill_slug
		FROM lavoval_user_skill_enrollments e
		JOIN lavoval_users  u ON u.id = e.user_id
		JOIN lavoval_skills s ON s.id = e.skill_id
		ORDER BY e.assigned_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list all enrollments: %w", err)
	}
	defer rows.Close()

	details := make([]domain.EnrollmentDetail, 0)
	for rows.Next() {
		var d domain.EnrollmentDetail
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.SkillID, &d.Status, &d.ProgressPercent, &d.AssignedAt, &d.CompletedAt,
			&d.UserEmail, &d.SkillTitle, &d.SkillSlug,
		); err != nil {
			return nil, fmt.Errorf("scan enrollment detail: %w", err)
		}
		details = append(details, d)
	}
	return details, rows.Err()
}

// Create assigns a skill to a user with status=assigned, progress=0.
// The UNIQUE constraint on (user_id, skill_id) prevents duplicate assignments.
func (r *EnrollmentRepository) Create(ctx context.Context, userID, skillID string) (domain.Enrollment, error) {
	query := `
		INSERT INTO lavoval_user_skill_enrollments (id, user_id, skill_id, status, progress_percent)
		VALUES ($1, $2, $3, 'assigned', 0)
		RETURNING id, user_id, skill_id, status, progress_percent, assigned_at, completed_at`

	var e domain.Enrollment
	if err := r.pool.QueryRow(ctx, query, uuid.NewString(), userID, skillID).Scan(
		&e.ID, &e.UserID, &e.SkillID, &e.Status, &e.ProgressPercent, &e.AssignedAt, &e.CompletedAt,
	); err != nil {
		return domain.Enrollment{}, fmt.Errorf("create enrollment: %w", err)
	}
	return e, nil
}

// UpdateStatus changes enrollment status and progress.
// Sets completed_at automatically when status transitions to 'completed'.
func (r *EnrollmentRepository) UpdateStatus(ctx context.Context, id string, status domain.EnrollmentStatus, progress int) (domain.Enrollment, error) {
	query := `
		UPDATE lavoval_user_skill_enrollments
		SET
			status           = $2,
			progress_percent = $3,
			completed_at     = CASE WHEN $2 = 'completed' THEN NOW() ELSE completed_at END,
			updated_at       = NOW()
		WHERE id = $1
		RETURNING id, user_id, skill_id, status, progress_percent, assigned_at, completed_at`

	var e domain.Enrollment
	if err := r.pool.QueryRow(ctx, query, id, status, progress).Scan(
		&e.ID, &e.UserID, &e.SkillID, &e.Status, &e.ProgressPercent, &e.AssignedAt, &e.CompletedAt,
	); err != nil {
		return domain.Enrollment{}, fmt.Errorf("update enrollment: %w", err)
	}
	return e, nil
}
