package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/sergehall/lavoval/apps/api/internal/config"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type OAuthStartResponse struct {
	URL string `json:"url"`
}

type CompleteGoogleOAuthInput struct {
	Code  string `json:"code" validate:"required,min=8"`
	State string `json:"state" validate:"required,min=16"`
}

type CompleteGitHubOAuthInput struct {
	Code  string `json:"code" validate:"required,min=8"`
	State string `json:"state" validate:"required,min=16"`
}

type oauthIdentityProfile struct {
	Subject       string
	Email         string
	EmailVerified bool
	GivenName     string
	FamilyName    string
	FullName      string
}

type googleOAuthProvider interface {
	Enabled() bool
	AuthorizationURL(state string) string
	ExchangeCode(context.Context, string) (oauthIdentityProfile, error)
}

type githubOAuthProvider interface {
	Enabled() bool
	AuthorizationURL(state string) string
	ExchangeCode(context.Context, string) (oauthIdentityProfile, error)
}

type liveGoogleOAuthProvider struct {
	client       *http.Client
	clientID     string
	clientSecret string
	redirectURL  string
}

type liveGitHubOAuthProvider struct {
	client       *http.Client
	clientID     string
	clientSecret string
	redirectURL  string
}

func newGoogleOAuthProvider(cfg config.Config) googleOAuthProvider {
	return &liveGoogleOAuthProvider{
		client:       &http.Client{Timeout: 15 * time.Second},
		clientID:     cfg.GoogleOAuthClientID,
		clientSecret: cfg.GoogleOAuthSecret,
		redirectURL:  strings.TrimRight(cfg.AppURL, "/") + "/auth/oauth/google/callback",
	}
}

func newGitHubOAuthProvider(cfg config.Config) githubOAuthProvider {
	return &liveGitHubOAuthProvider{
		client:       &http.Client{Timeout: 15 * time.Second},
		clientID:     cfg.GitHubOAuthClientID,
		clientSecret: cfg.GitHubOAuthSecret,
		redirectURL:  strings.TrimRight(cfg.AppURL, "/") + "/auth/oauth/github/callback",
	}
}

func (p *liveGoogleOAuthProvider) Enabled() bool {
	return p.clientID != "" && p.clientSecret != ""
}

func (p *liveGitHubOAuthProvider) Enabled() bool {
	return p.clientID != "" && p.clientSecret != ""
}

func (p *liveGoogleOAuthProvider) AuthorizationURL(state string) string {
	query := url.Values{}
	query.Set("client_id", p.clientID)
	query.Set("redirect_uri", p.redirectURL)
	query.Set("response_type", "code")
	query.Set("scope", "openid email profile")
	query.Set("access_type", "offline")
	query.Set("include_granted_scopes", "true")
	query.Set("prompt", "consent")
	query.Set("state", state)

	return "https://accounts.google.com/o/oauth2/v2/auth?" + query.Encode()
}

func (p *liveGitHubOAuthProvider) AuthorizationURL(state string) string {
	query := url.Values{}
	query.Set("client_id", p.clientID)
	query.Set("redirect_uri", p.redirectURL)
	query.Set("scope", "read:user user:email")
	query.Set("state", state)

	return "https://github.com/login/oauth/authorize?" + query.Encode()
}

func (p *liveGoogleOAuthProvider) ExchangeCode(ctx context.Context, code string) (oauthIdentityProfile, error) {
	form := url.Values{}
	form.Set("client_id", p.clientID)
	form.Set("client_secret", p.clientSecret)
	form.Set("redirect_uri", p.redirectURL)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://oauth2.googleapis.com/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("build google token request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := p.client.Do(request)
	if err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("exchange google oauth code: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return oauthIdentityProfile{}, fmt.Errorf("google token exchange failed: %s", strings.TrimSpace(string(body)))
	}

	var tokenPayload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&tokenPayload); err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("decode google token response: %w", err)
	}
	if tokenPayload.AccessToken == "" {
		return oauthIdentityProfile{}, fmt.Errorf("google token exchange returned no access token")
	}

	userInfoRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://openidconnect.googleapis.com/v1/userinfo",
		nil,
	)
	if err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("build google userinfo request: %w", err)
	}
	userInfoRequest.Header.Set("Authorization", "Bearer "+tokenPayload.AccessToken)

	userInfoResponse, err := p.client.Do(userInfoRequest)
	if err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("fetch google userinfo: %w", err)
	}
	defer userInfoResponse.Body.Close()

	if userInfoResponse.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(userInfoResponse.Body, 2048))
		return oauthIdentityProfile{}, fmt.Errorf("google userinfo request failed: %s", strings.TrimSpace(string(body)))
	}

	var userInfo struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Name          string `json:"name"`
	}
	if err := json.NewDecoder(userInfoResponse.Body).Decode(&userInfo); err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("decode google userinfo: %w", err)
	}

	return oauthIdentityProfile{
		Subject:       userInfo.Sub,
		Email:         strings.TrimSpace(userInfo.Email),
		EmailVerified: userInfo.EmailVerified,
		GivenName:     strings.TrimSpace(userInfo.GivenName),
		FamilyName:    strings.TrimSpace(userInfo.FamilyName),
		FullName:      strings.TrimSpace(userInfo.Name),
	}, nil
}

