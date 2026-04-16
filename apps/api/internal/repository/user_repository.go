package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

// userCols is the canonical SELECT column list for the users table.
// All query helpers must scan exactly these columns in this order.
const userCols = `id, email, password_hash, role, status, email_verified_at,
	mfa_enabled, mfa_totp_secret_encrypted, mfa_pending_totp_secret_encrypted, mfa_enrolled_at,
	suspension_reason, suspended_at, suspended_by, block_reason, blocked_at, blocked_by,
	created_at, updated_at`

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.EmailVerifiedAt,
		&u.MFAEnabled, &u.MFATOTPSecretEncrypted, &u.MFAPendingTOTPSecretEncrypted, &u.MFAEnrolledAt,
		&u.SuspensionReason, &u.SuspendedAt, &u.SuspendedBy,
		&u.BlockReason, &u.BlockedAt, &u.BlockedBy,
		&u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	query := `
		INSERT INTO users (id, email, password_hash, role, status, email_verified_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + userCols
	row := r.pool.QueryRow(ctx, query, user.ID, user.Email, user.PasswordHash, user.Role, user.Status, user.EmailVerifiedAt)
	created, err := scanUser(row)
	if err != nil {
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}
	return created, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `SELECT ` + userCols + ` FROM users WHERE email = $1 AND deleted_at IS NULL`
	user, err := scanUser(r.pool.QueryRow(ctx, query, email))
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (domain.User, error) {
	query := `SELECT ` + userCols + ` FROM users WHERE id = $1 AND deleted_at IS NULL`
	user, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepository) MarkEmailVerified(ctx context.Context, userID string) (domain.User, error) {
	query := `
		UPDATE users SET email_verified_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + userCols
	user, err := scanUser(r.pool.QueryRow(ctx, query, userID))
	if err != nil {
		return domain.User{}, fmt.Errorf("mark email verified: %w", err)
	}
	return user, nil
}

func (r *UserRepository) UpdatePasswordHash(ctx context.Context, userID string, passwordHash string) (domain.User, error) {
	query := `
		UPDATE users SET password_hash = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + userCols
	user, err := scanUser(r.pool.QueryRow(ctx, query, userID, passwordHash))
	if err != nil {
		return domain.User{}, fmt.Errorf("update password hash: %w", err)
	}
	return user, nil
}

// UpdateRoleAndStatus is kept for backward compatibility; it does not touch moderation fields.
func (r *UserRepository) UpdateRoleAndStatus(ctx context.Context, id string, role domain.Role, status domain.AccountStatus) (domain.User, error) {
	query := `
		UPDATE users SET role = $2, status = $3, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + userCols
	user, err := scanUser(r.pool.QueryRow(ctx, query, id, role, status))
	if err != nil {
		return domain.User{}, fmt.Errorf("update user role/status: %w", err)
	}
	return user, nil
}

// UpdateRoleStatusModeration atomically updates role, status, and writes moderation audit fields
// (suspension_reason / block_reason, timestamps, actor ID) based on the target status.
func (r *UserRepository) UpdateRoleStatusModeration(ctx context.Context, id, actorID string, role domain.Role, status domain.AccountStatus, reason *string) (domain.User, error) {
	query := `
		UPDATE users
		SET
			role       = $2,
			status     = $3,
			updated_at = NOW(),
			suspension_reason = CASE WHEN $3 = 'suspended' THEN $4 ELSE NULL END,
			suspended_at      = CASE WHEN $3 = 'suspended' THEN NOW() ELSE NULL END,
			suspended_by      = CASE WHEN $3 = 'suspended' THEN $5::uuid ELSE NULL END,
			block_reason      = CASE WHEN $3 = 'blocked'   THEN $4 ELSE NULL END,
			blocked_at        = CASE WHEN $3 = 'blocked'   THEN NOW() ELSE NULL END,
			blocked_by        = CASE WHEN $3 = 'blocked'   THEN $5::uuid ELSE NULL END
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + userCols
	user, err := scanUser(r.pool.QueryRow(ctx, query, id, role, status, reason, actorID))
	if err != nil {
		return domain.User{}, fmt.Errorf("update user moderation: %w", err)
	}
	return user, nil
}

// GetStats returns aggregate user counts for the admin dashboard.
func (r *UserRepository) GetStats(ctx context.Context) (domain.AdminUserStats, error) {
	query := `
		SELECT
			COUNT(*)                                                   AS total,
			COUNT(*) FILTER (WHERE status = 'active')                 AS active,
			COUNT(*) FILTER (WHERE status = 'suspended')              AS suspended,
			COUNT(*) FILTER (WHERE status = 'blocked')                AS blocked,
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '7 days')  AS new_last_7d,
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '30 days') AS new_last_30d
		FROM users
		WHERE deleted_at IS NULL`

	var s domain.AdminUserStats
	if err := r.pool.QueryRow(ctx, query).Scan(
		&s.Total, &s.Active, &s.Suspended, &s.Blocked, &s.NewLast7d, &s.NewLast30d,
	); err != nil {
		return domain.AdminUserStats{}, fmt.Errorf("user stats: %w", err)
	}
	return s, nil
}

func (r *UserRepository) StartTOTPEnrollment(ctx context.Context, id string, pendingSecretEncrypted string) (domain.User, error) {
	query := `
		UPDATE users SET mfa_pending_totp_secret_encrypted = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + userCols
	user, err := scanUser(r.pool.QueryRow(ctx, query, id, pendingSecretEncrypted))
	if err != nil {
		return domain.User{}, fmt.Errorf("start totp enrollment: %w", err)
	}
	return user, nil
}

func (r *UserRepository) CancelTOTPEnrollment(ctx context.Context, id string) (domain.User, error) {
	query := `
		UPDATE users SET mfa_pending_totp_secret_encrypted = NULL, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + userCols
	user, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		return domain.User{}, fmt.Errorf("cancel totp enrollment: %w", err)
	}
	return user, nil
}

func (r *UserRepository) EnableTOTP(ctx context.Context, id string, secretEncrypted string) (domain.User, error) {
	query := `
		UPDATE users
		SET
			mfa_enabled                    = TRUE,
			mfa_totp_secret_encrypted      = $2,
			mfa_pending_totp_secret_encrypted = NULL,
			mfa_enrolled_at                = NOW(),
			updated_at                     = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + userCols
	user, err := scanUser(r.pool.QueryRow(ctx, query, id, secretEncrypted))
	if err != nil {
		return domain.User{}, fmt.Errorf("enable totp: %w", err)
	}
	return user, nil
}

func (r *UserRepository) DisableTOTP(ctx context.Context, id string) (domain.User, error) {
	query := `
		UPDATE users
		SET
			mfa_enabled                       = FALSE,
			mfa_totp_secret_encrypted         = NULL,
			mfa_pending_totp_secret_encrypted = NULL,
			mfa_enrolled_at                   = NULL,
			updated_at                        = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + userCols
	user, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		return domain.User{}, fmt.Errorf("disable totp: %w", err)
	}
	return user, nil
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	query := `SELECT ` + userCols + ` FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *UserRepository) SoftDelete(ctx context.Context, id string) error {
	if _, err := r.pool.Exec(ctx,
		`UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id,
	); err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	return nil
}
