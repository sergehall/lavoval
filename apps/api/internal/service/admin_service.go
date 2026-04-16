package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
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
	auditLog     repository.AdminAuditLogStore
	skillAccess  repository.SkillAccessStore
	sessions     SessionRevoker
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

// AdminServiceOption allows optional stores to be injected without changing the base signature.
type AdminServiceOption func(*AdminService)

func WithAuditLog(store repository.AdminAuditLogStore) AdminServiceOption {
	return func(s *AdminService) { s.auditLog = store }
}

func WithSkillAccess(store repository.SkillAccessStore) AdminServiceOption {
	return func(s *AdminService) { s.skillAccess = store }
}

func WithSessionRevoker(revoker SessionRevoker) AdminServiceOption {
	return func(s *AdminService) { s.sessions = revoker }
}

func NewAdminService(
	users repository.UserStore,
	profiles repository.ProfileStore,
	skills repository.SkillStore,
	enrollments repository.EnrollmentStore,
	modules repository.ModuleStore,
	mailJobs repository.MailJobStore,
	mailEvents repository.MailEventStore,
	suppressions repository.MailSuppressionStore,
	cleanupRuns repository.MailCleanupRunStore,
	retention MailRetentionPolicy,
	opts ...AdminServiceOption,
) *AdminService {
	svc := &AdminService{
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
	for _, o := range opts {
		o(svc)
	}
	if svc.sessions == nil {
		svc.sessions = NoopSessionRevoker{}
	}
	return svc
}

// ── Types ─────────────────────────────────────────────────────────────────────

type UserDetail struct {
	User    domain.User    `json:"user"`
	Profile domain.Profile `json:"profile"`
}

type UpdateUserInput struct {
	Role   domain.Role          `json:"role"   validate:"required,oneof=user admin root_owner"`
	Status domain.AccountStatus `json:"status" validate:"required,oneof=active invited suspended blocked"`
	Reason *string              `json:"reason,omitempty"`
}

type UpdateUserStatusInput struct {
	Status domain.AccountStatus `json:"status" validate:"required,oneof=active invited suspended blocked"`
	Reason *string              `json:"reason,omitempty"`
}

type UpdateUserRoleInput struct {
	Role   domain.Role `json:"role" validate:"required,oneof=user admin root_owner"`
	Reason *string     `json:"reason,omitempty"`
}

type CreateMailSuppressionInput struct {
	Kind   domain.MailSuppressionKind `json:"kind" validate:"required,oneof=email domain"`
	Value  string                     `json:"value" validate:"required,min=3,max=255"`
	Reason string                     `json:"reason" validate:"required,min=3,max=255"`
}

type SkillGovernanceInput struct {
	Status   domain.SkillStatus `json:"status"   validate:"required,oneof=draft pending_review published hidden archived rejected"`
	Reason   *string            `json:"reason,omitempty"`
	Featured bool               `json:"featured"`
	Verified bool               `json:"verified"`
}

type SkillPricingInput struct {
	PriceCents int                    `json:"priceCents" validate:"min=0"`
	Currency   string                 `json:"currency"   validate:"required,len=3"`
	AccessType domain.SkillAccessType `json:"accessType" validate:"required,oneof=free paid invite_only"`
}

var ErrAdminReasonRequired = errors.New("reason is required for this admin action")
var ErrInvalidSkillPricing = errors.New("invalid skill pricing")
var ErrOnlyRootOwnerCanManageRoles = errors.New("only root_owner can manage privileged roles")
var ErrCannotChangeOwnRole = errors.New("cannot change your own role")
var ErrRootOwnerRequiresMFA = errors.New("root_owner requires mfa")
var ErrPrivilegedRoleRequiresActiveVerifiedAccount = errors.New("privileged role requires active verified account")
var ErrLastRootOwnerDemotion = errors.New("cannot demote the last root_owner")
var ErrPrivilegedUserModerationRequiresRootOwner = errors.New("only root_owner can moderate privileged users")
var ErrLastRootOwnerStatusLockout = errors.New("cannot suspend or block the last root_owner")

// ── User methods ──────────────────────────────────────────────────────────────

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

// UpdateUser changes role/status with full audit trail.
// actorID is the admin performing the action (from JWT claims).
func (s *AdminService) UpdateUser(ctx context.Context, actorID, targetID string, input UpdateUserInput) (domain.User, error) {
	actor, err := s.users.FindByID(ctx, actorID)
	if err != nil {
		return domain.User{}, fmt.Errorf("load actor: %w", err)
	}

	if requiresUserModerationReason(input.Status) && isBlankPtr(input.Reason) {
		return domain.User{}, ErrAdminReasonRequired
	}

	before, err := s.users.FindByID(ctx, targetID)
	if err != nil {
		return domain.User{}, fmt.Errorf("load user before update: %w", err)
	}

	if before.Role != input.Role {
		if err := s.ensureRoleChangeAllowed(ctx, actor, before, input.Role); err != nil {
			return domain.User{}, err
		}
	}
	if err := s.ensureStatusChangeAllowed(ctx, actor, before, input.Status); err != nil {
		return domain.User{}, err
	}
	if isBlankPtr(input.Reason) && before.Role != input.Role {
		return domain.User{}, ErrAdminReasonRequired
	}

	user, err := s.users.UpdateRoleStatusModeration(ctx, targetID, actorID, input.Role, input.Status, input.Reason)
	if err != nil {
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}

	s.writeAuditLog(ctx, domain.AdminAuditLog{
		ID:           uuid.NewString(),
		EntityType:   "user",
		EntityID:     targetID,
		Action:       auditActionForStatus(input.Status),
		OldValueJSON: marshalJSON(userAuditSnapshot(before)),
		NewValueJSON: marshalJSON(userAuditSnapshot(user)),
		Reason:       input.Reason,
		ActorID:      actorID,
	})

	if before.Role != input.Role || before.Status != input.Status {
		if err := s.sessions.RevokeAllSessionsForUser(ctx, targetID, "admin_user_update"); err != nil {
			return domain.User{}, fmt.Errorf("revoke sessions after user update: %w", err)
		}
	}

	return user, nil
}

func (s *AdminService) UpdateUserStatus(ctx context.Context, actorID, targetID string, input UpdateUserStatusInput) (domain.User, error) {
	if requiresUserModerationReason(input.Status) && isBlankPtr(input.Reason) {
		return domain.User{}, ErrAdminReasonRequired
	}

	actor, err := s.users.FindByID(ctx, actorID)
	if err != nil {
		return domain.User{}, fmt.Errorf("load actor: %w", err)
	}

	before, err := s.users.FindByID(ctx, targetID)
	if err != nil {
		return domain.User{}, fmt.Errorf("load user before status update: %w", err)
	}
	if err := s.ensureStatusChangeAllowed(ctx, actor, before, input.Status); err != nil {
		return domain.User{}, err
	}

	user, err := s.users.UpdateRoleStatusModeration(ctx, targetID, actorID, before.Role, input.Status, input.Reason)
	if err != nil {
		return domain.User{}, fmt.Errorf("update user status: %w", err)
	}

	s.writeAuditLog(ctx, domain.AdminAuditLog{
		ID:           uuid.NewString(),
		EntityType:   "user",
		EntityID:     targetID,
		Action:       auditActionForStatus(input.Status),
		OldValueJSON: marshalJSON(userAuditSnapshot(before)),
		NewValueJSON: marshalJSON(userAuditSnapshot(user)),
		Reason:       input.Reason,
		ActorID:      actorID,
	})

	if before.Status != input.Status {
		if err := s.sessions.RevokeAllSessionsForUser(ctx, targetID, "admin_user_status_update"); err != nil {
			return domain.User{}, fmt.Errorf("revoke sessions after status update: %w", err)
		}
	}

	return user, nil
}

func (s *AdminService) UpdateUserRole(ctx context.Context, actorID, targetID string, input UpdateUserRoleInput) (domain.User, error) {
	if isBlankPtr(input.Reason) {
		return domain.User{}, ErrAdminReasonRequired
	}

	actor, err := s.users.FindByID(ctx, actorID)
	if err != nil {
		return domain.User{}, fmt.Errorf("load actor: %w", err)
	}

	before, err := s.users.FindByID(ctx, targetID)
	if err != nil {
		return domain.User{}, fmt.Errorf("load user before role update: %w", err)
	}

	if before.Role == input.Role {
		return before, nil
	}

	if err := s.ensureRoleChangeAllowed(ctx, actor, before, input.Role); err != nil {
		return domain.User{}, err
	}

	user, err := s.users.UpdateRoleAndStatus(ctx, targetID, input.Role, before.Status)
	if err != nil {
		return domain.User{}, fmt.Errorf("update user role: %w", err)
	}

	s.writeAuditLog(ctx, domain.AdminAuditLog{
		ID:           uuid.NewString(),
		EntityType:   "user",
		EntityID:     targetID,
		Action:       roleAuditAction(before.Role, input.Role),
		OldValueJSON: marshalJSON(userAuditSnapshot(before)),
		NewValueJSON: marshalJSON(userAuditSnapshot(user)),
		Reason:       input.Reason,
		ActorID:      actorID,
	})

	if err := s.sessions.RevokeAllSessionsForUser(ctx, targetID, "admin_user_role_update"); err != nil {
		return domain.User{}, fmt.Errorf("revoke sessions after role update: %w", err)
	}

	return user, nil
}

func auditActionForStatus(status domain.AccountStatus) string {
	switch status {
	case domain.AccountStatusSuspended:
		return "user_suspended"
	case domain.AccountStatusBlocked:
		return "user_blocked"
	case domain.AccountStatusActive:
		return "user_restored"
	default:
		return "user_updated"
	}
}

func roleAuditAction(before, after domain.Role) string {
	return fmt.Sprintf("user_role_changed:%s->%s", before, after)
}

// GetUserAuditLog returns the governance history for a user.
func (s *AdminService) GetUserAuditLog(ctx context.Context, userID string, limit int) ([]domain.AdminAuditLog, error) {
	if s.auditLog == nil {
		return []domain.AdminAuditLog{}, nil
	}
	entries, err := s.auditLog.ListByEntity(ctx, "user", userID, limit)
	if err != nil {
		return nil, fmt.Errorf("user audit log: %w", err)
	}
	return entries, nil
}

// GetAdminStats returns the top-level admin dashboard snapshot.
func (s *AdminService) GetAdminStats(ctx context.Context) (domain.AdminStats, error) {
	userStats, err := s.users.GetStats(ctx)
	if err != nil {
		return domain.AdminStats{}, fmt.Errorf("user stats: %w", err)
	}
	skillStats, err := s.skills.GetStats(ctx)
	if err != nil {
		return domain.AdminStats{}, fmt.Errorf("skill stats: %w", err)
	}
	return domain.AdminStats{Users: userStats, Skills: skillStats}, nil
}

// ── Skill governance methods ──────────────────────────────────────────────────

func (s *AdminService) ListSkills(ctx context.Context) ([]domain.Skill, error) {
	skills, err := s.skills.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list admin skills: %w", err)
	}
	return skills, nil
}

