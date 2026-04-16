package repository

import (
	"context"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type UserStore interface {
	Create(context.Context, domain.User) (domain.User, error)
	FindByEmail(context.Context, string) (domain.User, error)
	FindByID(context.Context, string) (domain.User, error)
	MarkEmailVerified(context.Context, string) (domain.User, error)
	UpdatePasswordHash(context.Context, string, string) (domain.User, error)
	// UpdateRoleAndStatus is kept for backward compatibility; prefer UpdateRoleStatusModeration.
	UpdateRoleAndStatus(context.Context, string, domain.Role, domain.AccountStatus) (domain.User, error)
	// UpdateRoleStatusModeration updates role, status and the moderation audit fields atomically.
	UpdateRoleStatusModeration(ctx context.Context, id, actorID string, role domain.Role, status domain.AccountStatus, reason *string) (domain.User, error)
	BumpSessionVersion(context.Context, string) (domain.User, error)
	// GetStats returns aggregate user counts for the admin dashboard.
	GetStats(context.Context) (domain.AdminUserStats, error)
	StartTOTPEnrollment(context.Context, string, string) (domain.User, error)
	CancelTOTPEnrollment(context.Context, string) (domain.User, error)
	EnableTOTP(context.Context, string, string) (domain.User, error)
	DisableTOTP(context.Context, string) (domain.User, error)
	SoftDelete(context.Context, string) error
	List(context.Context) ([]domain.User, error)
}

type EmailVerificationStore interface {
	Create(context.Context, domain.EmailVerificationToken) (domain.EmailVerificationToken, error)
	FindByTokenHash(context.Context, string) (domain.EmailVerificationToken, error)
	Consume(context.Context, string, string) error
	RevokeActiveByUserID(context.Context, string) error
}

type PasswordResetStore interface {
	Create(context.Context, domain.PasswordResetToken) (domain.PasswordResetToken, error)
	FindByTokenHash(context.Context, string) (domain.PasswordResetToken, error)
	Consume(context.Context, string, string) error
	RevokeActiveByUserID(context.Context, string) error
}

type MailJobStore interface {
	Enqueue(context.Context, domain.MailJob) (domain.MailJob, bool, error)
	ClaimNext(context.Context, time.Duration) (domain.MailJob, bool, error)
	MarkSent(context.Context, string, string, string) error
	MarkRetry(context.Context, string, string, string, time.Time) error
	MarkDeadLetter(context.Context, string, string, string) error
	CountByStatus(context.Context) (map[domain.MailJobStatus]int64, error)
	OperationalSnapshot(context.Context) (domain.MailOperationalSnapshot, error)
	ListDeadLetters(context.Context, domain.MailJobFilter) ([]domain.MailJob, error)
	RequeueDeadLetter(context.Context, string) (domain.MailJob, error)
	FindByID(context.Context, string) (domain.MailJob, error)
	Replay(context.Context, domain.MailJob) (domain.MailJob, bool, error)
	CountTerminalBefore(context.Context, time.Time) (int64, error)
	DeleteTerminalBefore(context.Context, time.Time, int) (int64, error)
}

type MailEventStore interface {
	Append(context.Context, domain.MailEvent) (domain.MailEvent, error)
	ListRecent(context.Context, domain.MailEventFilter) ([]domain.MailEvent, error)
	ListByJobID(context.Context, string, domain.MailEventFilter) ([]domain.MailEvent, error)
	CountBefore(context.Context, time.Time) (int64, error)
	DeleteBefore(context.Context, time.Time, int) (int64, error)
}

type MailSuppressionStore interface {
	Create(context.Context, domain.MailSuppression) (domain.MailSuppression, error)
	Delete(context.Context, string) error
	List(context.Context) ([]domain.MailSuppression, error)
	FindMatch(context.Context, string) (*domain.MailSuppression, error)
}

type MailCleanupRunStore interface {
	Create(context.Context, domain.MailCleanupRun) (domain.MailCleanupRun, error)
	ListRecent(context.Context, int) ([]domain.MailCleanupRun, error)
}

type MFARecoveryCodeStore interface {
	ReplaceForUser(context.Context, string, []domain.MFARecoveryCode) error
	FindActiveByCodeHash(context.Context, string, string) (domain.MFARecoveryCode, error)
	Consume(context.Context, string, string) error
	RevokeActiveByUserID(context.Context, string) error
}

type SignInChallengeStore interface {
	Create(context.Context, domain.AuthSignInChallenge) (domain.AuthSignInChallenge, error)
	FindByID(context.Context, string) (domain.AuthSignInChallenge, error)
	Consume(context.Context, string, string) error
	RevokeActiveByUserID(context.Context, string) error
}

type OAuthStateStore interface {
	Create(context.Context, domain.OAuthState) (domain.OAuthState, error)
	FindByStateHash(context.Context, domain.OAuthProvider, string) (domain.OAuthState, error)
	Consume(context.Context, string) error
}

type OAuthIdentityStore interface {
	Create(context.Context, domain.OAuthIdentity) (domain.OAuthIdentity, error)
	FindByProviderSubject(context.Context, domain.OAuthProvider, string) (domain.OAuthIdentity, error)
	ListByUserID(context.Context, string) ([]domain.OAuthIdentity, error)
}

type ProfileStore interface {
	Create(context.Context, domain.Profile) (domain.Profile, error)
	Update(context.Context, domain.Profile) (domain.Profile, error)
	FindByUserID(context.Context, string) (domain.Profile, error)
	SoftDeleteByUserID(context.Context, string) error
}

type SkillStore interface {
	ListPublished(context.Context) ([]domain.Skill, error)
	ListAll(context.Context) ([]domain.Skill, error)
	ListByCreatorID(context.Context, string) ([]domain.Skill, error)
	FindByID(context.Context, string) (domain.Skill, error)
	Create(context.Context, domain.Skill) (domain.Skill, error)
	Update(context.Context, domain.Skill) (domain.Skill, error)
	// UpdateGovernance sets moderation status, reason, and featured/verified flags atomically.
	UpdateGovernance(ctx context.Context, id, actorID string, status domain.SkillStatus, reason *string, featured, verified bool) (domain.Skill, error)
	// UpdatePricing updates price and access control fields.
	UpdatePricing(ctx context.Context, id string, priceCents int, currency string, accessType domain.SkillAccessType) (domain.Skill, error)
	// GetStats returns aggregate skill counts for the admin dashboard.
	GetStats(context.Context) (domain.AdminSkillStats, error)
	SoftDelete(context.Context, string) error
}

type EnrollmentStore interface {
	ListByUserID(context.Context, string) ([]domain.Enrollment, error)
	ListAll(context.Context) ([]domain.EnrollmentDetail, error)
	Create(context.Context, string, string) (domain.Enrollment, error)
	UpdateStatus(context.Context, string, domain.EnrollmentStatus, int) (domain.Enrollment, error)
}

type ModuleStore interface {
	ListBySkillID(context.Context, string) ([]domain.Module, error)
	FindByID(context.Context, string) (domain.Module, error)
	Create(context.Context, domain.Module) (domain.Module, error)
	Update(context.Context, domain.Module) (domain.Module, error)
	SoftDelete(context.Context, string) error
}

type SkillRunStore interface {
	Create(context.Context, domain.SkillRun) (domain.SkillRun, error)
	Update(context.Context, domain.SkillRun) (domain.SkillRun, error)
	FindByID(context.Context, string) (domain.SkillRun, error)
	ListByUserID(context.Context, string) ([]domain.SkillRun, error)
	ListAll(context.Context) ([]domain.SkillRun, error)
}

// AdminAuditLogStore persists admin governance actions.
type AdminAuditLogStore interface {
	Create(context.Context, domain.AdminAuditLog) (domain.AdminAuditLog, error)
	ListByEntity(ctx context.Context, entityType, entityID string, limit int) ([]domain.AdminAuditLog, error)
}

// SkillAccessStore tracks explicit user access grants for paid / invite-only skills.
type SkillAccessStore interface {
	Create(context.Context, domain.SkillAccess) (domain.SkillAccess, error)
	FindBySkillAndUser(ctx context.Context, skillID, userID string) (*domain.SkillAccess, error)
	ListBySkillID(ctx context.Context, skillID string) ([]domain.SkillAccess, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.SkillAccess, error)
}
