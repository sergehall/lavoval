package middleware

import (
	"context"

	"github.com/sergehall/lavoval/apps/api/internal/auth"
)

type contextKey string

const claimsKey contextKey = "claims"

func WithClaims(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func ClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*auth.Claims)
	return claims, ok
}