// GovernSkill updates skill moderation status and writes an audit log entry.
func (s *AdminService) GovernSkill(ctx context.Context, actorID, skillID string, input SkillGovernanceInput) (domain.Skill, error) {
	if requiresSkillModerationReason(input.Status) && isBlankPtr(input.Reason) {
		return domain.Skill{}, ErrAdminReasonRequired
	}

	before, err := s.skills.FindByID(ctx, skillID)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("load skill before governance update: %w", err)
	}

	skill, err := s.skills.UpdateGovernance(ctx, skillID, actorID, input.Status, input.Reason, input.Featured, input.Verified)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("govern skill: %w", err)
	}

	s.writeAuditLog(ctx, domain.AdminAuditLog{
		ID:           uuid.NewString(),
		EntityType:   "skill",
		EntityID:     skillID,
		Action:       "skill_" + string(input.Status),
		OldValueJSON: marshalJSON(skillGovernanceAuditSnapshot(before)),
		NewValueJSON: marshalJSON(skillGovernanceAuditSnapshot(skill)),
		Reason:       input.Reason,
		ActorID:      actorID,
	})

	return skill, nil
}

// UpdateSkillPricing sets the price and access model for a skill.
func (s *AdminService) UpdateSkillPricing(ctx context.Context, actorID, skillID string, input SkillPricingInput) (domain.Skill, error) {
	before, err := s.skills.FindByID(ctx, skillID)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("load skill before pricing update: %w", err)
	}

	normalized, err := normalizeSkillPricingInput(input)
	if err != nil {
		return domain.Skill{}, err
	}

	skill, err := s.skills.UpdatePricing(ctx, skillID, normalized.PriceCents, normalized.Currency, normalized.AccessType)
	if err != nil {
		return domain.Skill{}, fmt.Errorf("update skill pricing: %w", err)
	}

	s.writeAuditLog(ctx, domain.AdminAuditLog{
		ID:           uuid.NewString(),
		EntityType:   "skill",
		EntityID:     skillID,
		Action:       "skill_pricing_updated",
		OldValueJSON: marshalJSON(skillPricingAuditSnapshot(before)),
		NewValueJSON: marshalJSON(skillPricingAuditSnapshot(skill)),
		ActorID:      actorID,
	})

	return skill, nil
}

