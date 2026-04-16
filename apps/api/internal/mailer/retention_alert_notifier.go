package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/config"
)

type cleanupAlertNotifier interface {
	Notify(context.Context, cleanupAlert) error
}

type cleanupAlert struct {
	Type            string
	Severity        string
	Message         string
	Mode            string
	DryRun          bool
	CandidateJobs   int64
	CandidateEvents int64
	DeletedJobs     int64
	DeletedEvents   int64
	ErrorMessage    string
}

type webhookCleanupAlertNotifier struct {
	cfg      config.Config
	client   *http.Client
	cooldown time.Duration
	mu       sync.Mutex
	lastSent map[string]time.Time
	metrics  *PrometheusHandler
}

type cleanupAlertPayload struct {
	App             string    `json:"app"`
	Environment     string    `json:"environment"`
	Type            string    `json:"type"`
	Severity        string    `json:"severity"`
	Message         string    `json:"message"`
	Mode            string    `json:"mode"`
	DryRun          bool      `json:"dryRun"`
	CandidateJobs   int64     `json:"candidateJobs"`
	CandidateEvents int64     `json:"candidateEvents"`
	DeletedJobs     int64     `json:"deletedJobs"`
	DeletedEvents   int64     `json:"deletedEvents"`
	ErrorMessage    string    `json:"errorMessage,omitempty"`
	OccurredAt      time.Time `json:"occurredAt"`
}

func newCleanupAlertNotifier(cfg config.Config, metrics *PrometheusHandler) cleanupAlertNotifier {
	if cfg.MailAlertWebhookURL == "" {
		return nil
	}

	timeout := cfg.MailAlertWebhookTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &webhookCleanupAlertNotifier{
		cfg:      cfg,
		client:   &http.Client{Timeout: timeout},
		cooldown: cfg.MailAlertWebhookCooldown,
		lastSent: make(map[string]time.Time),
		metrics:  metrics,
	}
}

func (n *webhookCleanupAlertNotifier) Notify(ctx context.Context, alert cleanupAlert) error {
	if alert.Type == "" {
		return fmt.Errorf("alert type is required")
	}
	if n.shouldSuppress(alert.Type) {
		return nil
	}

	payload, err := json.Marshal(cleanupAlertPayload{
		App:             n.cfg.AppName,
		Environment:     n.cfg.AppEnv,
		Type:            alert.Type,
		Severity:        alert.Severity,
		Message:         alert.Message,
		Mode:            alert.Mode,
		DryRun:          alert.DryRun,
		CandidateJobs:   alert.CandidateJobs,
		CandidateEvents: alert.CandidateEvents,
		DeletedJobs:     alert.DeletedJobs,
		DeletedEvents:   alert.DeletedEvents,
		ErrorMessage:    alert.ErrorMessage,
		OccurredAt:      time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("marshal cleanup alert payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.cfg.MailAlertWebhookURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build cleanup alert request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		n.recordMetric(alert.Type, "failed")
		return fmt.Errorf("send cleanup alert webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		n.recordMetric(alert.Type, "failed")
		return fmt.Errorf("cleanup alert webhook returned status %d", resp.StatusCode)
	}

	n.recordMetric(alert.Type, "sent")
	return nil
}

func (n *webhookCleanupAlertNotifier) shouldSuppress(alertType string) bool {
	if n.cooldown <= 0 {
		return false
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if last, ok := n.lastSent[alertType]; ok && time.Since(last) < n.cooldown {
		return true
	}
	n.lastSent[alertType] = time.Now()
	return false
}

func (n *webhookCleanupAlertNotifier) recordMetric(alertType string, status string) {
	if n.metrics == nil {
		return
	}
	n.metrics.RecordAlertNotification(alertType, status)
}

func buildBacklogCleanupAlert(cfg config.Config, candidateJobs int64, candidateEvents int64, mode string, dryRun bool) *cleanupAlert {
	if (cfg.MailCleanupAlertJobsThreshold <= 0 || candidateJobs < cfg.MailCleanupAlertJobsThreshold) &&
		(cfg.MailCleanupAlertEventsThreshold <= 0 || candidateEvents < cfg.MailCleanupAlertEventsThreshold) {
		return nil
	}

	return &cleanupAlert{
		Type:            "cleanup_backlog_threshold_exceeded",
		Severity:        "warning",
		Message:         fmt.Sprintf("Mail retention backlog exceeded threshold: jobs=%d events=%d", candidateJobs, candidateEvents),
		Mode:            mode,
		DryRun:          dryRun,
		CandidateJobs:   candidateJobs,
		CandidateEvents: candidateEvents,
	}
}

func buildFailureCleanupAlert(mode string, dryRun bool, candidateJobs int64, candidateEvents int64, err error) cleanupAlert {
	return cleanupAlert{
		Type:            "cleanup_failed",
		Severity:        "critical",
		Message:         "Mail retention cleanup failed",
		Mode:            mode,
		DryRun:          dryRun,
		CandidateJobs:   candidateJobs,
		CandidateEvents: candidateEvents,
		ErrorMessage:    err.Error(),
	}
}
