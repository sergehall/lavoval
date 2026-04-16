package mailer

import (
	"context"
	"time"
)

type sendRateLimiter struct {
	tokens chan struct{}
}

func newSendRateLimiter(ctx context.Context, perSecond int, burst int) *sendRateLimiter {
	if perSecond <= 0 {
		return nil
	}
	if burst <= 0 {
		burst = 1
	}

	limiter := &sendRateLimiter{tokens: make(chan struct{}, burst)}
	for i := 0; i < burst; i++ {
		limiter.tokens <- struct{}{}
	}

	interval := time.Second / time.Duration(perSecond)
	if interval <= 0 {
		interval = time.Second
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				select {
				case limiter.tokens <- struct{}{}:
				default:
				}
			}
		}
	}()

	return limiter
}

func (l *sendRateLimiter) Wait(ctx context.Context) error {
	if l == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-l.tokens:
		return nil
	}
}
