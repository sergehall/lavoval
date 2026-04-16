package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

type AdminService struct {
	users        repository.UserStore
	profiles     repository.ProfileStore
	skills       repository.SkillStore
	enrollments  repository.EnrollmentStore
	modules      repository.ModuleStore
	mailJobs     repository.MailJobStore
	mailEvents   repository.MailEventStore
	suppressions repository.MailSuppressionStore
	cleanupRuns  repository.MailCleanupRunStore
	retention    MailRetentionPolicy
}

type MailRetentionPolicy struct {
	JobsRetention          time.Duration
	EventsRetention        time.Duration
	CleanupBatchSize       int
	CleanupInterval        time.Duration
	CleanupDryRun          bool
	AlertJobsThreshold     int64
	AlertEventsThreshold   int64
	AlertFailureStreak     int
	AlertStaleAfter        time.Duration
	WebhookAlertingEnabled bool
}

func NewAdminService(users repository.UserStore, profiles repository.ProfileStore, skills repository.SkillStore, enrollments repository.EnrollmentStore, modules repository.ModuleStore, mailJobs repository.MailJobStore, mailEvents repository.MailEventStore, suppressions repository.MailSuppressionStore, cleanupRuns repository.MailCleanupRunStore, retention MailRetentionPolicy) *AdminService {
	return &AdminService{
		users:        users,
		profiles:     profiles,
		skills:       skills,
		enrollments:  enrollments,
		modules:      modules,
		mailJobs:     mailJobs,
		mailEvents:   mailEvents,
		suppressions: suppressions,
		cleanupRuns:  cleanupRuns,
		retention:    retention,
	}
}

type UserDetail struct {
	User    domain.User    `json:"user"`
	Profile domain.Profile `json:"profile"`
}

type UpdateUserInput struct {
	Role   domain.Role          `json:"role"   validate:"required,oneof=user admin"`
	Status domain.AccountStatus `json:"status" validate:"required,oneof=active invited suspended"`
}

type CreateMailSuppressionInput struct {
	Kind   domain.MailSuppressionKind `json:"kind" validate:"required,oneof=email domain"`
	Value  string                     `json:"value" validate:"required,min=3,max=255"`
	Reason string                     `json:"reason" validate:"required,min=3,max=255"`
}

func (s *AdminService) ListUsers(ctx context.Context) ([]domain.User, error) {
	users, err := s.users.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (s *AdminService) GetUser(ctx context.Context, id string) (UserDetail, error) {
	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		return UserDetail{}, fmt.Errorf("find user: %w", err)
	}
	profile, err := s.profiles.FindByUserID(ctx, id)
	if err != nil {
		return UserDetail{}, fmt.Errorf("find profile: %w", err)
	}
	return UserDetail{User: user, Profile: profile}, nil
}

