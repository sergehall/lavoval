package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sergehall/lavoval/apps/api/internal/config"
)

func TestPublicHandlerIndexRendersAPIPage(t *testing.T) {
	handler := NewPublicHandler(config.Config{
		AppName: "Lavoval",
		AppEnv:  "production",
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Host = "api.lavoval.com"
	recorder := httptest.NewRecorder()

	handler.Index(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("expected html content type, got %q", contentType)
	}

	body := recorder.Body.String()
	for _, snippet := range []string{
		"Lavoval API",
		"/livez",
		"/healthz",
		"/readyz",
		"/metrics",
		"/api/v1/auth",
		"/api/v1/admin",
		"Quick start",
		"Authentication",
		"Public skills",
		"Runtime",
		"api.lavoval.com",
		"api v1",
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("expected body to contain %q", snippet)
		}
	}
}
