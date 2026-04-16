package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/auth"
	"github.com/sergehall/lavoval/apps/api/internal/config"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type routerUserStoreStub struct {
	users map[string]domain.User
}

func (s *routerUserStoreStub) Create(_ context.Context, u domain.User) (domain.User, error) {
	s.users[u.ID] = u
	return u, nil
}

func (s *routerUserStoreStub) FindByEmail(_ context.Context, email string) (domain.User, error) {
	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return domain.User{}, nil
}

func (s *routerUserStoreStub) FindByID(_ context.Context, id string) (domain.User, error) {
	return s.users[id], nil
}

func (s *routerUserStoreStub) MarkEmailVerified(_ context.Context, id string) (domain.User, error) {
	user := s.users[id]
	now := time.Now().UTC()
	user.EmailVerifiedAt = &now
	s.users[id] = user
	return user, nil
}

func (s *routerUserStoreStub) UpdatePasswordHash(_ context.Context, id string, hash string) (domain.User, error) {
	user := s.users[id]
	user.PasswordHash = hash
	s.users[id] = user
	return user, nil
}

func (s *routerUserStoreStub) UpdateRoleAndStatus(_ context.Context, id string, role domain.Role, status domain.AccountStatus) (domain.User, error) {
	user := s.users[id]
	user.Role = role
	user.Status = status
	s.users[id] = user
	return user, nil
}

func (s *routerUserStoreStub) UpdateRoleStatusModeration(_ context.Context, id, actorID string, role domain.Role, status domain.AccountStatus, reason *string) (domain.User, error) {
	user := s.users[id]
	user.Role = role
	user.Status = status
	switch status {
	case domain.AccountStatusSuspended:
		user.SuspensionReason = reason
		user.SuspendedBy = &actorID
	case domain.AccountStatusBlocked:
		user.BlockReason = reason
		user.BlockedBy = &actorID
	}
	s.users[id] = user
	return user, nil
}

func (s *routerUserStoreStub) BumpSessionVersion(_ context.Context, id string) (domain.User, error) {
	user := s.users[id]
	user.SessionVersion++
	s.users[id] = user
	return user, nil
}

func (s *routerUserStoreStub) GetStats(_ context.Context) (domain.AdminUserStats, error) {
	return domain.AdminUserStats{}, nil
}

func (s *routerUserStoreStub) StartTOTPEnrollment(_ context.Context, id string, pending string) (domain.User, error) {
	user := s.users[id]
	user.MFAPendingTOTPSecretEncrypted = &pending
	s.users[id] = user
	return user, nil
}

func (s *routerUserStoreStub) CancelTOTPEnrollment(_ context.Context, id string) (domain.User, error) {
	user := s.users[id]
	user.MFAPendingTOTPSecretEncrypted = nil
	s.users[id] = user
	return user, nil
}

func (s *routerUserStoreStub) EnableTOTP(_ context.Context, id string, secret string) (domain.User, error) {
	user := s.users[id]
	user.MFAEnabled = true
	user.MFATOTPSecretEncrypted = &secret
	user.MFAPendingTOTPSecretEncrypted = nil
	s.users[id] = user
	return user, nil
}

func (s *routerUserStoreStub) DisableTOTP(_ context.Context, id string) (domain.User, error) {
	user := s.users[id]
	user.MFAEnabled = false
	user.MFATOTPSecretEncrypted = nil
	user.MFAPendingTOTPSecretEncrypted = nil
	s.users[id] = user
	return user, nil
}

func (s *routerUserStoreStub) SoftDelete(_ context.Context, id string) error {
	user := s.users[id]
	user.UpdatedAt = time.Now().UTC()
	s.users[id] = user
	return nil
}

func (s *routerUserStoreStub) List(_ context.Context) ([]domain.User, error) {
	items := make([]domain.User, 0, len(s.users))
	for _, user := range s.users {
		items = append(items, user)
	}
	return items, nil
}

func testRouterConfig() config.Config {
	return config.Config{
		AppName:       "Lavoval",
		AppEnv:        "test",
		JWTIssuer:     "test",
		JWTAudience:   "test",
		JWTSecret:     "super-secret",
		JWTAccessTTL:  time.Minute,
		JWTRefreshTTL: time.Hour,
	}
}

func TestRouterRoleChangeRevokesOldAccessToken(t *testing.T) {
	now := time.Now().UTC()
	users := &routerUserStoreStub{users: map[string]domain.User{
		"root-1": {
			ID:              "root-1",
			Email:           "root@example.com",
			Role:            domain.RoleRootOwner,
			Status:          domain.AccountStatusActive,
			EmailVerifiedAt: &now,
			MFAEnabled:      true,
			SessionVersion:  1,
		},
		"user-1": {
			ID:              "user-1",
			Email:           "user@example.com",
			Role:            domain.RoleUser,
			Status:          domain.AccountStatusActive,
			EmailVerifiedAt: &now,
			MFAEnabled:      true,
			SessionVersion:  1,
		},
	}}

	tokens := auth.NewTokenManager(testRouterConfig())
	adminService := service.NewAdminService(
		users,
		hAdminProfileStub{},
		hAdminSkillStub{},
		hAdminEnrollmentStub{},
		hAdminModuleStub{},
		nil,
		nil,
		nil,
		nil,
		service.MailRetentionPolicy{},
		service.WithSessionRevoker(service.NewUserSessionRevoker(users)),
	)

	router := NewRouter(
		testRouterConfig(),
		tokens,
		users,
		&service.AuthService{},
		nil,
		nil,
		nil,
		nil,
		adminService,
		nil,
	)

	rootPair, err := tokens.IssueTokens(users.users["root-1"])
	if err != nil {
		t.Fatalf("issue root token: %v", err)
	}
	userPair, err := tokens.IssueTokens(users.users["user-1"])
	if err != nil {
		t.Fatalf("issue user token: %v", err)
	}

	roleChangeReq := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/admin/users/user-1/role",
		bytes.NewBufferString(`{"role":"admin","reason":"Promoting operator"}`),
	)
	roleChangeReq.Header.Set("Authorization", "Bearer "+rootPair.AccessToken)
	roleChangeReq.Header.Set("Content-Type", "application/json")
	roleChangeRec := httptest.NewRecorder()

	router.ServeHTTP(roleChangeRec, roleChangeReq)

	if roleChangeRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from role change, got %d", roleChangeRec.Code)
	}

	protectedReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	protectedReq.Header.Set("Authorization", "Bearer "+userPair.AccessToken)
	protectedRec := httptest.NewRecorder()

	router.ServeHTTP(protectedRec, protectedReq)

	if protectedRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for revoked access token, got %d", protectedRec.Code)
	}

	updatedUser := users.users["user-1"]
	if updatedUser.Role != domain.RoleAdmin {
		t.Fatalf("expected updated role admin, got %s", updatedUser.Role)
	}
	if updatedUser.SessionVersion != 2 {
		t.Fatalf("expected session version bumped to 2, got %d", updatedUser.SessionVersion)
	}
}
