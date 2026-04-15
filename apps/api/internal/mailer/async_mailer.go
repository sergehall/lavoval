package mailer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/textproto"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sergehall/lavoval/apps/api/internal/config"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

const (
	mailJobTypeVerification   = "verification"
	mailJobTypePasswordReset  = "password_reset"
	mailJobTypePasswordChange = "password_changed"
)

type PrometheusHandler struct {
	mu         sync.RWMutex
	counters   map[string]map[string]int64
	jobCounter repository.MailJobStore
}

func NewPrometheusHandler(jobCounter repository.MailJobStore) *PrometheusHandler {
	return &PrometheusHandler{
		counters: map[string]map[string]int64{
			"sent":        {},
			"failed":      {},
			"retry":       {},
			"dead_letter": {},
		},
		jobCounter: jobCounter,
	}
}

func (h *PrometheusHandler) Inc(result string, messageType string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.counters[result]; !ok {
		h.counters[result] = map[string]int64{}
	}
	h.counters[result][messageType]++
}

func (h *PrometheusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	writeCounterFamily := func(metric string, result string) {
		fmt.Fprintf(w, "# TYPE %s counter\n", metric)
		h.mu.RLock()
		defer h.mu.RUnlock()

		keys := make([]string, 0, len(h.counters[result]))
		for k := range h.counters[result] {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, messageType := range keys {
			fmt.Fprintf(w, "%s{message_type=%q} %d\n", metric, messageType, h.counters[result][messageType])
		}
	}

	writeCounterFamily("lavoval_mail_sent_total", "sent")
	writeCounterFamily("lavoval_mail_failed_total", "failed")
	writeCounterFamily("lavoval_mail_retry_total", "retry")
	writeCounterFamily("lavoval_mail_dead_letter_total", "dead_letter")

	if h.jobCounter == nil {
		return
	}

	counts, err := h.jobCounter.CountByStatus(r.Context())
	if err != nil {
		fmt.Fprintf(w, "# lavoval_mail_jobs_count unavailable: %v\n", err)
		return
	}

	fmt.Fprintln(w, "# TYPE lavoval_mail_jobs gauge")
	statuses := []domain.MailJobStatus{
		domain.MailJobStatusQueued,
		domain.MailJobStatusRetrying,
		domain.MailJobStatusProcessing,
		domain.MailJobStatusSent,
		domain.MailJobStatusDeadLetter,
	}
	for _, status := range statuses {
		fmt.Fprintf(w, "lavoval_mail_jobs{status=%q} %d\n", string(status), counts[status])
	}
}

type PostgresVerificationMailer struct {
	cfg         config.Config
	jobs        repository.MailJobStore
	maxAttempts int
}

func NewPostgresVerificationMailer(cfg config.Config, jobs repository.MailJobStore) *PostgresVerificationMailer {
	maxAttempts := cfg.MailMaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 4
	}

	return &PostgresVerificationMailer{
		cfg:         cfg,
		jobs:        jobs,
		maxAttempts: maxAttempts,
	}
}

func (m *PostgresVerificationMailer) SendVerificationEmail(ctx context.Context, email VerificationEmail) error {
	return m.enqueue(ctx, mailJobTypeVerification, email.ToEmail, email)
}

func (m *PostgresVerificationMailer) SendPasswordResetEmail(ctx context.Context, email PasswordResetEmail) error {
	return m.enqueue(ctx, mailJobTypePasswordReset, email.ToEmail, email)
}

func (m *PostgresVerificationMailer) SendPasswordChangedEmail(ctx context.Context, email PasswordChangedEmail) error {
	return m.enqueue(ctx, mailJobTypePasswordChange, email.ToEmail, email)
}

func (m *PostgresVerificationMailer) enqueue(ctx context.Context, messageType string, recipient string, payload any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal mail job payload: %w", err)
	}

	_, err = m.jobs.Enqueue(ctx, domain.MailJob{
		ID:             uuid.NewString(),
		MessageType:    messageType,
		RecipientEmail: recipient,
		Payload:        encoded,
		Status:         domain.MailJobStatusQueued,
		Attempts:       0,
		MaxAttempts:    m.maxAttempts,
		NextAttemptAt:  time.Now(),
	})
	if err != nil {
		return fmt.Errorf("enqueue mail job: %w", err)
	}

	return nil
}

type MailDispatcher struct {
	cfg          config.Config
	jobs         repository.MailJobStore
	transport    *SMTPVerificationMailer
	metrics      *PrometheusHandler
	rootCtx      context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	workerCount  int
	pollInterval time.Duration
	leaseTTL     time.Duration
}

func NewMailDispatcher(cfg config.Config, jobs repository.MailJobStore, metrics *PrometheusHandler) *MailDispatcher {
	workerCount := cfg.MailWorkerCount
	if workerCount <= 0 {
		workerCount = 4
	}

	pollInterval := cfg.MailPollInterval
	if pollInterval <= 0 {
		pollInterval = 500 * time.Millisecond
	}

	leaseTTL := cfg.MailLeaseTTL
	if leaseTTL <= 0 {
		leaseTTL = 30 * time.Second
	}

	rootCtx, cancel := context.WithCancel(context.Background())
	dispatcher := &MailDispatcher{
		cfg:          cfg,
		jobs:         jobs,
		transport:    NewSMTPVerificationMailer(cfg),
		metrics:      metrics,
		rootCtx:      rootCtx,
		cancel:       cancel,
		workerCount:  workerCount,
		pollInterval: pollInterval,
		leaseTTL:     leaseTTL,
	}

	for workerID := 1; workerID <= workerCount; workerID++ {
		dispatcher.wg.Add(1)
		go dispatcher.workerLoop(workerID)
	}

	return dispatcher
}

