package mailer

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
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
	mu                  sync.RWMutex
	counters            map[string]map[counterKey]int64
	jobCounter          repository.MailJobStore
	cleanupRuns         map[cleanupRunKey]int64
	cleanupDeletedTotal map[string]int64
	cleanupCandidates   map[string]int64
	lastCleanupAt       time.Time
	lastCleanupDuration float64
}

type counterKey struct {
	Provider    string
	MessageType string
	ErrorCode   string
}

type cleanupRunKey struct {
	Mode   string
	Status string
}

func NewPrometheusHandler(jobCounter repository.MailJobStore) *PrometheusHandler {
	return &PrometheusHandler{
		counters: map[string]map[counterKey]int64{
			"sent":         {},
			"failed":       {},
			"retry":        {},
			"dead_letter":  {},
			"deduplicated": {},
		},
		jobCounter:          jobCounter,
		cleanupRuns:         map[cleanupRunKey]int64{},
		cleanupDeletedTotal: map[string]int64{"jobs": 0, "events": 0},
		cleanupCandidates:   map[string]int64{"jobs": 0, "events": 0},
	}
}

func (h *PrometheusHandler) Inc(result string, provider string, messageType string, errorCode string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.counters[result]; !ok {
		h.counters[result] = map[counterKey]int64{}
	}
	h.counters[result][counterKey{
		Provider:    provider,
		MessageType: messageType,
		ErrorCode:   errorCode,
	}]++
}

func (h *PrometheusHandler) RecordCleanup(mode string, status string, duration time.Duration, deletedJobs int64, deletedEvents int64, candidateJobs int64, candidateEvents int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.cleanupRuns[cleanupRunKey{Mode: mode, Status: status}]++
	h.cleanupCandidates["jobs"] = candidateJobs
	h.cleanupCandidates["events"] = candidateEvents
	if status == "success" && mode == "apply" {
		h.cleanupDeletedTotal["jobs"] += deletedJobs
		h.cleanupDeletedTotal["events"] += deletedEvents
	}
	h.lastCleanupAt = time.Now().UTC()
	h.lastCleanupDuration = duration.Seconds()
}

func (h *PrometheusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	writeCounterFamily := func(metric string, result string) {
		fmt.Fprintf(w, "# TYPE %s counter\n", metric)
		h.mu.RLock()
		defer h.mu.RUnlock()

		keys := make([]counterKey, 0, len(h.counters[result]))
		for k := range h.counters[result] {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			if keys[i].Provider != keys[j].Provider {
				return keys[i].Provider < keys[j].Provider
			}
			if keys[i].MessageType != keys[j].MessageType {
				return keys[i].MessageType < keys[j].MessageType
			}
			return keys[i].ErrorCode < keys[j].ErrorCode
		})
		for _, key := range keys {
			fmt.Fprintf(
				w,
				"%s{provider=%q,message_type=%q,error_code=%q} %d\n",
				metric,
				key.Provider,
				key.MessageType,
				key.ErrorCode,
				h.counters[result][key],
			)
		}
	}

	writeCounterFamily("lavoval_mail_sent_total", "sent")
	writeCounterFamily("lavoval_mail_failed_total", "failed")
	writeCounterFamily("lavoval_mail_retry_total", "retry")
	writeCounterFamily("lavoval_mail_dead_letter_total", "dead_letter")
	writeCounterFamily("lavoval_mail_deduplicated_total", "deduplicated")

	if h.jobCounter == nil {
		return
	}

	snapshot, err := h.jobCounter.OperationalSnapshot(r.Context())
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
		fmt.Fprintf(w, "lavoval_mail_jobs{status=%q} %d\n", string(status), snapshot.CountsByStatus[status])
	}

	fmt.Fprintln(w, "# TYPE lavoval_mail_oldest_ready_age_seconds gauge")
	fmt.Fprintf(w, "lavoval_mail_oldest_ready_age_seconds %f\n", snapshot.OldestReadyAgeSeconds)

	fmt.Fprintln(w, "# TYPE lavoval_mail_dead_letters gauge")
	keys := make([]string, 0, len(snapshot.DeadLettersByErrorCode))
	for errorCode := range snapshot.DeadLettersByErrorCode {
		keys = append(keys, errorCode)
	}
	sort.Strings(keys)
	for _, errorCode := range keys {
		fmt.Fprintf(w, "lavoval_mail_dead_letters{error_code=%q} %d\n", errorCode, snapshot.DeadLettersByErrorCode[errorCode])
	}

	fmt.Fprintln(w, "# TYPE lavoval_mail_cleanup_runs_total counter")
	h.mu.RLock()
	cleanupRunKeys := make([]cleanupRunKey, 0, len(h.cleanupRuns))
	for key := range h.cleanupRuns {
		cleanupRunKeys = append(cleanupRunKeys, key)
	}
	sort.Slice(cleanupRunKeys, func(i, j int) bool {
		if cleanupRunKeys[i].Mode != cleanupRunKeys[j].Mode {
			return cleanupRunKeys[i].Mode < cleanupRunKeys[j].Mode
		}
		return cleanupRunKeys[i].Status < cleanupRunKeys[j].Status
	})
	for _, key := range cleanupRunKeys {
		fmt.Fprintf(w, "lavoval_mail_cleanup_runs_total{mode=%q,status=%q} %d\n", key.Mode, key.Status, h.cleanupRuns[key])
	}

	fmt.Fprintln(w, "# TYPE lavoval_mail_cleanup_deleted_total counter")
	for _, resource := range []string{"jobs", "events"} {
		fmt.Fprintf(w, "lavoval_mail_cleanup_deleted_total{resource=%q} %d\n", resource, h.cleanupDeletedTotal[resource])
	}

	fmt.Fprintln(w, "# TYPE lavoval_mail_cleanup_candidates gauge")
	for _, resource := range []string{"jobs", "events"} {
		fmt.Fprintf(w, "lavoval_mail_cleanup_candidates{resource=%q} %d\n", resource, h.cleanupCandidates[resource])
	}

	fmt.Fprintln(w, "# TYPE lavoval_mail_cleanup_last_run_timestamp_seconds gauge")
	if !h.lastCleanupAt.IsZero() {
		fmt.Fprintf(w, "lavoval_mail_cleanup_last_run_timestamp_seconds %d\n", h.lastCleanupAt.Unix())
	} else {
		fmt.Fprintln(w, "lavoval_mail_cleanup_last_run_timestamp_seconds 0")
	}

	fmt.Fprintln(w, "# TYPE lavoval_mail_cleanup_last_duration_seconds gauge")
	fmt.Fprintf(w, "lavoval_mail_cleanup_last_duration_seconds %f\n", h.lastCleanupDuration)
	h.mu.RUnlock()
}

