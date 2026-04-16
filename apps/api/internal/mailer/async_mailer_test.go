package mailer

import (
	"context"
	"errors"
	"net/textproto"
	"testing"
	"time"
)

func TestClassifyDeliveryErrorUsesTypedDeliveryError(t *testing.T) {
	retryable, code := classifyDeliveryError(temporaryDeliveryError("smtp_dial_failed", errors.New("dial failed")))
	if !retryable {
		t.Fatalf("expected typed temporary delivery error to be retryable")
	}
	if code != "smtp_dial_failed" {
		t.Fatalf("expected code smtp_dial_failed, got %s", code)
	}

	retryable, code = classifyDeliveryError(permanentDeliveryError("render_failed", errors.New("template boom")))
	if retryable {
		t.Fatalf("expected typed permanent delivery error to be non-retryable")
	}
	if code != "render_failed" {
		t.Fatalf("expected code render_failed, got %s", code)
	}
}

func TestClassifyDeliveryErrorHandlesContextDeadline(t *testing.T) {
	retryable, code := classifyDeliveryError(context.DeadlineExceeded)
	if !retryable {
		t.Fatalf("expected deadline exceeded to be retryable")
	}
	if code != "deadline_exceeded" {
		t.Fatalf("expected deadline_exceeded code, got %s", code)
	}
}

func TestRetryBackoffCapsAtThirtySeconds(t *testing.T) {
	if got := retryBackoff(time.Second, 1); got != time.Second {
		t.Fatalf("attempt 1: got %s want %s", got, time.Second)
	}

	if got := retryBackoff(time.Second, 3); got != 4*time.Second {
		t.Fatalf("attempt 3: got %s want %s", got, 4*time.Second)
	}

	if got := retryBackoff(5*time.Second, 5); got != 30*time.Second {
		t.Fatalf("expected capped backoff, got %s", got)
	}
}

func TestBuildVerificationMessageAddsStableHeaders(t *testing.T) {
	message, err := buildVerificationMessage("job-123", VerificationEmail{
		ToEmail:     "serge@example.com",
		ToName:      "Serge",
		VerifyURL:   "https://lavoval.test/verify-email?token=abc123",
		ProductName: "Lavoval",
	}, "https://lavoval.test")
	if err != nil {
		t.Fatalf("buildVerificationMessage returned error: %v", err)
	}

	assertHeaderValue(t, message.Headers, "X-Lavoval-Message-ID", "job-123")
	assertHeaderValue(t, message.Headers, "X-Lavoval-Message-Type", mailJobTypeVerification)
	if message.IdempotencyKey != "job-123" {
		t.Fatalf("expected idempotency key to reuse job id, got %s", message.IdempotencyKey)
	}
}

func TestMailIdempotencyKeyIsStable(t *testing.T) {
	payload := []byte(`{"verifyURL":"https://lavoval.test/verify-email?token=abc123"}`)

	first := mailIdempotencyKey(mailJobTypeVerification, "Serge@example.com", payload)
	second := mailIdempotencyKey(mailJobTypeVerification, "serge@example.com", payload)

	if first != second {
		t.Fatalf("expected stable idempotency key, got %q and %q", first, second)
	}
}

func assertHeaderValue(t *testing.T, headers textproto.MIMEHeader, key string, want string) {
	t.Helper()

	if got := headers.Get(key); got != want {
		t.Fatalf("%s: got %q want %q", key, got, want)
	}
}
