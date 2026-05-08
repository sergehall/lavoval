package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoginIPThrottleSlowsAndLimitsByIP(t *testing.T) {
	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	delays := make([]time.Duration, 0, 2)
	calls := 0

	throttle := newLoginIPThrottle(LoginIPThrottleConfig{
		Enabled:       true,
		Window:        time.Minute,
		MaxAttempts:   4,
		SlowAfter:     2,
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      250 * time.Millisecond,
		BlockDuration: 5 * time.Minute,
		StateTTL:      10 * time.Minute,
	}, func() time.Time {
		return now
	}, func(_ context.Context, delay time.Duration) bool {
		delays = append(delays, delay)
		return true
	})

	handler := throttle.middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := 0; i < 4; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.RemoteAddr = "203.0.113.10:44321"

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("attempt %d: expected 204, got %d", i+1, rec.Code)
		}
	}

	if calls != 4 {
		t.Fatalf("expected downstream handler to be called 4 times, got %d", calls)
	}
	if len(delays) != 2 {
		t.Fatalf("expected 2 delayed attempts, got %d", len(delays))
	}
	if delays[0] != 100*time.Millisecond || delays[1] != 200*time.Millisecond {
		t.Fatalf("unexpected delays: %v", delays)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "203.0.113.10:44321"

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after max attempts, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
	if calls != 4 {
		t.Fatalf("limited request should not reach downstream handler, got %d calls", calls)
	}
}

func TestLoginIPThrottleTracksDifferentIPsSeparately(t *testing.T) {
	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	calls := 0

	throttle := newLoginIPThrottle(LoginIPThrottleConfig{
		Enabled:       true,
		Window:        time.Minute,
		MaxAttempts:   1,
		SlowAfter:     1,
		BaseDelay:     time.Millisecond,
		MaxDelay:      time.Millisecond,
		BlockDuration: time.Minute,
		StateTTL:      time.Minute,
	}, func() time.Time {
		return now
	}, func(_ context.Context, _ time.Duration) bool {
		return true
	})

	handler := throttle.middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))

	first := httptest.NewRecorder()
	firstReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	firstReq.RemoteAddr = "203.0.113.10:5000"
	handler.ServeHTTP(first, firstReq)
	if first.Code != http.StatusNoContent {
		t.Fatalf("expected first IP to pass, got %d", first.Code)
	}

	second := httptest.NewRecorder()
	secondReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	secondReq.RemoteAddr = "203.0.113.11:5000"
	handler.ServeHTTP(second, secondReq)
	if second.Code != http.StatusNoContent {
		t.Fatalf("expected second IP to pass independently, got %d", second.Code)
	}

	limited := httptest.NewRecorder()
	limitedReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	limitedReq.RemoteAddr = "203.0.113.10:5001"
	handler.ServeHTTP(limited, limitedReq)
	if limited.Code != http.StatusTooManyRequests {
		t.Fatalf("expected repeated first IP to be limited, got %d", limited.Code)
	}

	if calls != 2 {
		t.Fatalf("expected two downstream calls, got %d", calls)
	}
}

func TestLoginIPThrottleResetsAfterWindow(t *testing.T) {
	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	calls := 0

	throttle := newLoginIPThrottle(LoginIPThrottleConfig{
		Enabled:       true,
		Window:        time.Minute,
		MaxAttempts:   1,
		SlowAfter:     1,
		BaseDelay:     time.Millisecond,
		MaxDelay:      time.Millisecond,
		BlockDuration: time.Minute,
		StateTTL:      3 * time.Minute,
	}, func() time.Time {
		return now
	}, func(_ context.Context, _ time.Duration) bool {
		return true
	})

	handler := throttle.middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "203.0.113.10:5000"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected initial attempt to pass, got %d", rec.Code)
	}

	now = now.Add(time.Minute + time.Second)

	nextReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	nextReq.RemoteAddr = "203.0.113.10:5001"
	nextRec := httptest.NewRecorder()
	handler.ServeHTTP(nextRec, nextReq)
	if nextRec.Code != http.StatusNoContent {
		t.Fatalf("expected attempt after window reset to pass, got %d", nextRec.Code)
	}

	if calls != 2 {
		t.Fatalf("expected two downstream calls, got %d", calls)
	}
}

func TestLoginIPThrottleStopsWhenRequestContextIsCanceledDuringDelay(t *testing.T) {
	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	calls := 0

	throttle := newLoginIPThrottle(LoginIPThrottleConfig{
		Enabled:       true,
		Window:        time.Minute,
		MaxAttempts:   2,
		SlowAfter:     1,
		BaseDelay:     time.Millisecond,
		MaxDelay:      time.Millisecond,
		BlockDuration: time.Minute,
		StateTTL:      time.Minute,
	}, func() time.Time {
		return now
	}, func(ctx context.Context, _ time.Duration) bool {
		return ctx.Err() == nil
	})

	handler := throttle.middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))

	first := httptest.NewRecorder()
	firstReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	firstReq.RemoteAddr = "203.0.113.10:5000"
	handler.ServeHTTP(first, firstReq)
	if first.Code != http.StatusNoContent {
		t.Fatalf("expected initial request to pass, got %d", first.Code)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	delayed := httptest.NewRecorder()
	delayedReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil).WithContext(ctx)
	delayedReq.RemoteAddr = "203.0.113.10:5000"
	handler.ServeHTTP(delayed, delayedReq)

	if calls != 1 {
		t.Fatalf("expected canceled delayed request not to reach downstream, got %d calls", calls)
	}
}
