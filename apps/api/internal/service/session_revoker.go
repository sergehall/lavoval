package service

import "context"

type SessionRevoker interface {
	RevokeAllSessionsForUser(context.Context, string, string) error
}

type NoopSessionRevoker struct{}

func (NoopSessionRevoker) RevokeAllSessionsForUser(context.Context, string, string) error {
	return nil
}
