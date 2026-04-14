package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/sergehall/lavoval/backend/api/internal/auth"
	"github.com/sergehall/lavoval/backend/api/internal/config"
	"github.com/sergehall/lavoval/backend/api/internal/domain"
	"github.com/sergehall/lavoval/backend/api/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	users    repository.UserStore
	profiles repository.ProfileStore
	tokens   auth.TokenManager
	cfg      config.Config
}

type AuthPayload struct {
	AccessToken  string      `json:"accessToken"`
	RefreshToken string      `json:"refreshToken"`
	User         SessionUser `json:"user"`
}

type SessionUser struct {
	ID        string      `json:"id"`
	Email     string      `json:"email"`
	Role      domain.Role `json:"role"`
	FirstName string      `json:"firstName,omitempty"`
	LastName  string      `json:"lastName,omitempty"`
}

type RegisterInput struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=12"`
	FirstName string `json:"firstName" validate:"required,min=2"`
	LastName  string `json:"lastName" validate:"required,min=2"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func NewAuthService(users repository.UserStore, profiles repository.ProfileStore, tokens auth.TokenManager, cfg config.Config) *AuthService {
	return &AuthService{users: users, profiles: profiles, tokens: tokens, cfg: cfg}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (AuthPayload, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthPayload{}, fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{
		ID:           uuid.NewString(),
		Email:        input.Email,
		PasswordHash: string(hash),
		Role:         domain.RoleUser,
		Status:       domain.AccountStatusActive,
	}

	createdUser, err := s.users.Create(ctx, user)
	if err != nil {
		return AuthPayload{}, fmt.Errorf("create user: %w", err)
	}

	profile := domain.Profile{
		UserID:    createdUser.ID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Timezone:  "UTC",
	}

	createdProfile, err := s.profiles.Create(ctx, profile)
	if err != nil {
		return AuthPayload{}, fmt.Errorf("create profile: %w", err)
	}

	return s.buildAuthPayload(createdUser, createdProfile)
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (AuthPayload, error) {
	user, err := s.users.FindByEmail(ctx, input.Email)
	if err != nil {
		return AuthPayload{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return AuthPayload{}, ErrInvalidCredentials
	}

	profile, err := s.profiles.FindByUserID(ctx, user.ID)
	if err != nil {
		return AuthPayload{}, fmt.Errorf("find profile: %w", err)
	}

	return s.buildAuthPayload(user, profile)
}

func (s *AuthService) buildAuthPayload(user domain.User, profile domain.Profile) (AuthPayload, error) {
	pair, err := s.tokens.IssueTokens(user)
	if err != nil {
		return AuthPayload{}, fmt.Errorf("issue tokens: %w", err)
	}

	return AuthPayload{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User: SessionUser{
			ID:        user.ID,
			Email:     user.Email,
			Role:      user.Role,
			FirstName: profile.FirstName,
			LastName:  profile.LastName,
		},
	}, nil
}
