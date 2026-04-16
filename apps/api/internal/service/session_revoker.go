package service

import (
	"context"
	"fmt"

	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

type SessionRevoker interface {
	RevokeAllSessionsForUser(context.Context, string, string) error
}

type NoopSessionRevoker struct{}

func (NoopSessionRevoker) RevokeAllSessionsForUser(context.Context, string, string) error {
	return nil
}

type UserSessionRevoker struct {
	users repository.UserStore
}

func NewUserSessionRevoker(users repository.UserStore) UserSessionRevoker {
	return UserSessionRevoker{users: users}
}

func (r UserSessionRevoker) RevokeAllSessionsForUser(ctx context.Context, userID string, _ string) error {
	if r.users == nil {
		return nil
	}
	if _, err := r.users.BumpSessionVersion(ctx, userID); err != nil {
		return fmt.Errorf("bump user session version: %w", err)
	}
	return nil
}