func (d *MailDispatcher) Close(ctx context.Context) error {
	d.cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		d.wg.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *MailDispatcher) workerLoop(workerID int) {
	defer d.wg.Done()

	for {
		select {
		case <-d.rootCtx.Done():
			return
		default:
		}

		job, found, err := d.jobs.ClaimNext(d.rootCtx, d.leaseTTL)
		if err != nil {
			log.Printf("mailer: worker_id=%d status=claim_failed err=%v", workerID, err)
			if !d.waitForNextPoll() {
				return
			}
			continue
		}
		if !found {
			if !d.waitForNextPoll() {
				return
			}
			continue
		}

		d.processJob(workerID, job)
	}
}

func (d *MailDispatcher) processJob(workerID int, job domain.MailJob) {
	start := time.Now()
	err := d.dispatch(job)
	durationMs := time.Since(start).Milliseconds()
	recipientDomain := recipientDomain(job.RecipientEmail)

	if err == nil {
		if err := d.jobs.MarkSent(d.rootCtx, job.ID, "smtp"); err != nil {
			log.Printf("mailer: worker_id=%d status=mark_sent_failed job_id=%s err=%v", workerID, job.ID, err)
			return
		}
		if d.metrics != nil {
			d.metrics.Inc("sent", job.MessageType)
		}
		log.Printf("mailer: status=sent provider=smtp message_type=%s recipient_domain=%s attempt=%d duration_ms=%d worker_id=%d", job.MessageType, recipientDomain, job.Attempts, durationMs, workerID)
		return
	}

	retryable, errorCode := classifyDeliveryError(err)
	if d.metrics != nil {
		d.metrics.Inc("failed", job.MessageType)
	}
	log.Printf("mailer: status=failed provider=smtp message_type=%s recipient_domain=%s attempt=%d duration_ms=%d worker_id=%d retryable=%t error_code=%s err=%v", job.MessageType, recipientDomain, job.Attempts, durationMs, workerID, retryable, errorCode, err)

	if retryable && job.Attempts < job.MaxAttempts {
		nextAttemptAt := time.Now().Add(retryBackoff(d.cfg.MailRetryBaseDelay, job.Attempts))
		if err := d.jobs.MarkRetry(d.rootCtx, job.ID, err.Error(), errorCode, nextAttemptAt); err != nil {
			log.Printf("mailer: worker_id=%d status=mark_retry_failed job_id=%s err=%v", workerID, job.ID, err)
			return
		}
		if d.metrics != nil {
			d.metrics.Inc("retry", job.MessageType)
		}
		return
	}

	if err := d.jobs.MarkDeadLetter(d.rootCtx, job.ID, err.Error(), errorCode); err != nil {
		log.Printf("mailer: worker_id=%d status=mark_dead_letter_failed job_id=%s err=%v", workerID, job.ID, err)
		return
	}
	if d.metrics != nil {
		d.metrics.Inc("dead_letter", job.MessageType)
	}
}

func (d *MailDispatcher) dispatch(job domain.MailJob) error {
	switch job.MessageType {
	case mailJobTypeVerification:
		var payload VerificationEmail
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("decode verification payload: %w", err)
		}
		return d.transport.SendVerificationEmail(d.rootCtx, payload)
	case mailJobTypePasswordReset:
		var payload PasswordResetEmail
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("decode password reset payload: %w", err)
		}
		return d.transport.SendPasswordResetEmail(d.rootCtx, payload)
	case mailJobTypePasswordChange:
		var payload PasswordChangedEmail
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("decode password changed payload: %w", err)
		}
		return d.transport.SendPasswordChangedEmail(d.rootCtx, payload)
	default:
		return fmt.Errorf("unknown mail job type: %s", job.MessageType)
	}
}

func (d *MailDispatcher) waitForNextPoll() bool {
	timer := time.NewTimer(d.pollInterval)
	defer timer.Stop()

	select {
	case <-d.rootCtx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func classifyDeliveryError(err error) (retryable bool, code string) {
	switch {
	case err == nil:
		return false, ""
	case errors.Is(err, context.Canceled):
		return true, "context_canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return true, "deadline_exceeded"
	}

	var smtpErr *textproto.Error
	if errors.As(err, &smtpErr) {
		switch {
		case smtpErr.Code >= 400 && smtpErr.Code < 500:
			return true, fmt.Sprintf("smtp_%d", smtpErr.Code)
		case smtpErr.Code >= 500:
			return false, fmt.Sprintf("smtp_%d", smtpErr.Code)
		}
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "render "):
		return false, "render_failed"
	case strings.Contains(msg, "invalid address"):
		return false, "address_invalid"
	case strings.Contains(msg, "timeout"):
		return true, "timeout"
	case strings.Contains(msg, "not configured"):
		return false, "not_configured"
	case strings.Contains(msg, "dial smtp"):
		return true, "smtp_dial_failed"
	case strings.Contains(msg, "starttls"):
		return true, "smtp_starttls"
	case strings.Contains(msg, "smtp auth"):
		return false, "smtp_auth"
	}

	return false, "delivery_failed"
}

func retryBackoff(base time.Duration, attempt int) time.Duration {
	if base <= 0 {
		base = time.Second
	}
	if attempt <= 1 {
		return base
	}

	backoff := base
	for i := 1; i < attempt; i++ {
		backoff *= 2
		if backoff > 30*time.Second {
			return 30 * time.Second
		}
	}
	return backoff
}

func recipientDomain(email string) string {
	parts := strings.Split(strings.TrimSpace(email), "@")
	if len(parts) != 2 || parts[1] == "" {
		return "unknown"
	}
	return strings.ToLower(parts[1])
}