func (s *AdminService) UpdateUser(ctx context.Context, id string, input UpdateUserInput) (domain.User, error) {
	user, err := s.users.UpdateRoleAndStatus(ctx, id, input.Role, input.Status)
	if err != nil {
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

type AssignSkillInput struct {
	UserID  string `json:"userId"  validate:"required,uuid4"`
	SkillID string `json:"skillId" validate:"required,uuid4"`
}

type UpdateEnrollmentInput struct {
	Status          domain.EnrollmentStatus `json:"status"          validate:"required,oneof=assigned in_progress completed"`
	ProgressPercent int                     `json:"progressPercent" validate:"min=0,max=100"`
}

func (s *AdminService) ListEnrollments(ctx context.Context) ([]domain.EnrollmentDetail, error) {
	items, err := s.enrollments.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enrollments: %w", err)
	}
	return items, nil
}

func (s *AdminService) AssignSkill(ctx context.Context, input AssignSkillInput) (domain.Enrollment, error) {
	enrollment, err := s.enrollments.Create(ctx, input.UserID, input.SkillID)
	if err != nil {
		return domain.Enrollment{}, fmt.Errorf("assign skill: %w", err)
	}
	return enrollment, nil
}

func (s *AdminService) UpdateEnrollment(ctx context.Context, id string, input UpdateEnrollmentInput) (domain.Enrollment, error) {
	enrollment, err := s.enrollments.UpdateStatus(ctx, id, input.Status, input.ProgressPercent)
	if err != nil {
		return domain.Enrollment{}, fmt.Errorf("update enrollment: %w", err)
	}
	return enrollment, nil
}

// ── Module methods ────────────────────────────────────────────────────────

type ModuleMutationInput struct {
	Slug     string             `json:"slug"     validate:"required,min=2,max=100"`
	Title    string             `json:"title"    validate:"required,min=2,max=200"`
	Summary  string             `json:"summary"  validate:"required,min=2,max=500"`
	Content  string             `json:"content"  validate:"required,min=10"`
	Position int                `json:"position" validate:"min=0"`
	Status   domain.SkillStatus `json:"status"   validate:"required,oneof=draft published archived"`
}

func (s *AdminService) ListModules(ctx context.Context, skillID string) ([]domain.Module, error) {
	items, err := s.modules.ListBySkillID(ctx, skillID)
	if err != nil {
		return nil, fmt.Errorf("list modules: %w", err)
	}
	return items, nil
}

func (s *AdminService) CreateModule(ctx context.Context, skillID string, input ModuleMutationInput) (domain.Module, error) {
	module := domain.Module{
		ID:       uuid.NewString(),
		SkillID:  skillID,
		Slug:     input.Slug,
		Title:    input.Title,
		Summary:  input.Summary,
		Content:  input.Content,
		Position: input.Position,
		Status:   input.Status,
	}
	created, err := s.modules.Create(ctx, module)
	if err != nil {
		return domain.Module{}, fmt.Errorf("create module: %w", err)
	}
	return created, nil
}

func (s *AdminService) UpdateModule(ctx context.Context, moduleID string, input ModuleMutationInput) (domain.Module, error) {
	module := domain.Module{
		ID:       moduleID,
		Slug:     input.Slug,
		Title:    input.Title,
		Summary:  input.Summary,
		Content:  input.Content,
		Position: input.Position,
		Status:   input.Status,
	}
	updated, err := s.modules.Update(ctx, module)
	if err != nil {
		return domain.Module{}, fmt.Errorf("update module: %w", err)
	}
	return updated, nil
}

func (s *AdminService) DeleteModule(ctx context.Context, moduleID string) error {
	if err := s.modules.SoftDelete(ctx, moduleID); err != nil {
		return fmt.Errorf("delete module: %w", err)
	}
	return nil
}

func (s *AdminService) ListSkills(ctx context.Context) ([]domain.Skill, error) {
	skills, err := s.skills.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list admin skills: %w", err)
	}
	return skills, nil
}

func (s *AdminService) MailOperations(ctx context.Context) (domain.MailOperationalSnapshot, error) {
	if s.mailJobs == nil {
		return domain.MailOperationalSnapshot{}, fmt.Errorf("mail jobs store is not configured")
	}

	snapshot, err := s.mailJobs.OperationalSnapshot(ctx)
	if err != nil {
		return domain.MailOperationalSnapshot{}, fmt.Errorf("mail operations snapshot: %w", err)
	}
	return snapshot, nil
}

func (s *AdminService) ListDeadLetters(ctx context.Context, filter domain.MailJobFilter) ([]domain.MailJob, error) {
	if s.mailJobs == nil {
		return nil, fmt.Errorf("mail jobs store is not configured")
	}

	items, err := s.mailJobs.ListDeadLetters(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list dead letters: %w", err)
	}
	return items, nil
}

func (s *AdminService) ListMailEvents(ctx context.Context, filter domain.MailEventFilter) ([]domain.MailEvent, error) {
	if s.mailEvents == nil {
		return nil, fmt.Errorf("mail events store is not configured")
	}

	events, err := s.mailEvents.ListRecent(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list mail events: %w", err)
	}
	return events, nil
}

