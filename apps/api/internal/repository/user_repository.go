package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

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
		RETURNING mfa_enabled, mfa_totp_secret_encrypted, mfa_pending_totp_secret_encrypted, mfa_enrolled_at, created_at, updated_at`
	if err := r.pool.QueryRow(ctx, query, user.ID, user.Email, user.PasswordHash, user.Role, user.Status, user.EmailVerifiedAt).Scan(
		&user.MFAEnabled,
		&user.MFATOTPSecretEncrypted,
		&user.MFAPendingTOTPSecretEncrypted,
		&user.MFAEnrolledAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, status, email_verified_at, mfa_enabled, mfa_totp_secret_encrypted, mfa_pending_totp_secret_encrypted, mfa_enrolled_at, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.EmailVerifiedAt,
		&user.MFAEnabled,
		&user.MFATOTPSecretEncrypted,
		&user.MFAPendingTOTPSecretEncrypted,
		&user.MFAEnrolledAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, status, email_verified_at, mfa_enabled, mfa_totp_secret_encrypted, mfa_pending_totp_secret_encrypted, mfa_enrolled_at, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.EmailVerifiedAt,
		&user.MFAEnabled,
		&user.MFATOTPSecretEncrypted,
		&user.MFAPendingTOTPSecretEncrypted,
		&user.MFAEnrolledAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepository) MarkEmailVerified(ctx context.Context, userID string) (domain.User, error) {
	query := `
		UPDATE users
		SET email_verified_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, role, status, email_verified_at, mfa_enabled, mfa_totp_secret_encrypted, mfa_pending_totp_secret_encrypted, mfa_enrolled_at, created_at, updated_at`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.EmailVerifiedAt,
		&user.MFAEnabled,
		&user.MFATOTPSecretEncrypted,
		&user.MFAPendingTOTPSecretEncrypted,
		&user.MFAEnrolledAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return domain.User{}, fmt.Errorf("mark email verified: %w", err)
	}

	return user, nil
}

func (r *UserRepository) UpdatePasswordHash(ctx context.Context, userID string, passwordHash string) (domain.User, error) {
	query := `
		UPDATE users
		SET password_hash = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, role, status, email_verified_at, mfa_enabled, mfa_totp_secret_encrypted, mfa_pending_totp_secret_encrypted, mfa_enrolled_at, created_at, updated_at`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, userID, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.EmailVerifiedAt,
		&user.MFAEnabled,
		&user.MFATOTPSecretEncrypted,
		&user.MFAPendingTOTPSecretEncrypted,
		&user.MFAEnrolledAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return domain.User{}, fmt.Errorf("update password hash: %w", err)
	}

	return user, nil
}

func (r *UserRepository) StartTOTPEnrollment(ctx context.Context, id string, pendingSecretEncrypted string) (domain.User, error) {
	query := `
		UPDATE users
		SET mfa_pending_totp_secret_encrypted = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, role, status, email_verified_at, mfa_enabled, mfa_totp_secret_encrypted, mfa_pending_totp_secret_encrypted, mfa_enrolled_at, created_at, updated_at`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, id, pendingSecretEncrypted).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.EmailVerifiedAt,
		&user.MFAEnabled,
		&user.MFATOTPSecretEncrypted,
		&user.MFAPendingTOTPSecretEncrypted,
		&user.MFAEnrolledAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return domain.User{}, fmt.Errorf("start totp enrollment: %w", err)
	}

	return user, nil
}

func (r *UserRepository) EnableTOTP(ctx context.Context, id string, secretEncrypted string) (domain.User, error) {
	query := `
		UPDATE users
		SET
			mfa_enabled = TRUE,
			mfa_totp_secret_encrypted = $2,
			mfa_pending_totp_secret_encrypted = NULL,
			mfa_enrolled_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, role, status, email_verified_at, mfa_enabled, mfa_totp_secret_encrypted, mfa_pending_totp_secret_encrypted, mfa_enrolled_at, created_at, updated_at`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, id, secretEncrypted).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.EmailVerifiedAt,
		&user.MFAEnabled,
		&user.MFATOTPSecretEncrypted,
		&user.MFAPendingTOTPSecretEncrypted,
		&user.MFAEnrolledAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return domain.User{}, fmt.Errorf("enable totp: %w", err)
	}

	return user, nil
}

func (r *UserRepository) DisableTOTP(ctx context.Context, id string) (domain.User, error) {
	query := `
		UPDATE users
		SET
			mfa_enabled = FALSE,
			mfa_totp_secret_encrypted = NULL,
			mfa_pending_totp_secret_encrypted = NULL,
			mfa_enrolled_at = NULL,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, role, status, email_verified_at, mfa_enabled, mfa_totp_secret_encrypted, mfa_pending_totp_secret_encrypted, mfa_enrolled_at, created_at, updated_at`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.EmailVerifiedAt,
		&user.MFAEnabled,
		&user.MFATOTPSecretEncrypted,
		&user.MFAPendingTOTPSecretEncrypted,
		&user.MFAEnrolledAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return domain.User{}, fmt.Errorf("disable totp: %w", err)
	}

	return user, nil
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, status, email_verified_at, mfa_enabled, mfa_totp_secret_encrypted, mfa_pending_totp_secret_encrypted, mfa_enrolled_at, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.Status,
			&user.EmailVerifiedAt,
			&user.MFAEnabled,
			&user.MFATOTPSecretEncrypted,
			&user.MFAPendingTOTPSecretEncrypted,
			&user.MFAEnrolledAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}
