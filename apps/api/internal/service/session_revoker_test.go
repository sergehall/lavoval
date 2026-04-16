package service

import (
	"context"
	"errors"
	"testing"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type sessionRevokerUserStoreStub struct {
	authUserRepoStub
	bumpFn       func(context.Context, string) (domain.User, error)
	bumpedUserID string
}

func (s *sessionRevokerUserStoreStub) BumpSessionVersion(ctx context.Context, userID string) (domain.User, error) {
	s.bumpedUserID = userID
	if s.bumpFn != nil {
		return s.bumpFn(ctx, userID)
	}
	return s.authUserRepoStub.BumpSessionVersion(ctx, userID)
}

func TestUserSessionRevokerReturnsNilWithoutUserStore(t *testing.T) {
	revoker := NewUserSessionRevoker(nil)

	err := revoker.RevokeAllSessionsForUser(context.Background(), "user-1", "password_reset")
	if err != nil {
		t.Fatalf("expected nil error when user store is missing, got %v", err)
	}
}

func TestUserSessionRevokerBumpsUserSessionVersion(t *testing.T) {
	users := &sessionRevokerUserStoreStub{
		authUserRepoStub: authUserRepoStub{
			user: domain.User{ID: "user-1", SessionVersion: 3},
		},
	}
	revoker := NewUserSessionRevoker(users)

	err := revoker.RevokeAllSessionsForUser(context.Background(), "user-1", "admin_forced_sign_out")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if users.bumpedUserID != "user-1" {
		t.Fatalf("expected session bump for user-1, got %q", users.bumpedUserID)
	}
}

func TestUserSessionRevokerWrapsRepositoryErrors(t *testing.T) {
	repositoryErr := errors.New("write failed")
	users := &sessionRevokerUserStoreStub{
		bumpFn: func(context.Context, string) (domain.User, error) {
			return domain.User{}, repositoryErr
		},
	}
	revoker := NewUserSessionRevoker(users)

	err := revoker.RevokeAllSessionsForUser(context.Background(), "user-1", "security_event")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repositoryErr) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
	if err.Error() != "bump user session version: write failed" {
		t.Fatalf("expected wrapped error message, got %q", err.Error())
	}
}
