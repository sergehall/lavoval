package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/apps/api/internal/auth"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/httpx"
	appmiddleware "github.com/sergehall/lavoval/apps/api/internal/middleware"
	"github.com/sergehall/lavoval/apps/api/internal/service"
)

type meHandlerProfileRepoStub struct {
	updateErr    error
	updatedInput domain.Profile
}

func (s *meHandlerProfileRepoStub) Create(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	return profile, nil
}

func (s *meHandlerProfileRepoStub) Update(_ context.Context, profile domain.Profile) (domain.Profile, error) {
	s.updatedInput = profile
	if s.updateErr != nil {
		return domain.Profile{}, s.updateErr
	}
	return profile, nil
}

func (s *meHandlerProfileRepoStub) FindByUserID(_ context.Context, _ string) (domain.Profile, error) {
	return domain.Profile{}, nil
}

func (s *meHandlerProfileRepoStub) SoftDeleteByUserID(_ context.Context, _ string) error {
	return nil
}

func TestMeHandlerUpdateProfileRejectsInvalidJSON(t *testing.T) {
	// nil service is safe: decode failure exits before the service is called
	h := NewMeHandler(validator.New(validator.WithRequiredStructEnabled()), nil, nil)

	r := httptest.NewRequest(http.MethodPut, "/api/v1/me/profile", bytes.NewBufferString(`{bad json}`))
	w := httptest.NewRecorder()
	h.UpdateProfile(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestMeHandlerUpdateProfileRejectsMissingRequiredFields(t *testing.T) {
	// nil service is safe: validation failure exits before the service is called
	h := NewMeHandler(validator.New(validator.WithRequiredStructEnabled()), nil, nil)

	// firstName and timezone are required but absent
	r := httptest.NewRequest(http.MethodPut, "/api/v1/me/profile", bytes.NewBufferString(`{"lastName":"Lovelace"}`))
	w := httptest.NewRecorder()
	h.UpdateProfile(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing required fields, got %d", w.Code)
	}
}

func TestMeHandlerUpdateProfileReturnsUpdatedProfile(t *testing.T) {
	repo := &meHandlerProfileRepoStub{}
	h := NewMeHandler(
		validator.New(validator.WithRequiredStructEnabled()),
		service.NewProfileService(repo),
		nil,
	)

	body := `{
		"firstName":"Serge",
		"lastName":"Hall",
		"timezone":"America/Los_Angeles",
		"username":"sergehall",
		"skills":["TypeScript","Go"],
		"languages":["en","be"],
		"availabilityStatus":"open",
		"isPublicProfile":true
	}`

	r := httptest.NewRequest(http.MethodPut, "/api/v1/me/profile", bytes.NewBufferString(body))
	r = r.WithContext(appmiddleware.WithClaims(r.Context(), &auth.Claims{UserID: "user-123"}))
	w := httptest.NewRecorder()

	h.UpdateProfile(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for successful update, got %d", w.Code)
	}

	var envelope httpx.Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("expected valid JSON response, got %v", err)
	}
	if envelope.Error != nil {
		t.Fatalf("expected no error in response, got %+v", envelope.Error)
	}

	payload, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatalf("expected serializable response payload, got %v", err)
	}

	var profile domain.Profile
	if err := json.Unmarshal(payload, &profile); err != nil {
		t.Fatalf("expected profile payload, got %v", err)
	}

	if profile.UserID != "user-123" {
		t.Fatalf("expected userID user-123, got %s", profile.UserID)
	}
	if profile.FirstName != "Serge" {
		t.Fatalf("expected firstName Serge, got %s", profile.FirstName)
	}
	if len(repo.updatedInput.Skills) != 2 || repo.updatedInput.Skills[0] != "TypeScript" {
		t.Fatalf("expected skills to reach service layer, got %v", repo.updatedInput.Skills)
	}
	if !repo.updatedInput.IsPublicProfile {
		t.Fatalf("expected public profile flag to be true")
	}
}

func TestMeHandlerUpdateProfileRejectsUnsafeAvatarURL(t *testing.T) {
	repo := &meHandlerProfileRepoStub{}
	h := NewMeHandler(
		validator.New(validator.WithRequiredStructEnabled()),
		service.NewProfileService(repo),
		nil,
	)

	body := `{
		"firstName":"Serge",
		"lastName":"Hall",
		"timezone":"America/Los_Angeles",
		"avatarUrl":"https://user:pass@avatars.githubusercontent.com/u/60080971"
	}`

	r := httptest.NewRequest(http.MethodPut, "/api/v1/me/profile", bytes.NewBufferString(body))
	r = r.WithContext(appmiddleware.WithClaims(r.Context(), &auth.Claims{UserID: "user-123"}))
	w := httptest.NewRecorder()

	h.UpdateProfile(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unsafe avatar URL, got %d", w.Code)
	}
	if repo.updatedInput.UserID != "" {
		t.Fatalf("expected repository not to be called, got %+v", repo.updatedInput)
	}
}

func TestMeHandlerUpdateProfileReturnsInternalErrorWhenServiceFails(t *testing.T) {
	repo := &meHandlerProfileRepoStub{updateErr: errors.New("db unavailable")}
	h := NewMeHandler(
		validator.New(validator.WithRequiredStructEnabled()),
		service.NewProfileService(repo),
		nil,
	)

	r := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/me/profile",
		bytes.NewBufferString(`{"firstName":"Serge","lastName":"Hall","timezone":"UTC"}`),
	)
	r = r.WithContext(appmiddleware.WithClaims(r.Context(), &auth.Claims{UserID: "user-456"}))
	w := httptest.NewRecorder()

	h.UpdateProfile(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when service update fails, got %d", w.Code)
	}

	var envelope httpx.Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("expected valid JSON response, got %v", err)
	}
	if envelope.Error == nil {
		t.Fatal("expected error response body, got nil")
	}
	if envelope.Error.Code != "profile_update_failed" {
		t.Fatalf("expected error code profile_update_failed, got %s", envelope.Error.Code)
	}
}
