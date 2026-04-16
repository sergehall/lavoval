package mailer

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sergehall/lavoval/apps/api/internal/config"
	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

type MailRetentionWorker struct {
	cfg         config.Config
	jobs        repository.MailJobStore
	events      repository.MailEventStore
	runs        repository.MailCleanupRunStore
	metrics     *PrometheusHandler
	rootCtx     context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	interval    time.Duration
	batchSize   int
	jobsTTL     time.Duration
	eventsTTL   time.Duration
	dryRun      bool
	shouldStart bool
}

func NewMailRetentionWorker(cfg config.Config, jobs repository.MailJobStore, events repository.MailEventStore, runs repository.MailCleanupRunStore, metrics *PrometheusHandler) *MailRetentionWorker {
	worker := &MailRetentionWorker{
		cfg:         cfg,
		jobs:        jobs,
		events:      events,
		runs:        runs,
		metrics:     metrics,
		interval:    cfg.MailCleanupInterval,
		batchSize:   cfg.MailCleanupBatchSize,
		jobsTTL:     cfg.MailJobsRetention,
		eventsTTL:   cfg.MailEventsRetention,
		dryRun:      cfg.MailCleanupDryRun,
		shouldStart: cfg.MailCleanupInterval > 0 && (cfg.MailJobsRetention > 0 || cfg.MailEventsRetention > 0),
	}

	if !worker.shouldStart {
		return worker
	}

	worker.rootCtx, worker.cancel = context.WithCancel(context.Background())
	worker.wg.Add(1)
	go worker.loop()

	return worker
}

func (w *MailRetentionWorker) Close(ctx context.Context) error {
	if !w.shouldStart || w.cancel == nil {
		return nil
	}

	w.cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		w.wg.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *MailRetentionWorker) loop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.rootCtx.Done():
			return
		case <-ticker.C:
			w.runOnce()
		}
	}
}

func (w *MailRetentionWorker) runOnce() {
	start := time.Now()
	mode := "apply"
	if w.dryRun {
		mode = "dry_run"
	}

	var candidateJobs int64
	var candidateEvents int64
	var deletedJobs int64
	var deletedEvents int64

	now := time.Now().UTC()

	if w.jobsTTL > 0 {
		count, err := w.jobs.CountTerminalBefore(w.rootCtx, now.Add(-w.jobsTTL))
		if err != nil {
			w.recordFailure(mode, start, candidateJobs, candidateEvents, err)
			log.Printf("mailer: component=retention_cleanup status=count_jobs_failed mode=%s err=%v", mode, err)
			return
		}
		candidateJobs = count
	}

	if w.eventsTTL > 0 {
		count, err := w.events.CountBefore(w.rootCtx, now.Add(-w.eventsTTL))
		if err != nil {
			w.recordFailure(mode, start, candidateJobs, candidateEvents, err)
			log.Printf("mailer: component=retention_cleanup status=count_events_failed mode=%s err=%v", mode, err)
			return
		}
		candidateEvents = count
	}

	if !w.dryRun {
		if w.jobsTTL > 0 {
			count, err := w.jobs.DeleteTerminalBefore(w.rootCtx, now.Add(-w.jobsTTL), normalizedCleanupBatchSize(w.batchSize))
			if err != nil {
				w.recordFailure(mode, start, candidateJobs, candidateEvents, err)
				log.Printf("mailer: component=retention_cleanup status=delete_jobs_failed mode=%s err=%v", mode, err)
				return
			}
			deletedJobs = count
		}

		if w.eventsTTL > 0 {
			count, err := w.events.DeleteBefore(w.rootCtx, now.Add(-w.eventsTTL), normalizedCleanupBatchSize(w.batchSize))
			if err != nil {
				w.recordFailure(mode, start, candidateJobs, candidateEvents, err)
				log.Printf("mailer: component=retention_cleanup status=delete_events_failed mode=%s err=%v", mode, err)
				return
			}
			deletedEvents = count
		}
	}

	if w.metrics != nil {
		w.metrics.RecordCleanup(mode, "success", time.Since(start), deletedJobs, deletedEvents, candidateJobs, candidateEvents)
	}
	w.recordRun(domain.MailCleanupRun{
		ID:              uuid.NewString(),
		Mode:            mode,
		Status:          "success",
		DryRun:          w.dryRun,
		CandidateJobs:   candidateJobs,
		CandidateEvents: candidateEvents,
		DeletedJobs:     deletedJobs,
		DeletedEvents:   deletedEvents,
		DurationMs:      time.Since(start).Milliseconds(),
	})

	log.Printf(
		"mailer: component=retention_cleanup status=success mode=%s candidate_jobs=%d candidate_events=%d deleted_jobs=%d deleted_events=%d duration_ms=%d",
		mode,
		candidateJobs,
		candidateEvents,
		deletedJobs,
		deletedEvents,
		time.Since(start).Milliseconds(),
	)
}

func (w *MailRetentionWorker) recordFailure(mode string, start time.Time, candidateJobs int64, candidateEvents int64, err error) {
	if w.metrics != nil {
		w.metrics.RecordCleanup(mode, "failed", time.Since(start), 0, 0, candidateJobs, candidateEvents)
	}
	message := err.Error()
	w.recordRun(domain.MailCleanupRun{
		ID:              uuid.NewString(),
		Mode:            mode,
		Status:          "failed",
		DryRun:          w.dryRun,
		CandidateJobs:   candidateJobs,
		CandidateEvents: candidateEvents,
		DeletedJobs:     0,
		DeletedEvents:   0,
		ErrorMessage:    &message,
		DurationMs:      time.Since(start).Milliseconds(),
	})
}

func normalizedCleanupBatchSize(size int) int {
	if size <= 0 {
		return 500
	}
	return size
}

func (w *MailRetentionWorker) recordRun(run domain.MailCleanupRun) {
	if w.runs == nil {
		return
	}
	if _, err := w.runs.Create(context.Background(), run); err != nil {
		log.Printf("mailer: component=retention_cleanup status=record_run_failed err=%v", err)
	}
}