func (s *AdminService) ListMailEventsByJob(ctx context.Context, jobID string, filter domain.MailEventFilter) ([]domain.MailEvent, error) {
	if s.mailEvents == nil {
		return nil, fmt.Errorf("mail events store is not configured")
	}

	events, err := s.mailEvents.ListByJobID(ctx, jobID, filter)
	if err != nil {
		return nil, fmt.Errorf("list mail events by job: %w", err)
	}
	return events, nil
}

func (s *AdminService) RequeueDeadLetter(ctx context.Context, jobID string) (domain.MailJob, error) {
	if s.mailJobs == nil {
		return domain.MailJob{}, fmt.Errorf("mail jobs store is not configured")
	}

	job, err := s.mailJobs.RequeueDeadLetter(ctx, jobID)
	if err != nil {
		return domain.MailJob{}, fmt.Errorf("requeue dead letter: %w", err)
	}
	if s.mailEvents != nil {
		attempt := job.Attempts
		if _, err := s.mailEvents.Append(ctx, domain.MailEvent{
			ID:             uuid.NewString(),
			JobID:          job.ID,
			EventType:      "requeued",
			MessageType:    job.MessageType,
			RecipientEmail: job.RecipientEmail,
			Attempt:        &attempt,
		}); err != nil {
			log.Printf("admin: append mail requeue event failed for job %s: %v", job.ID, err)
		}
	}
	return job, nil
}

func (s *AdminService) ReplayMailJob(ctx context.Context, jobID string) (domain.MailJob, error) {
	if s.mailJobs == nil {
		return domain.MailJob{}, fmt.Errorf("mail jobs store is not configured")
	}

	source, err := s.mailJobs.FindByID(ctx, jobID)
	if err != nil {
		return domain.MailJob{}, fmt.Errorf("find mail job for replay: %w", err)
	}

	replayKey := uuid.NewString()
	if source.IdempotencyKey != nil && *source.IdempotencyKey != "" {
		replayKey = *source.IdempotencyKey + ":replay:" + uuid.NewString()
	}

	replay := domain.MailJob{
		ID:             uuid.NewString(),
		MessageType:    source.MessageType,
		RecipientEmail: source.RecipientEmail,
		IdempotencyKey: &replayKey,
		Payload:        source.Payload,
		Status:         domain.MailJobStatusQueued,
		Attempts:       0,
		MaxAttempts:    source.MaxAttempts,
		NextAttemptAt:  time.Now().UTC(),
	}

	job, deduplicated, err := s.mailJobs.Replay(ctx, replay)
	if err != nil {
		return domain.MailJob{}, fmt.Errorf("replay mail job: %w", err)
	}
	if s.mailEvents != nil {
		metadata := []byte(fmt.Sprintf(`{"sourceJobId":"%s","deduplicated":%t}`, source.ID, deduplicated))
		if _, err := s.mailEvents.Append(ctx, domain.MailEvent{
			ID:             uuid.NewString(),
			JobID:          job.ID,
			EventType:      "replayed",
			MessageType:    job.MessageType,
			RecipientEmail: job.RecipientEmail,
			Metadata:       metadata,
		}); err != nil {
			log.Printf("admin: append mail replay event failed for job %s: %v", job.ID, err)
		}
	}

	return job, nil
}

func (s *AdminService) ListMailSuppressions(ctx context.Context) ([]domain.MailSuppression, error) {
	if s.suppressions == nil {
		return nil, fmt.Errorf("mail suppressions store is not configured")
	}
	items, err := s.suppressions.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list mail suppressions: %w", err)
	}
	return items, nil
}