func (p *liveGitHubOAuthProvider) ExchangeCode(ctx context.Context, code string) (oauthIdentityProfile, error) {
	form := url.Values{}
	form.Set("client_id", p.clientID)
	form.Set("client_secret", p.clientSecret)
	form.Set("redirect_uri", p.redirectURL)
	form.Set("code", code)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("build github token request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "lavoval-app")

	response, err := p.client.Do(request)
	if err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("exchange github oauth code: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return oauthIdentityProfile{}, fmt.Errorf("github token exchange failed: %s", strings.TrimSpace(string(body)))
	}

	var tokenPayload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&tokenPayload); err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("decode github token response: %w", err)
	}
	if tokenPayload.AccessToken == "" {
		return oauthIdentityProfile{}, fmt.Errorf("github token exchange returned no access token")
	}

	profileRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("build github user request: %w", err)
	}
	profileRequest.Header.Set("Authorization", "Bearer "+tokenPayload.AccessToken)
	profileRequest.Header.Set("Accept", "application/vnd.github+json")
	profileRequest.Header.Set("User-Agent", "lavoval-app")

	profileResponse, err := p.client.Do(profileRequest)
	if err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("fetch github user: %w", err)
	}
	defer profileResponse.Body.Close()

	if profileResponse.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(profileResponse.Body, 2048))
		return oauthIdentityProfile{}, fmt.Errorf("github user request failed: %s", strings.TrimSpace(string(body)))
	}

	var profile struct {
		ID    int64   `json:"id"`
		Name  string  `json:"name"`
		Login string  `json:"login"`
		Email *string `json:"email"`
	}
	if err := json.NewDecoder(profileResponse.Body).Decode(&profile); err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("decode github user: %w", err)
	}

	emailsRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("build github emails request: %w", err)
	}
	emailsRequest.Header.Set("Authorization", "Bearer "+tokenPayload.AccessToken)
	emailsRequest.Header.Set("Accept", "application/vnd.github+json")
	emailsRequest.Header.Set("User-Agent", "lavoval-app")

	emailsResponse, err := p.client.Do(emailsRequest)
	if err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("fetch github emails: %w", err)
	}
	defer emailsResponse.Body.Close()

	if emailsResponse.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(io.LimitReader(emailsResponse.Body, 2048))
		return oauthIdentityProfile{}, fmt.Errorf("github emails request failed: %s", strings.TrimSpace(string(body)))
	}

	var emails []struct {
		Email      string `json:"email"`
		Primary    bool   `json:"primary"`
		Verified   bool   `json:"verified"`
		Visibility string `json:"visibility"`
	}
	if err := json.NewDecoder(emailsResponse.Body).Decode(&emails); err != nil {
		return oauthIdentityProfile{}, fmt.Errorf("decode github emails: %w", err)
	}

	email, verified := resolveGitHubVerifiedEmail(emails, profile.Email)

	return oauthIdentityProfile{
		Subject:       strconv.FormatInt(profile.ID, 10),
		Email:         email,
		EmailVerified: verified,
		FullName:      strings.TrimSpace(firstNonEmpty(profile.Name, profile.Login)),
	}, nil
}

func (s *AuthService) GoogleOAuthStart(ctx context.Context) (OAuthStartResponse, error) {
	return s.oauthStart(ctx, domain.OAuthProviderGoogle, s.googleOAuth)
}

