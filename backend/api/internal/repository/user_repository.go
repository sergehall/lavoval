package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sergehall/lavoval/backend/api/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	query := `
		INSERT INTO users (id, email, password_hash, role, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`
	if err := r.pool.QueryRow(ctx, query, user.ID, user.Email, user.PasswordHash, user.Role, user.Status).Scan(&user.CreatedAt, &user.UpdatedAt); err != nil {
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, status, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, status, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`

	var user domain.User
	if err := r.pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, status, created_at, updated_at
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
		if err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}
