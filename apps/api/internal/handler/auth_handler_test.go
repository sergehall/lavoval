package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"

	"github.com/sergehall/lavoval/apps/api/internal/service"
)

func TestAuthHandlerRejectsInvalidRegisterPayload(t *testing.T) {
	handler := NewAuthHandler(validator.New(validator.WithRequiredStructEnabled()), &service.AuthService{})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{"email":"bad","password":"short"}`))
	response := httptest.NewRecorder()

	handler.Register(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid payload, got %d", response.Code)
	}
}