// GetSkillAuditLog returns the governance history for a skill.
func (s *AdminService) GetSkillAuditLog(ctx context.Context, skillID string, limit int) ([]domain.AdminAuditLog, error) {
	if s.auditLog == nil {
		return []domain.AdminAuditLog{}, nil
	}
	entries, err := s.auditLog.ListByEntity(ctx, "skill", skillID, limit)
	if err != nil {
		return nil, fmt.Errorf("skill audit log: %w", err)
	}
	return entries, nil
}

// ── Enrollment methods ────────────────────────────────────────────────────────

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

// ── Module methods ────────────────────────────────────────────────────────────

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

// ── Mail methods ──────────────────────────────────────────────────────────────

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
		JobsRetention:         s.retention.JobsRetention.String(),
		EventsRetention:       s.retention.EventsRetention.String(),
		CleanupBatchSize:      normalizedCleanupBatchSize(s.retention.CleanupBatchSize),
		CleanupInterval:       s.retention.CleanupInterval.String(),
		CleanupDryRun:         s.retention.CleanupDryRun,
		AutoCleanupEnabled:    s.retention.CleanupInterval > 0,
		JobsRetentionActive:   s.retention.JobsRetention > 0,
		EventsRetentionActive: s.retention.EventsRetention > 0,
		Alerts:                make([]domain.MailRetentionAlert, 0),
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
	result := domain.MailCleanupResult{CompletedAt: time.Now().UTC()}

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

