package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestMeHandlerUpdateProfileRejectsInvalidJSON(t *testing.T) {
	// nil service is safe: decode failure exits before the service is called
	h := NewMeHandler(validator.New(validator.WithRequiredStructEnabled()), nil)

	r := httptest.NewRequest(http.MethodPut, "/api/v1/me/profile", bytes.NewBufferString(`{bad json}`))
	w := httptest.NewRecorder()
	h.UpdateProfile(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestMeHandlerUpdateProfileRejectsMissingRequiredFields(t *testing.T) {
	// nil service is safe: validation failure exits before the service is called
	h := NewMeHandler(validator.New(validator.WithRequiredStructEnabled()), nil)

	// firstName and timezone are required but absent
	r := httptest.NewRequest(http.MethodPut, "/api/v1/me/profile", bytes.NewBufferString(`{"lastName":"Lovelace"}`))
	w := httptest.NewRecorder()
	h.UpdateProfile(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing required fields, got %d", w.Code)
	}
}