func (s *AuthService) GitHubOAuthStart(ctx context.Context) (OAuthStartResponse, error) {
	return s.oauthStart(ctx, domain.OAuthProviderGitHub, s.githubOAuth)
}

func (s *AuthService) oauthStart(
	ctx context.Context,
	provider domain.OAuthProvider,
	oauthProvider interface {
		Enabled() bool
		AuthorizationURL(string) string
	},
) (OAuthStartResponse, error) {
	if oauthProvider == nil || !oauthProvider.Enabled() {
		return OAuthStartResponse{}, ErrOAuthNotConfigured
	}

	state, stateHash, err := newOAuthStateToken()
	if err != nil {
		return OAuthStartResponse{}, err
	}

	record := domain.OAuthState{
		ID:        uuid.NewString(),
		Provider:  provider,
		StateHash: stateHash,
		ExpiresAt: time.Now().Add(s.cfg.OAuthStateTTL),
	}

	if _, err := s.oauthStates.Create(ctx, record); err != nil {
		return OAuthStartResponse{}, fmt.Errorf("store oauth state: %w", err)
	}

	return OAuthStartResponse{URL: oauthProvider.AuthorizationURL(state)}, nil
}

func (s *AuthService) CompleteGoogleOAuth(ctx context.Context, input CompleteGoogleOAuthInput) (AuthPayload, error) {
	return s.completeOAuth(ctx, domain.OAuthProviderGoogle, s.googleOAuth, input.Code, input.State)
}

func (s *AuthService) CompleteGitHubOAuth(ctx context.Context, input CompleteGitHubOAuthInput) (AuthPayload, error) {
	return s.completeOAuth(ctx, domain.OAuthProviderGitHub, s.githubOAuth, input.Code, input.State)
}

func (s *AuthService) completeOAuth(
	ctx context.Context,
	provider domain.OAuthProvider,
	oauthProvider interface {
		Enabled() bool
		ExchangeCode(context.Context, string) (oauthIdentityProfile, error)
	},
	code string,
	statePlain string,
) (AuthPayload, error) {
	if oauthProvider == nil || !oauthProvider.Enabled() {
		return AuthPayload{}, ErrOAuthNotConfigured
	}

	stateHash := hashVerificationToken(statePlain)
	state, err := s.oauthStates.FindByStateHash(ctx, provider, stateHash)
	if err != nil {
		return AuthPayload{}, ErrOAuthStateInvalid
	}
	if state.ConsumedAt != nil {
		return AuthPayload{}, ErrOAuthStateInvalid
	}
	if time.Now().After(state.ExpiresAt) {
		return AuthPayload{}, ErrOAuthStateExpired
	}
	if err := s.oauthStates.Consume(ctx, state.ID); err != nil {
		return AuthPayload{}, fmt.Errorf("consume oauth state: %w", err)
	}

	identity, err := oauthProvider.ExchangeCode(ctx, code)
	if err != nil {
		return AuthPayload{}, fmt.Errorf("exchange oauth code: %w", err)
	}
	if !identity.EmailVerified || identity.Email == "" || identity.Subject == "" {
		return AuthPayload{}, ErrOAuthEmailNotVerified
	}

	user, profile, err := s.resolveOAuthUser(ctx, provider, identity)
	if err != nil {
		return AuthPayload{}, err
	}
	if user.MFAEnabled {
		return AuthPayload{}, ErrOAuthMFASignInNotSupported
	}

	return s.buildAuthPayload(user, profile)
}