// ── Internal helpers ──────────────────────────────────────────────────────────

func (s *AdminService) writeAuditLog(ctx context.Context, entry domain.AdminAuditLog) {
	if s.auditLog == nil {
		return
	}
	if _, err := s.auditLog.Create(ctx, entry); err != nil {
		log.Printf("admin: write audit log failed (entity=%s id=%s action=%s): %v",
			entry.EntityType, entry.EntityID, entry.Action, err)
	}
}

// marshalJSON is a helper for building audit log old/new values.
func marshalJSON(v any) *string {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}

func normalizedCleanupBatchSize(size int) int {
	if size <= 0 {
		return 500
	}
	return size
}

func requiresUserModerationReason(status domain.AccountStatus) bool {
	return status == domain.AccountStatusSuspended || status == domain.AccountStatusBlocked
}

func requiresSkillModerationReason(status domain.SkillStatus) bool {
	return status == domain.SkillStatusHidden || status == domain.SkillStatusRejected
}

func isBlankPtr(value *string) bool {
	return value == nil || strings.TrimSpace(*value) == ""
}

func (s *AdminService) ensureRoleChangeAllowed(ctx context.Context, actor, target domain.User, nextRole domain.Role) error {
	if actor.Role != domain.RoleRootOwner {
		return ErrOnlyRootOwnerCanManageRoles
	}
	if actor.ID == target.ID {
		return ErrCannotChangeOwnRole
	}
	if domain.CanAccessAdmin(nextRole) {
		if target.Status != domain.AccountStatusActive || target.EmailVerifiedAt == nil {
			return ErrPrivilegedRoleRequiresActiveVerifiedAccount
		}
	}
	if nextRole == domain.RoleRootOwner && !target.MFAEnabled {
		return ErrRootOwnerRequiresMFA
	}
	if target.Role == domain.RoleRootOwner && nextRole != domain.RoleRootOwner {
		users, err := s.users.List(ctx)
		if err != nil {
			return fmt.Errorf("list users for root_owner check: %w", err)
		}
		rootOwners := 0
		for _, user := range users {
			if user.Role == domain.RoleRootOwner && user.Status == domain.AccountStatusActive {
				rootOwners++
			}
		}
		if rootOwners <= 1 {
			return ErrLastRootOwnerDemotion
		}
	}
	return nil
}