type PostgresVerificationMailer struct {
	cfg          config.Config
	jobs         repository.MailJobStore
	events       repository.MailEventStore
	suppressions repository.MailSuppressionStore
	metrics      *PrometheusHandler
	maxAttempts  int
}

func NewPostgresVerificationMailer(cfg config.Config, jobs repository.MailJobStore, events repository.MailEventStore, suppressions repository.MailSuppressionStore, metrics *PrometheusHandler) *PostgresVerificationMailer {
	maxAttempts := cfg.MailMaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 4
	}

	return &PostgresVerificationMailer{
		cfg:          cfg,
		jobs:         jobs,
		events:       events,
		suppressions: suppressions,
		metrics:      metrics,
		maxAttempts:  maxAttempts,
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
	if m.suppressions != nil {
		match, err := m.suppressions.FindMatch(ctx, recipient)
		if err != nil {
			return fmt.Errorf("check mail suppression: %w", err)
		}
		if match != nil {
			if m.metrics != nil {
				m.metrics.Inc("suppressed", "queue", messageType, string(match.Kind))
			}
			log.Printf(
				"mailer: status=suppressed message_type=%s recipient_domain=%s suppression_kind=%s suppression_value=%s reason=%s",
				messageType,
				recipientDomain(recipient),
				match.Kind,
				match.Value,
				match.Reason,
			)
			return nil
		}
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal mail job payload: %w", err)
	}

	idempotencyKey := mailIdempotencyKey(messageType, recipient, encoded)

	job, deduplicated, err := m.jobs.Enqueue(ctx, domain.MailJob{
		ID:             uuid.NewString(),
		MessageType:    messageType,
		RecipientEmail: recipient,
		IdempotencyKey: &idempotencyKey,
		Payload:        encoded,
		Status:         domain.MailJobStatusQueued,
		Attempts:       0,
		MaxAttempts:    m.maxAttempts,
		NextAttemptAt:  time.Now(),
	})
	if err != nil {
		return fmt.Errorf("enqueue mail job: %w", err)
	}
	if deduplicated {
		if m.metrics != nil {
			m.metrics.Inc("deduplicated", "queue", messageType, "")
		}
		m.recordEvent(ctx, queuedJobEvent(job, "deduplicated", "queue", "", nil, map[string]any{
			"idempotencyKey": idempotencyKey,
		}))
		log.Printf("mailer: status=deduplicated message_type=%s recipient_domain=%s idempotency_key=%s", messageType, recipientDomain(recipient), idempotencyKey)
		return nil
	}
	m.recordEvent(ctx, queuedJobEvent(job, "queued", "queue", "", nil, map[string]any{
		"idempotencyKey": idempotencyKey,
	}))

	return nil
}

