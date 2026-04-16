package mailer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/config"
)

func TestWebhookCleanupAlertNotifierSendsPayload(t *testing.T) {
	var calls int32
	payloadCh := make(chan cleanupAlertPayload, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		defer r.Body.Close()

		var payload cleanupAlertPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		payloadCh <- payload
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	notifier := newCleanupAlertNotifier(config.Config{
		AppName:                  "Lavoval",
		AppEnv:                   "test",
		MailAlertWebhookURL:      server.URL,
		MailAlertWebhookTimeout:  2 * time.Second,
		MailAlertWebhookCooldown: 0,
	}, nil)

	if notifier == nil {
		t.Fatal("expected notifier")
	}

	err := notifier.Notify(context.Background(), cleanupAlert{
		Type:            "cleanup_failed",
		Severity:        "critical",
		Message:         "Mail retention cleanup failed",
		Mode:            "apply",
		CandidateJobs:   12,
		CandidateEvents: 34,
		ErrorMessage:    "db unavailable",
	})
	if err != nil {
		t.Fatalf("notify returned error: %v", err)
	}

	select {
	case payload := <-payloadCh:
		if payload.Type != "cleanup_failed" {
			t.Fatalf("expected type cleanup_failed, got %s", payload.Type)
		}
		if payload.CandidateJobs != 12 || payload.CandidateEvents != 34 {
			t.Fatalf("unexpected candidate counts: %+v", payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected webhook payload")
	}
}

func TestWebhookCleanupAlertNotifierSuppressesWithinCooldown(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	notifier := newCleanupAlertNotifier(config.Config{
		AppName:                  "Lavoval",
		AppEnv:                   "test",
		MailAlertWebhookURL:      server.URL,
		MailAlertWebhookTimeout:  2 * time.Second,
		MailAlertWebhookCooldown: time.Hour,
	}, nil)

	alert := cleanupAlert{
		Type:            "cleanup_backlog_threshold_exceeded",
		Severity:        "warning",
		Message:         "Backlog too large",
		Mode:            "dry_run",
		DryRun:          true,
		CandidateJobs:   100,
		CandidateEvents: 1000,
	}

	if err := notifier.Notify(context.Background(), alert); err != nil {
		t.Fatalf("first notify returned error: %v", err)
	}
	if err := notifier.Notify(context.Background(), alert); err != nil {
		t.Fatalf("second notify returned error: %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected one webhook call due to cooldown, got %d", got)
	}
}
