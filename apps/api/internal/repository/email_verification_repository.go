package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type EmailVerificationRepository struct {
	pool *pgxpool.Pool
}

func NewEmailVerificationRepository(pool *pgxpool.Pool) *EmailVerificationRepository {
	return &EmailVerificationRepository{pool: pool}
}

func (r *EmailVerificationRepository) Create(ctx context.Context, token domain.EmailVerificationToken) (domain.EmailVerificationToken, error) {
	query := `
		INSERT INTO lavoval_email_verification_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	if err := r.pool.QueryRow(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt).Scan(&token.CreatedAt); err != nil {
		return domain.EmailVerificationToken{}, fmt.Errorf("insert email verification token: %w", err)
	}

	return token, nil
}

func (r *EmailVerificationRepository) FindByTokenHash(ctx context.Context, tokenHash string) (domain.EmailVerificationToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, consumed_at, created_at
		FROM lavoval_email_verification_tokens
		WHERE token_hash = $1`

	var token domain.EmailVerificationToken
	if err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.ConsumedAt,
		&token.CreatedAt,
	); err != nil {
		return domain.EmailVerificationToken{}, fmt.Errorf("find email verification token: %w", err)
	}

	return token, nil
}

func (r *EmailVerificationRepository) Consume(ctx context.Context, tokenID string, userID string) error {
	query := `
		UPDATE lavoval_email_verification_tokens
		SET consumed_at = NOW()
		WHERE id = $1 AND user_id = $2 AND consumed_at IS NULL`

	result, err := r.pool.Exec(ctx, query, tokenID, userID)
	if err != nil {
		return fmt.Errorf("consume email verification token: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("consume email verification token: no rows affected")
	}
	return nil
}

func (r *EmailVerificationRepository) RevokeActiveByUserID(ctx context.Context, userID string) error {
	query := `
		UPDATE lavoval_email_verification_tokens
		SET consumed_at = NOW()
		WHERE user_id = $1 AND consumed_at IS NULL`

	if _, err := r.pool.Exec(ctx, query, userID); err != nil {
		return fmt.Errorf("revoke active email verification tokens: %w", err)
	}

	return nil
}
