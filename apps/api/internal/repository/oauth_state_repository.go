package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type OAuthStateRepository struct {
	pool *pgxpool.Pool
}

func NewOAuthStateRepository(pool *pgxpool.Pool) *OAuthStateRepository {
	return &OAuthStateRepository{pool: pool}
}

func (r *OAuthStateRepository) Create(ctx context.Context, state domain.OAuthState) (domain.OAuthState, error) {
	query := `
		INSERT INTO lavoval_oauth_states (id, provider, state_hash, expires_at, consumed_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at`

	if err := r.pool.QueryRow(ctx, query, state.ID, state.Provider, state.StateHash, state.ExpiresAt, state.ConsumedAt).
		Scan(&state.CreatedAt); err != nil {
		return domain.OAuthState{}, fmt.Errorf("insert oauth state: %w", err)
	}

	return state, nil
}

func (r *OAuthStateRepository) FindByStateHash(
	ctx context.Context,
	provider domain.OAuthProvider,
	stateHash string,
) (domain.OAuthState, error) {
	query := `
		SELECT id, provider, state_hash, expires_at, consumed_at, created_at
		FROM lavoval_oauth_states
		WHERE provider = $1 AND state_hash = $2`

	var state domain.OAuthState
	if err := r.pool.QueryRow(ctx, query, provider, stateHash).Scan(
		&state.ID,
		&state.Provider,
		&state.StateHash,
		&state.ExpiresAt,
		&state.ConsumedAt,
		&state.CreatedAt,
	); err != nil {
		return domain.OAuthState{}, fmt.Errorf("find oauth state: %w", err)
	}

	return state, nil
}

func (r *OAuthStateRepository) Consume(ctx context.Context, id string) error {
	query := `
		UPDATE lavoval_oauth_states
		SET consumed_at = NOW()
		WHERE id = $1 AND consumed_at IS NULL`

	if _, err := r.pool.Exec(ctx, query, id); err != nil {
		return fmt.Errorf("consume oauth state: %w", err)
	}

	return nil
}
