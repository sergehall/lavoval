package middleware

import (
	"context"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/httpx"
)

type LoginIPThrottleConfig struct {
	Enabled       bool
	Window        time.Duration
	MaxAttempts   int
	SlowAfter     int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	BlockDuration time.Duration
	StateTTL      time.Duration
}

type loginIPThrottle struct {
	cfg     LoginIPThrottleConfig
	mu      sync.Mutex
	entries map[string]loginIPThrottleEntry
	now     func() time.Time
	wait    func(context.Context, time.Duration) bool
}

type loginIPThrottleEntry struct {
	windowStart  time.Time
	attempts     int
	blockedUntil time.Time
	lastSeen     time.Time
}

func LoginIPThrottle(cfg LoginIPThrottleConfig) func(http.Handler) http.Handler {
	return newLoginIPThrottle(cfg, time.Now, waitForLoginIPThrottleDelay).middleware
}

func newLoginIPThrottle(cfg LoginIPThrottleConfig, now func() time.Time, wait func(context.Context, time.Duration) bool) *loginIPThrottle {
	cfg = normalizeLoginIPThrottleConfig(cfg)
	return &loginIPThrottle{
		cfg:     cfg,
		entries: map[string]loginIPThrottleEntry{},
		now:     now,
		wait:    wait,
	}
}

func waitForLoginIPThrottleDelay(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func normalizeLoginIPThrottleConfig(cfg LoginIPThrottleConfig) LoginIPThrottleConfig {
	if cfg.Window <= 0 {
		cfg.Window = 10 * time.Minute
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 8
	}
	if cfg.SlowAfter < 0 {
		cfg.SlowAfter = 0
	}
	if cfg.SlowAfter == 0 {
		cfg.SlowAfter = 3
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 500 * time.Millisecond
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 4 * time.Second
	}
	if cfg.BlockDuration <= 0 {
		cfg.BlockDuration = 15 * time.Minute
	}
	if cfg.StateTTL <= 0 {
		cfg.StateTTL = cfg.Window + cfg.BlockDuration
	}
	if cfg.SlowAfter > cfg.MaxAttempts {
		cfg.SlowAfter = cfg.MaxAttempts
	}
	return cfg
}

func (t *loginIPThrottle) middleware(next http.Handler) http.Handler {
	if !t.cfg.Enabled {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delay, retryAfter, limited := t.check(r)
		if limited {
			seconds := int(math.Ceil(retryAfter.Seconds()))
			if seconds < 1 {
				seconds = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(seconds))
			httpx.Error(w, http.StatusTooManyRequests, "login_rate_limited", "Too many sign-in attempts from this network. Wait a moment and try again.")
			return
		}
		if delay > 0 {
			if ok := t.wait(r.Context(), delay); !ok {
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (t *loginIPThrottle) check(r *http.Request) (delay time.Duration, retryAfter time.Duration, limited bool) {
	now := t.now()
	key := loginIPKey(r)

	t.mu.Lock()
	defer t.mu.Unlock()

	t.pruneLocked(now)

	entry := t.entries[key]
	if entry.windowStart.IsZero() || now.Sub(entry.windowStart) >= t.cfg.Window {
		entry = loginIPThrottleEntry{windowStart: now}
	}

	entry.lastSeen = now
	if entry.blockedUntil.After(now) {
		t.entries[key] = entry
		return 0, entry.blockedUntil.Sub(now), true
	}

	entry.attempts++
	if entry.attempts > t.cfg.MaxAttempts {
		entry.blockedUntil = now.Add(t.cfg.BlockDuration)
		t.entries[key] = entry
		return 0, t.cfg.BlockDuration, true
	}

	delay = t.delayForAttempt(entry.attempts)
	t.entries[key] = entry
	return delay, 0, false
}

func (t *loginIPThrottle) delayForAttempt(attempts int) time.Duration {
	if attempts <= t.cfg.SlowAfter {
		return 0
	}
	delay := time.Duration(attempts-t.cfg.SlowAfter) * t.cfg.BaseDelay
	if delay > t.cfg.MaxDelay {
		return t.cfg.MaxDelay
	}
	return delay
}

func (t *loginIPThrottle) pruneLocked(now time.Time) {
	for key, entry := range t.entries {
		if now.Sub(entry.lastSeen) > t.cfg.StateTTL && !entry.blockedUntil.After(now) {
			delete(t.entries, key)
		}
	}
}

func loginIPKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}
