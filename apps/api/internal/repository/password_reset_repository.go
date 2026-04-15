package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type PasswordResetRepository struct {
	pool *pgxpool.Pool
}

func NewPasswordResetRepository(pool *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{pool: pool}
}

func (r *PasswordResetRepository) Create(ctx context.Context, token domain.PasswordResetToken) (domain.PasswordResetToken, error) {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	if err := r.pool.QueryRow(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt).Scan(&token.CreatedAt); err != nil {
		return domain.PasswordResetToken{}, fmt.Errorf("insert password reset token: %w", err)
	}

	return token, nil
}

func (r *PasswordResetRepository) FindByTokenHash(ctx context.Context, tokenHash string) (domain.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, consumed_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1`

	var token domain.PasswordResetToken
	if err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.ConsumedAt,
		&token.CreatedAt,
	); err != nil {
		return domain.PasswordResetToken{}, fmt.Errorf("find password reset token: %w", err)
	}

	return token, nil
}

func (r *PasswordResetRepository) Consume(ctx context.Context, tokenID string, userID string) error {
	query := `
		UPDATE password_reset_tokens
		SET consumed_at = NOW()
		WHERE id = $1 AND user_id = $2 AND consumed_at IS NULL`

	result, err := r.pool.Exec(ctx, query, tokenID, userID)
	if err != nil {
		return fmt.Errorf("consume password reset token: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("consume password reset token: no rows affected")
	}
	return nil
}

func (r *PasswordResetRepository) RevokeActiveByUserID(ctx context.Context, userID string) error {
	query := `
		UPDATE password_reset_tokens
		SET consumed_at = NOW()
		WHERE user_id = $1 AND consumed_at IS NULL`

	if _, err := r.pool.Exec(ctx, query, userID); err != nil {
		return fmt.Errorf("revoke active password reset tokens: %w", err)
	}
	return nil
}
