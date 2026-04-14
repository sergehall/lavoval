package auth

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/sergehall/lavoval/backend/api/internal/config"
	"github.com/sergehall/lavoval/backend/api/internal/domain"
)

type Claims struct {
	UserID string      `json:"uid"`
	Email  string      `json:"email"`
	Role   domain.Role `json:"role"`
	Type   string      `json:"type"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	cfg config.Config
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func NewTokenManager(cfg config.Config) TokenManager {
	return TokenManager{cfg: cfg}
}

func (m TokenManager) IssueTokens(user domain.User) (TokenPair, error) {
	accessToken, err := m.issue(user, "access", m.cfg.JWTAccessTTL)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, err := m.issue(user, "refresh", m.cfg.JWTRefreshTTL)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (m TokenManager) issue(user domain.User, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.cfg.JWTIssuer,
			Audience:  []string{m.cfg.JWTAudience},
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(m.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signedToken, nil
}

func (m TokenManager) Parse(token string) (*Claims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(_ *jwt.Token) (any, error) {
		return []byte(m.cfg.JWTSecret), nil
	}, jwt.WithAudience(m.cfg.JWTAudience), jwt.WithIssuer(m.cfg.JWTIssuer))
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