func (s *AdminService) MailRetentionSnapshot(ctx context.Context) (domain.MailRetentionSnapshot, error) {
	if s.mailJobs == nil || s.mailEvents == nil {
		return domain.MailRetentionSnapshot{}, fmt.Errorf("mail retention stores are not configured")
	}

	snapshot := domain.MailRetentionSnapshot{
		JobsRetention:          s.retention.JobsRetention.String(),
		EventsRetention:        s.retention.EventsRetention.String(),
		CleanupBatchSize:       normalizedCleanupBatchSize(s.retention.CleanupBatchSize),
		CleanupInterval:        s.retention.CleanupInterval.String(),
		CleanupDryRun:          s.retention.CleanupDryRun,
		AutoCleanupEnabled:     s.retention.CleanupInterval > 0,
		JobsRetentionActive:    s.retention.JobsRetention > 0,
		EventsRetentionActive:  s.retention.EventsRetention > 0,
		Alerts:                 make([]domain.MailRetentionAlert, 0),
		WebhookAlertingEnabled: s.retention.WebhookAlertingEnabled,
	}

	if s.retention.JobsRetention > 0 {
		cutoff := time.Now().UTC().Add(-s.retention.JobsRetention)
		snapshot.JobsCutoff = &cutoff

		count, err := s.mailJobs.CountTerminalBefore(ctx, cutoff)
		if err != nil {
			return domain.MailRetentionSnapshot{}, fmt.Errorf("count purgeable mail jobs: %w", err)
		}
		snapshot.EligibleJobs = count
	}

	if s.retention.EventsRetention > 0 {
		cutoff := time.Now().UTC().Add(-s.retention.EventsRetention)
		snapshot.EventsCutoff = &cutoff

		count, err := s.mailEvents.CountBefore(ctx, cutoff)
		if err != nil {
			return domain.MailRetentionSnapshot{}, fmt.Errorf("count purgeable mail events: %w", err)
		}
		snapshot.EligibleEvents = count
	}

	if s.cleanupRuns != nil {
		runs, err := s.cleanupRuns.ListRecent(ctx, 10)
		if err != nil {
			return domain.MailRetentionSnapshot{}, fmt.Errorf("list recent cleanup runs: %w", err)
		}
		if len(runs) > 0 {
			snapshot.LatestCleanupRun = &runs[0]
		}
		snapshot.Alerts = buildMailRetentionAlerts(snapshot, runs, s.retention)
	}

	return snapshot, nil
}

func (s *AdminService) ListMailCleanupRuns(ctx context.Context, limit int) ([]domain.MailCleanupRun, error) {
	if s.cleanupRuns == nil {
		return nil, fmt.Errorf("mail cleanup run store is not configured")
	}
	items, err := s.cleanupRuns.ListRecent(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list cleanup runs: %w", err)
	}
	return items, nil
}

func (s *AdminService) CleanupMailRetention(ctx context.Context) (domain.MailCleanupResult, error) {
	if s.mailJobs == nil || s.mailEvents == nil {
		return domain.MailCleanupResult{}, fmt.Errorf("mail retention stores are not configured")
	}

	batchSize := normalizedCleanupBatchSize(s.retention.CleanupBatchSize)
	result := domain.MailCleanupResult{
		CompletedAt: time.Now().UTC(),
	}

	if s.retention.JobsRetention > 0 {
		cutoff := result.CompletedAt.Add(-s.retention.JobsRetention)
		deleted, err := s.mailJobs.DeleteTerminalBefore(ctx, cutoff, batchSize)
		if err != nil {
			return domain.MailCleanupResult{}, fmt.Errorf("cleanup mail jobs: %w", err)
		}
		result.JobsDeleted = deleted

		remaining, err := s.mailJobs.CountTerminalBefore(ctx, cutoff)
		if err != nil {
			return domain.MailCleanupResult{}, fmt.Errorf("count remaining purgeable mail jobs: %w", err)
		}
		result.RemainingJobs = remaining
	}

	if s.retention.EventsRetention > 0 {
		cutoff := result.CompletedAt.Add(-s.retention.EventsRetention)
		deleted, err := s.mailEvents.DeleteBefore(ctx, cutoff, batchSize)
		if err != nil {
			return domain.MailCleanupResult{}, fmt.Errorf("cleanup mail events: %w", err)
		}
		result.EventsDeleted = deleted

		remaining, err := s.mailEvents.CountBefore(ctx, cutoff)
		if err != nil {
			return domain.MailCleanupResult{}, fmt.Errorf("count remaining purgeable mail events: %w", err)
		}
		result.RemainingEvents = remaining
	}

	s.recordCleanupRun(ctx, domain.MailCleanupRun{
		ID:              uuid.NewString(),
		Mode:            "manual",
		Status:          "success",
		DryRun:          false,
		CandidateJobs:   result.JobsDeleted + result.RemainingJobs,
		CandidateEvents: result.EventsDeleted + result.RemainingEvents,
		DeletedJobs:     result.JobsDeleted,
		DeletedEvents:   result.EventsDeleted,
		DurationMs:      time.Since(result.CompletedAt).Milliseconds(),
	})

	return result, nil
}