func (s *AdminService) ensureStatusChangeAllowed(ctx context.Context, actor, target domain.User, nextStatus domain.AccountStatus) error {
	if nextStatus == target.Status {
		return nil
	}
	if domain.CanAccessAdmin(target.Role) && actor.Role != domain.RoleRootOwner {
		return ErrPrivilegedUserModerationRequiresRootOwner
	}
	if target.Role == domain.RoleRootOwner && nextStatus != domain.AccountStatusActive {
		users, err := s.users.List(ctx)
		if err != nil {
			return fmt.Errorf("list users for root_owner status check: %w", err)
		}
		rootOwners := 0
		for _, user := range users {
			if user.Role == domain.RoleRootOwner && user.Status == domain.AccountStatusActive {
				rootOwners++
			}
		}
		if rootOwners <= 1 {
			return ErrLastRootOwnerStatusLockout
		}
	}
	return nil
}

func normalizeSkillPricingInput(input SkillPricingInput) (SkillPricingInput, error) {
	normalized := input
	normalized.Currency = strings.ToUpper(strings.TrimSpace(normalized.Currency))
	if normalized.Currency == "" {
		normalized.Currency = "USD"
	}

	switch normalized.AccessType {
	case domain.AccessTypeFree:
		normalized.PriceCents = 0
	case domain.AccessTypePaid:
		if normalized.PriceCents <= 0 {
			return SkillPricingInput{}, ErrInvalidSkillPricing
		}
	case domain.AccessTypeInviteOnly:
		normalized.PriceCents = 0
	default:
		return SkillPricingInput{}, ErrInvalidSkillPricing
	}

	return normalized, nil
}

func userAuditSnapshot(user domain.User) map[string]any {
	return map[string]any{
		"role":             user.Role,
		"status":           user.Status,
		"suspensionReason": user.SuspensionReason,
		"blockReason":      user.BlockReason,
		"suspendedAt":      user.SuspendedAt,
		"blockedAt":        user.BlockedAt,
	}
}

func skillGovernanceAuditSnapshot(skill domain.Skill) map[string]any {
	return map[string]any{
		"status":           skill.Status,
		"isFeatured":       skill.IsFeatured,
		"isVerified":       skill.IsVerified,
		"moderationReason": skill.ModerationReason,
		"moderatedBy":      skill.ModeratedBy,
		"moderatedAt":      skill.ModeratedAt,
	}
}

func skillPricingAuditSnapshot(skill domain.Skill) map[string]any {
	return map[string]any{
		"priceCents": skill.PriceCents,
		"currency":   skill.Currency,
		"accessType": skill.AccessType,
	}
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
