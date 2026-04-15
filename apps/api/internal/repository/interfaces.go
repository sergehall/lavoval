package repository

import (
	"context"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
)

type UserStore interface {
	Create(context.Context, domain.User) (domain.User, error)
	FindByEmail(context.Context, string) (domain.User, error)
	FindByID(context.Context, string) (domain.User, error)
	MarkEmailVerified(context.Context, string) (domain.User, error)
	UpdatePasswordHash(context.Context, string, string) (domain.User, error)
	StartTOTPEnrollment(context.Context, string, string) (domain.User, error)
	EnableTOTP(context.Context, string, string) (domain.User, error)
	DisableTOTP(context.Context, string) (domain.User, error)
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

type ProfileStore interface {
	Create(context.Context, domain.Profile) (domain.Profile, error)
	Update(context.Context, domain.Profile) (domain.Profile, error)
	FindByUserID(context.Context, string) (domain.Profile, error)
}

type SkillStore interface {
	ListPublished(context.Context) ([]domain.Skill, error)
	ListAll(context.Context) ([]domain.Skill, error)
	ListByCreatorID(context.Context, string) ([]domain.Skill, error)
	FindByID(context.Context, string) (domain.Skill, error)
	Create(context.Context, domain.Skill) (domain.Skill, error)
	Update(context.Context, domain.Skill) (domain.Skill, error)
	SoftDelete(context.Context, string) error
}

type EnrollmentStore interface {
	ListByUserID(context.Context, string) ([]domain.Enrollment, error)
}

type SkillRunStore interface {
	Create(context.Context, domain.SkillRun) (domain.SkillRun, error)
	Update(context.Context, domain.SkillRun) (domain.SkillRun, error)
	FindByID(context.Context, string) (domain.SkillRun, error)
	ListByUserID(context.Context, string) ([]domain.SkillRun, error)
	ListAll(context.Context) ([]domain.SkillRun, error)
}
