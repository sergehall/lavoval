package service

import (
	"context"
	"fmt"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

const securityTimeLayout = time.RFC3339

type AccountSecuritySummary struct {
	HasPassword       bool          `json:"hasPassword"`
	PasswordUpdatedAt *string       `json:"passwordUpdatedAt,omitempty"`
	Providers         []OAuthMethod `json:"providers"`
	MFAEnabled        bool          `json:"mfaEnabled"`
	MFAEnrolledAt     *string       `json:"mfaEnrolledAt,omitempty"`
}

type OAuthMethod struct {
	Provider    domain.OAuthProvider `json:"provider"`
	Connected   bool                 `json:"connected"`
	ConnectedAt string               `json:"connectedAt"`
}

type AccountSecurityService struct {
	users           repository.UserStore
	oauthIdentities repository.OAuthIdentityStore
}

func NewAccountSecurityService(
	users repository.UserStore,
	oauthIdentities repository.OAuthIdentityStore,
) *AccountSecurityService {
	return &AccountSecurityService{
		users:           users,
		oauthIdentities: oauthIdentities,
	}
}

func (s *AccountSecurityService) Summary(ctx context.Context, userID string) (AccountSecuritySummary, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return AccountSecuritySummary{}, fmt.Errorf("find user: %w", err)
	}

	identities, err := s.oauthIdentities.ListByUserID(ctx, userID)
	if err != nil {
		return AccountSecuritySummary{}, fmt.Errorf("list oauth identities: %w", err)
	}

	providers := make([]OAuthMethod, 0, len(identities))
	for _, identity := range identities {
		providers = append(providers, OAuthMethod{
			Provider:    identity.Provider,
			Connected:   true,
			ConnectedAt: identity.CreatedAt.Format(securityTimeLayout),
		})
	}

	summary := AccountSecuritySummary{
		HasPassword: user.PasswordHash != "",
		Providers:   providers,
		MFAEnabled:  user.MFAEnabled,
	}

	if user.MFAEnrolledAt != nil {
		value := user.MFAEnrolledAt.Format(securityTimeLayout)
		summary.MFAEnrolledAt = &value
	}

	return summary, nil
}