type MailDispatcher struct {
	cfg          config.Config
	jobs         repository.MailJobStore
	events       repository.MailEventStore
	provider     Provider
	metrics      *PrometheusHandler
	rootCtx      context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	workerCount  int
	pollInterval time.Duration
	leaseTTL     time.Duration
	rateLimiter  *sendRateLimiter
}

func NewMailDispatcher(cfg config.Config, jobs repository.MailJobStore, events repository.MailEventStore, metrics *PrometheusHandler) *MailDispatcher {
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
		events:       events,
		provider:     newConfiguredProvider(cfg),
		metrics:      metrics,
		rootCtx:      rootCtx,
		cancel:       cancel,
		workerCount:  workerCount,
		pollInterval: pollInterval,
		leaseTTL:     leaseTTL,
		rateLimiter:  newSendRateLimiter(rootCtx, cfg.MailRateLimitPerSecond, cfg.MailRateLimitBurst),
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
	sendCtx, cancel := context.WithTimeout(context.Background(), jobSendTimeout(d.cfg))
	defer cancel()

	result, err := d.dispatch(sendCtx, job)
	durationMs := time.Since(start).Milliseconds()
	recipientDomain := recipientDomain(job.RecipientEmail)
	provider := "unknown"
	if result.Provider != "" {
		provider = result.Provider
	}
	providerMessageID := result.ProviderMessageID

	if err == nil {
		stateCtx, stateCancel := newMailJobStateContext()
		defer stateCancel()

		if err := d.jobs.MarkSent(stateCtx, job.ID, provider, providerMessageID); err != nil {
			log.Printf("mailer: worker_id=%d status=mark_sent_failed job_id=%s err=%v", workerID, job.ID, err)
			return
		}
		if d.metrics != nil {
			d.metrics.Inc("sent", provider, job.MessageType, "")
		}
		d.recordEvent(queuedJobEvent(job, "sent", provider, "", &job.Attempts, map[string]any{
			"durationMs":        durationMs,
			"providerMessageId": providerMessageID,
		}))
		log.Printf("mailer: status=sent provider=%s provider_message_id=%s message_type=%s recipient_domain=%s attempt=%d duration_ms=%d worker_id=%d job_id=%s idempotency_key=%s", provider, providerMessageID, job.MessageType, recipientDomain, job.Attempts, durationMs, workerID, job.ID, derefString(job.IdempotencyKey))
		return
	}

	retryable, errorCode := classifyDeliveryError(err)
	if d.metrics != nil {
		d.metrics.Inc("failed", provider, job.MessageType, errorCode)
	}
	d.recordEvent(queuedJobEvent(job, "failed", provider, errorCode, &job.Attempts, map[string]any{
		"durationMs": durationMs,
		"error":      err.Error(),
		"retryable":  retryable,
	}))
	log.Printf("mailer: status=failed provider=%s message_type=%s recipient_domain=%s attempt=%d duration_ms=%d worker_id=%d retryable=%t error_code=%s job_id=%s idempotency_key=%s err=%v", provider, job.MessageType, recipientDomain, job.Attempts, durationMs, workerID, retryable, errorCode, job.ID, derefString(job.IdempotencyKey), err)

	if retryable && job.Attempts < job.MaxAttempts {
		nextAttemptAt := time.Now().Add(retryBackoff(d.cfg.MailRetryBaseDelay, job.Attempts))
		stateCtx, stateCancel := newMailJobStateContext()
		defer stateCancel()

		if err := d.jobs.MarkRetry(stateCtx, job.ID, err.Error(), errorCode, nextAttemptAt); err != nil {
			log.Printf("mailer: worker_id=%d status=mark_retry_failed job_id=%s err=%v", workerID, job.ID, err)
			return
		}
		if d.metrics != nil {
			d.metrics.Inc("retry", provider, job.MessageType, errorCode)
		}
		d.recordEvent(queuedJobEvent(job, "retry_scheduled", provider, errorCode, &job.Attempts, map[string]any{
			"nextAttemptAt": nextAttemptAt.UTC().Format(time.RFC3339Nano),
		}))
		return
	}

	stateCtx, stateCancel := newMailJobStateContext()
	defer stateCancel()

	if err := d.jobs.MarkDeadLetter(stateCtx, job.ID, err.Error(), errorCode); err != nil {
		log.Printf("mailer: worker_id=%d status=mark_dead_letter_failed job_id=%s err=%v", workerID, job.ID, err)
		return
	}
	if d.metrics != nil {
		d.metrics.Inc("dead_letter", provider, job.MessageType, errorCode)
	}
	d.recordEvent(queuedJobEvent(job, "dead_letter", provider, errorCode, &job.Attempts, nil))
}

