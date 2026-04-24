package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type OAuthIdentityRepository struct {
	pool *pgxpool.Pool
}

func NewOAuthIdentityRepository(pool *pgxpool.Pool) *OAuthIdentityRepository {
	return &OAuthIdentityRepository{pool: pool}
}

func (r *OAuthIdentityRepository) Create(ctx context.Context, identity domain.OAuthIdentity) (domain.OAuthIdentity, error) {
	query := `
		INSERT INTO lavoval_oauth_identities (id, user_id, provider, provider_user_id, email)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`

	if err := r.pool.QueryRow(
		ctx,
		query,
		identity.ID,
		identity.UserID,
		identity.Provider,
		identity.ProviderUserID,
		identity.Email,
	).Scan(&identity.CreatedAt, &identity.UpdatedAt); err != nil {
		return domain.OAuthIdentity{}, fmt.Errorf("insert oauth identity: %w", err)
	}

	return identity, nil
}

func (r *OAuthIdentityRepository) FindByProviderSubject(
	ctx context.Context,
	provider domain.OAuthProvider,
	providerUserID string,
) (domain.OAuthIdentity, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, email, created_at, updated_at
		FROM lavoval_oauth_identities
		WHERE provider = $1 AND provider_user_id = $2`

	var identity domain.OAuthIdentity
	if err := r.pool.QueryRow(ctx, query, provider, providerUserID).Scan(
		&identity.ID,
		&identity.UserID,
		&identity.Provider,
		&identity.ProviderUserID,
		&identity.Email,
		&identity.CreatedAt,
		&identity.UpdatedAt,
	); err != nil {
		return domain.OAuthIdentity{}, fmt.Errorf("find oauth identity: %w", err)
	}

	return identity, nil
}

func (r *OAuthIdentityRepository) ListByUserID(ctx context.Context, userID string) ([]domain.OAuthIdentity, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, email, created_at, updated_at
		FROM lavoval_oauth_identities
		WHERE user_id = $1
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list oauth identities: %w", err)
	}
	defer rows.Close()

	identities := make([]domain.OAuthIdentity, 0)
	for rows.Next() {
		var identity domain.OAuthIdentity
		if err := rows.Scan(
			&identity.ID,
			&identity.UserID,
			&identity.Provider,
			&identity.ProviderUserID,
			&identity.Email,
			&identity.CreatedAt,
			&identity.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan oauth identity: %w", err)
		}

		identities = append(identities, identity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate oauth identities: %w", err)
	}

	return identities, nil
}
