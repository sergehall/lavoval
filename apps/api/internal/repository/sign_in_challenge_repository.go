package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type SignInChallengeRepository struct {
	pool *pgxpool.Pool
}

func NewSignInChallengeRepository(pool *pgxpool.Pool) *SignInChallengeRepository {
	return &SignInChallengeRepository{pool: pool}
}

func (r *SignInChallengeRepository) Create(ctx context.Context, challenge domain.AuthSignInChallenge) (domain.AuthSignInChallenge, error) {
	if err := r.pool.QueryRow(ctx, `
		INSERT INTO lavoval_auth_sign_in_challenges (id, user_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING created_at
	`, challenge.ID, challenge.UserID, challenge.ExpiresAt).Scan(&challenge.CreatedAt); err != nil {
		return domain.AuthSignInChallenge{}, fmt.Errorf("create sign-in challenge: %w", err)
	}

	return challenge, nil
}

func (r *SignInChallengeRepository) FindByID(ctx context.Context, id string) (domain.AuthSignInChallenge, error) {
	var challenge domain.AuthSignInChallenge
	if err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, expires_at, consumed_at, created_at
		FROM lavoval_auth_sign_in_challenges
		WHERE id = $1
	`, id).Scan(
		&challenge.ID,
		&challenge.UserID,
		&challenge.ExpiresAt,
		&challenge.ConsumedAt,
		&challenge.CreatedAt,
	); err != nil {
		return domain.AuthSignInChallenge{}, fmt.Errorf("find sign-in challenge: %w", err)
	}

	return challenge, nil
}

func (r *SignInChallengeRepository) Consume(ctx context.Context, id string, userID string) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE lavoval_auth_sign_in_challenges
		SET consumed_at = NOW()
		WHERE id = $1 AND user_id = $2 AND consumed_at IS NULL
	`, id, userID)
	if err != nil {
		return fmt.Errorf("consume sign-in challenge: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("consume sign-in challenge: no rows affected")
	}
	return nil
}

func (r *SignInChallengeRepository) RevokeActiveByUserID(ctx context.Context, userID string) error {
	if _, err := r.pool.Exec(ctx, `
		UPDATE lavoval_auth_sign_in_challenges
		SET consumed_at = NOW()
		WHERE user_id = $1 AND consumed_at IS NULL
	`, userID); err != nil {
		return fmt.Errorf("revoke sign-in challenges: %w", err)
	}
	return nil
}