func (s *AdminService) recordCleanupRun(ctx context.Context, run domain.MailCleanupRun) {
	if s.cleanupRuns == nil {
		return
	}
	if _, err := s.cleanupRuns.Create(ctx, run); err != nil {
		log.Printf("admin: create mail cleanup run failed: %v", err)
	}
}

func normalizedCleanupBatchSize(size int) int {
	if size <= 0 {
		return 500
	}
	return size
}

func buildMailRetentionAlerts(snapshot domain.MailRetentionSnapshot, runs []domain.MailCleanupRun, policy MailRetentionPolicy) []domain.MailRetentionAlert {
	alerts := make([]domain.MailRetentionAlert, 0, 4)

	if policy.AlertJobsThreshold > 0 && snapshot.EligibleJobs >= policy.AlertJobsThreshold {
		alerts = append(alerts, domain.MailRetentionAlert{
			Severity: "warning",
			Code:     "cleanup_jobs_backlog",
			Message:  fmt.Sprintf("Eligible retained mail jobs reached %d, above threshold %d.", snapshot.EligibleJobs, policy.AlertJobsThreshold),
		})
	}
	if policy.AlertEventsThreshold > 0 && snapshot.EligibleEvents >= policy.AlertEventsThreshold {
		alerts = append(alerts, domain.MailRetentionAlert{
			Severity: "warning",
			Code:     "cleanup_events_backlog",
			Message:  fmt.Sprintf("Eligible retained mail events reached %d, above threshold %d.", snapshot.EligibleEvents, policy.AlertEventsThreshold),
		})
	}

	if len(runs) == 0 {
		if policy.CleanupInterval > 0 {
			alerts = append(alerts, domain.MailRetentionAlert{
				Severity: "warning",
				Code:     "cleanup_history_missing",
				Message:  "No cleanup runs have been recorded yet.",
			})
		}
		return alerts
	}

	if policy.AlertStaleAfter > 0 && time.Since(runs[0].CreatedAt) > policy.AlertStaleAfter {
		alerts = append(alerts, domain.MailRetentionAlert{
			Severity: "warning",
			Code:     "cleanup_stale",
			Message:  fmt.Sprintf("Last cleanup run is older than %s.", policy.AlertStaleAfter),
		})
	}

	if streakThreshold := policy.AlertFailureStreak; streakThreshold > 0 {
		streak := 0
		for _, run := range runs {
			if run.Status != "failed" {
				break
			}
			streak++
		}
		if streak >= streakThreshold {
			alerts = append(alerts, domain.MailRetentionAlert{
				Severity: "critical",
				Code:     "cleanup_failure_streak",
				Message:  fmt.Sprintf("Cleanup failed %d times in a row.", streak),
			})
		}
	}

	return alerts
}

func (s *AdminService) CreateMailSuppression(ctx context.Context, input CreateMailSuppressionInput) (domain.MailSuppression, error) {
	if s.suppressions == nil {
		return domain.MailSuppression{}, fmt.Errorf("mail suppressions store is not configured")
	}
	item, err := s.suppressions.Create(ctx, domain.MailSuppression{
		ID:     uuid.NewString(),
		Kind:   input.Kind,
		Value:  input.Value,
		Reason: input.Reason,
	})
	if err != nil {
		return domain.MailSuppression{}, fmt.Errorf("create mail suppression: %w", err)
	}
	return item, nil
}

func (s *AdminService) DeleteMailSuppression(ctx context.Context, id string) error {
	if s.suppressions == nil {
		return fmt.Errorf("mail suppressions store is not configured")
	}
	if err := s.suppressions.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete mail suppression: %w", err)
	}
	return nil
}