func (d *MailDispatcher) dispatch(ctx context.Context, job domain.MailJob) (SendResult, error) {
	if d.rateLimiter != nil {
		if err := d.rateLimiter.Wait(ctx); err != nil {
			return SendResult{}, temporaryDeliveryError("rate_limiter_wait_failed", err)
		}
	}

	switch job.MessageType {
	case mailJobTypeVerification:
		var payload VerificationEmail
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return SendResult{}, permanentDeliveryError("payload_decode_failed", fmt.Errorf("decode verification payload: %w", err))
		}
		msg, err := buildVerificationMessage(job.ID, payload, d.cfg.AppURL)
		if err != nil {
			return SendResult{}, err
		}
		return d.provider.Send(ctx, msg)
	case mailJobTypePasswordReset:
		var payload PasswordResetEmail
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return SendResult{}, permanentDeliveryError("payload_decode_failed", fmt.Errorf("decode password reset payload: %w", err))
		}
		msg, err := buildPasswordResetMessage(job.ID, payload, d.cfg.AppURL)
		if err != nil {
			return SendResult{}, err
		}
		return d.provider.Send(ctx, msg)
	case mailJobTypePasswordChange:
		var payload PasswordChangedEmail
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return SendResult{}, permanentDeliveryError("payload_decode_failed", fmt.Errorf("decode password changed payload: %w", err))
		}
		msg, err := buildPasswordChangedMessage(job.ID, payload, d.cfg.AppURL)
		if err != nil {
			return SendResult{}, err
		}
		return d.provider.Send(ctx, msg)
	default:
		return SendResult{}, permanentDeliveryError("message_type_unknown", fmt.Errorf("unknown mail job type: %s", job.MessageType))
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

	var deliveryErr *DeliveryError
	if errors.As(err, &deliveryErr) {
		return deliveryErr.Retryable(), deliveryErr.Code
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

func jobSendTimeout(cfg config.Config) time.Duration {
	if cfg.MailSendTimeout > 0 {
		return cfg.MailSendTimeout
	}
	if cfg.SMTPDialTimeout > 0 {
		return cfg.SMTPDialTimeout + 5*time.Second
	}
	return 15 * time.Second
}

func newMailJobStateContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (m *PostgresVerificationMailer) recordEvent(ctx context.Context, event domain.MailEvent) {
	if m.events == nil {
		return
	}
	if _, err := m.events.Append(ctx, event); err != nil {
		log.Printf("mailer: status=record_event_failed event_type=%s job_id=%s err=%v", event.EventType, event.JobID, err)
	}
}

func (d *MailDispatcher) recordEvent(event domain.MailEvent) {
	if d.events == nil {
		return
	}
	ctx, cancel := newMailJobStateContext()
	defer cancel()
	if _, err := d.events.Append(ctx, event); err != nil {
		log.Printf("mailer: status=record_event_failed event_type=%s job_id=%s err=%v", event.EventType, event.JobID, err)
	}
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

func mailIdempotencyKey(messageType string, recipient string, payload []byte) string {
	sum := sha256.Sum256([]byte(messageType + "|" + strings.ToLower(strings.TrimSpace(recipient)) + "|" + string(payload)))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func queuedJobEvent(job domain.MailJob, eventType string, provider string, errorCode string, attempt *int, metadata map[string]any) domain.MailEvent {
	var metadataJSON []byte
	if metadata != nil {
		if encoded, err := json.Marshal(metadata); err == nil {
			metadataJSON = encoded
		}
	}

	var providerValue *string
	if provider != "" {
		providerValue = &provider
	}
	var errorCodeValue *string
	if errorCode != "" {
		errorCodeValue = &errorCode
	}

	return domain.MailEvent{
		ID:             uuid.NewString(),
		JobID:          job.ID,
		EventType:      eventType,
		MessageType:    job.MessageType,
		Provider:       providerValue,
		RecipientEmail: job.RecipientEmail,
		ErrorCode:      errorCodeValue,
		Attempt:        attempt,
		Metadata:       metadataJSON,
	}
}