func (s *AuthService) resolveOAuthUser(
	ctx context.Context,
	provider domain.OAuthProvider,
	identity oauthIdentityProfile,
) (domain.User, domain.Profile, error) {
	oauthIdentity, err := s.oauthIdentities.FindByProviderSubject(ctx, provider, identity.Subject)
	if err == nil {
		user, err := s.users.FindByID(ctx, oauthIdentity.UserID)
		if err != nil {
			return domain.User{}, domain.Profile{}, fmt.Errorf("find oauth user: %w", err)
		}
		profile, err := s.profiles.FindByUserID(ctx, user.ID)
		if err != nil {
			return domain.User{}, domain.Profile{}, fmt.Errorf("find oauth profile: %w", err)
		}
		return user, profile, nil
	}
	if !isNotFoundError(err) {
		return domain.User{}, domain.Profile{}, fmt.Errorf("find oauth identity: %w", err)
	}

	user, err := s.users.FindByEmail(ctx, identity.Email)
	switch {
	case err == nil:
		if user.EmailVerifiedAt == nil {
			user, err = s.users.MarkEmailVerified(ctx, user.ID)
			if err != nil {
				return domain.User{}, domain.Profile{}, fmt.Errorf("mark oauth email verified: %w", err)
			}
		}
		profile, err := s.profiles.FindByUserID(ctx, user.ID)
		if err != nil {
			return domain.User{}, domain.Profile{}, fmt.Errorf("find existing oauth profile: %w", err)
		}
		if _, err := s.oauthIdentities.Create(ctx, domain.OAuthIdentity{
			ID:             uuid.NewString(),
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: identity.Subject,
			Email:          identity.Email,
		}); err != nil && !isUniqueConstraintError(err) {
			return domain.User{}, domain.Profile{}, fmt.Errorf("create oauth identity for existing user: %w", err)
		}
		return user, profile, nil
	case !isNotFoundError(err):
		return domain.User{}, domain.Profile{}, fmt.Errorf("find user by oauth email: %w", err)
	}

	passwordHash, err := randomOAuthPasswordHash()
	if err != nil {
		return domain.User{}, domain.Profile{}, fmt.Errorf("generate oauth password hash: %w", err)
	}

	verifiedAt := time.Now()
	user = domain.User{
		ID:              uuid.NewString(),
		Email:           identity.Email,
		PasswordHash:    passwordHash,
		Role:            domain.RoleUser,
		Status:          domain.AccountStatusActive,
		EmailVerifiedAt: &verifiedAt,
	}

	user, err = s.users.Create(ctx, user)
	if err != nil {
		return domain.User{}, domain.Profile{}, fmt.Errorf("create oauth user: %w", err)
	}

	firstName, lastName := splitOAuthName(identity)
	profile := domain.Profile{
		UserID:    user.ID,
		FirstName: firstName,
		LastName:  lastName,
		Timezone:  "UTC",
	}

	profile, err = s.profiles.Create(ctx, profile)
	if err != nil {
		return domain.User{}, domain.Profile{}, fmt.Errorf("create oauth profile: %w", err)
	}

	if _, err := s.oauthIdentities.Create(ctx, domain.OAuthIdentity{
		ID:             uuid.NewString(),
		UserID:         user.ID,
		Provider:       provider,
		ProviderUserID: identity.Subject,
		Email:          identity.Email,
	}); err != nil {
		return domain.User{}, domain.Profile{}, fmt.Errorf("create oauth identity: %w", err)
	}

	return user, profile, nil
}

func resolveGitHubVerifiedEmail(
	emails []struct {
		Email      string `json:"email"`
		Primary    bool   `json:"primary"`
		Verified   bool   `json:"verified"`
		Visibility string `json:"visibility"`
	},
	fallback *string,
) (string, bool) {
	for _, entry := range emails {
		if entry.Primary && entry.Verified && strings.TrimSpace(entry.Email) != "" {
			return strings.TrimSpace(entry.Email), true
		}
	}
	for _, entry := range emails {
		if entry.Verified && strings.TrimSpace(entry.Email) != "" {
			return strings.TrimSpace(entry.Email), true
		}
	}
	if fallback != nil && strings.TrimSpace(*fallback) != "" {
		return strings.TrimSpace(*fallback), false
	}
	return "", false
}

func newOAuthStateToken() (plain string, hashed string, err error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", "", fmt.Errorf("generate oauth state: %w", err)
	}
	plain = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buffer)
	return plain, hashVerificationToken(plain), nil
}

func randomOAuthPasswordHash() (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(base32.StdEncoding.EncodeToString(buffer)), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func splitOAuthName(identity oauthIdentityProfile) (firstName string, lastName string) {
	if identity.GivenName != "" {
		firstName = identity.GivenName
	} else if identity.FullName != "" {
		parts := strings.Fields(identity.FullName)
		if len(parts) > 0 {
			firstName = parts[0]
			if len(parts) > 1 {
				lastName = strings.Join(parts[1:], " ")
			}
		}
	}
	if identity.FamilyName != "" {
		lastName = identity.FamilyName
	}
	if firstName == "" {
		firstName = "Lavoval"
	}
	if lastName == "" {
		lastName = "Member"
	}
	return firstName, lastName
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func isNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), pgx.ErrNoRows.Error())
}

func isUniqueConstraintError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}
