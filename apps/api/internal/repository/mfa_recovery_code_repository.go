package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type MFARecoveryCodeRepository struct {
	pool *pgxpool.Pool
}

func NewMFARecoveryCodeRepository(pool *pgxpool.Pool) *MFARecoveryCodeRepository {
	return &MFARecoveryCodeRepository{pool: pool}
}

func (r *MFARecoveryCodeRepository) ReplaceForUser(ctx context.Context, userID string, codes []domain.MFARecoveryCode) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin recovery code transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `UPDATE mfa_recovery_codes SET consumed_at = NOW() WHERE user_id = $1 AND consumed_at IS NULL`, userID); err != nil {
		return fmt.Errorf("revoke active recovery codes: %w", err)
	}

	for _, code := range codes {
		if _, err := tx.Exec(ctx,
			`INSERT INTO mfa_recovery_codes (id, user_id, code_hash) VALUES ($1, $2, $3)`,
			code.ID, code.UserID, code.CodeHash,
		); err != nil {
			return fmt.Errorf("insert recovery code: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit recovery code transaction: %w", err)
	}

	return nil
}

func (r *MFARecoveryCodeRepository) FindActiveByCodeHash(ctx context.Context, userID string, codeHash string) (domain.MFARecoveryCode, error) {
	var code domain.MFARecoveryCode
	if err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, code_hash, consumed_at, created_at
		FROM mfa_recovery_codes
		WHERE user_id = $1 AND code_hash = $2 AND consumed_at IS NULL
	`, userID, codeHash).Scan(
		&code.ID,
		&code.UserID,
		&code.CodeHash,
		&code.ConsumedAt,
		&code.CreatedAt,
	); err != nil {
		return domain.MFARecoveryCode{}, fmt.Errorf("find recovery code: %w", err)
	}

	return code, nil
}

func (r *MFARecoveryCodeRepository) Consume(ctx context.Context, id string, userID string) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE mfa_recovery_codes
		SET consumed_at = NOW()
		WHERE id = $1 AND user_id = $2 AND consumed_at IS NULL
	`, id, userID)
	if err != nil {
		return fmt.Errorf("consume recovery code: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("consume recovery code: no rows affected")
	}
	return nil
}

func (r *MFARecoveryCodeRepository) RevokeActiveByUserID(ctx context.Context, userID string) error {
	if _, err := r.pool.Exec(ctx, `
		UPDATE mfa_recovery_codes
		SET consumed_at = NOW()
		WHERE user_id = $1 AND consumed_at IS NULL
	`, userID); err != nil {
		return fmt.Errorf("revoke active recovery codes: %w", err)
	}
	return nil
}

var _ MFARecoveryCodeStore = (*MFARecoveryCodeRepository)(nil)
