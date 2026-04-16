package handler

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/sergehall/lavoval/apps/api/internal/auth"
	"github.com/sergehall/lavoval/apps/api/internal/config"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type routerUserStoreStub struct {
	users map[string]domain.User
}

type routerProfileStoreStub struct {
	profiles map[string]domain.Profile
}

func (s *routerProfileStoreStub) Create(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	if s.profiles == nil {
		s.profiles = map[string]domain.Profile{}
	}
	s.profiles[profile.UserID] = profile
	return profile, nil
}

func (s *routerProfileStoreStub) Update(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	s.profiles[profile.UserID] = profile
	return profile, nil
}

func (s *routerProfileStoreStub) FindByUserID(_ context.Context, userID string) (domain.Profile, error) {
	return s.profiles[userID], nil
}

func (s *routerProfileStoreStub) SoftDeleteByUserID(_ context.Context, userID string) error {
	delete(s.profiles, userID)
	return nil
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
		AppName:               "Lavoval",
		AppEnv:                "test",
		AppURL:                "http://localhost:3000",
		JWTIssuer:             "test",
		JWTAudience:           "test",
		JWTSecret:             "super-secret",
		JWTAccessTTL:          time.Minute,
		JWTRefreshTTL:         time.Hour,
		MFATOTPPeriod:         30 * time.Second,
		MFATOTPIssuer:         "Lavoval",
		MFASecretKey:          "test-mfa-secret",
		MFASignInChallengeTTL: 10 * time.Minute,
	}
}

type routerEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code    string         `json:"code"`
		Message string         `json:"message"`
		Meta    map[string]any `json:"meta"`
	} `json:"error"`
}

func issueRouterTOTPCode(t *testing.T, secret string, at time.Time, period time.Duration) string {
	t.Helper()

	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		t.Fatalf("decode base32 secret: %v", err)
	}

	step := int64(period.Seconds())
	if step <= 0 {
		step = 30
	}
	counter := at.UTC().Unix() / step

	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(counter))

	mac := hmac.New(sha1.New, decoded)
	if _, err := mac.Write(msg); err != nil {
		t.Fatalf("write totp hmac: %v", err)
	}

	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1])&0xff)<<16 |
		(int(sum[offset+2])&0xff)<<8 |
		(int(sum[offset+3]) & 0xff)

	return fmt.Sprintf("%06d", value%1000000)
}

func encryptRouterMFASecret(t *testing.T, cfg config.Config, secret string) string {
	t.Helper()

	secretSource := cfg.MFASecretKey
	if secretSource == "" {
		secretSource = cfg.JWTSecret + ":mfa"
	}
	key := sha256.Sum256([]byte(secretSource))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		t.Fatalf("init cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("init gcm: %v", err)
	}

	nonce := bytes.Repeat([]byte{1}, gcm.NonceSize())
	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext)
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

func TestRouterLoginReturnsMFAChallengeMeta(t *testing.T) {
	cfg := testRouterConfig()
	now := time.Now().UTC()

	manager := auth.NewTokenManager(cfg)
	secret := "JBSWY3DPEHPK3PXP"
	encryptedSecret := encryptRouterMFASecret(t, cfg, secret)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	users := &routerUserStoreStub{users: map[string]domain.User{
		"user-1": {
			ID:                     "user-1",
			Email:                  "user@example.com",
			PasswordHash:           string(passwordHash),
			Role:                   domain.RoleUser,
			Status:                 domain.AccountStatusActive,
			EmailVerifiedAt:        &now,
			MFAEnabled:             true,
			MFATOTPSecretEncrypted: &encryptedSecret,
			SessionVersion:         1,
		},
	}}
	profiles := &routerProfileStoreStub{profiles: map[string]domain.Profile{
		"user-1": {UserID: "user-1", FirstName: "Ada", LastName: "Lovelace", Timezone: "UTC"},
	}}
	authService := service.NewAuthService(
		users,
		profiles,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		manager,
		nil,
		nil,
		cfg,
	)
	router := NewRouter(cfg, manager, users, authService, nil, nil, nil, nil, nil, nil)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		bytes.NewBufferString(`{"email":"user@example.com","password":"password123"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}

	var envelope routerEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error == nil {
		t.Fatal("expected error payload")
	}
	if envelope.Error.Code != "mfa_required" {
		t.Fatalf("expected mfa_required code, got %q", envelope.Error.Code)
	}
	challengeID, ok := envelope.Error.Meta["challengeId"].(string)
	if !ok || challengeID == "" {
		t.Fatalf("expected challengeId meta, got %#v", envelope.Error.Meta)
	}
}

func TestRouterDisableMFARevokesOldAccessToken(t *testing.T) {
	cfg := testRouterConfig()
	now := time.Now().UTC()
	tokens := auth.NewTokenManager(cfg)
	secret := "JBSWY3DPEHPK3PXP"
	encryptedSecret := encryptRouterMFASecret(t, cfg, secret)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	users := &routerUserStoreStub{users: map[string]domain.User{
		"user-1": {
			ID:                     "user-1",
			Email:                  "user@example.com",
			PasswordHash:           string(passwordHash),
			Role:                   domain.RoleUser,
			Status:                 domain.AccountStatusActive,
			EmailVerifiedAt:        &now,
			MFAEnabled:             true,
			MFATOTPSecretEncrypted: &encryptedSecret,
			MFAEnrolledAt:          &now,
			SessionVersion:         1,
		},
	}}
	authService := service.NewAuthService(
		users,
		&routerProfileStoreStub{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		tokens,
		nil,
		service.NewUserSessionRevoker(users),
		cfg,
	)
	router := NewRouter(cfg, tokens, users, authService, nil, nil, nil, nil, nil, nil)

	userPair, err := tokens.IssueTokens(users.users["user-1"])
	if err != nil {
		t.Fatalf("issue user token: %v", err)
	}

	code := issueRouterTOTPCode(t, secret, time.Now(), cfg.MFATOTPPeriod)
	disableReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/mfa/disable",
		bytes.NewBufferString(fmt.Sprintf(`{"password":"password123","code":"%s"}`, code)),
	)
	disableReq.Header.Set("Authorization", "Bearer "+userPair.AccessToken)
	disableReq.Header.Set("Content-Type", "application/json")
	disableRec := httptest.NewRecorder()

	router.ServeHTTP(disableRec, disableReq)

	if disableRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from mfa disable, got %d", disableRec.Code)
	}

	protectedReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	protectedReq.Header.Set("Authorization", "Bearer "+userPair.AccessToken)
	protectedRec := httptest.NewRecorder()

	router.ServeHTTP(protectedRec, protectedReq)

	if protectedRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for revoked token after mfa disable, got %d", protectedRec.Code)
	}

	updatedUser := users.users["user-1"]
	if updatedUser.MFAEnabled {
		t.Fatal("expected MFA to be disabled")
	}
	if updatedUser.SessionVersion != 2 {
		t.Fatalf("expected session version bumped to 2, got %d", updatedUser.SessionVersion)
	}
}
