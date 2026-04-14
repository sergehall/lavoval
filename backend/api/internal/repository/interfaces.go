package repository

import (
	"context"

	"github.com/sergehall/lavoval/backend/api/internal/domain"
)

type UserStore interface {
	Create(context.Context, domain.User) (domain.User, error)
	FindByEmail(context.Context, string) (domain.User, error)
	FindByID(context.Context, string) (domain.User, error)
	List(context.Context) ([]domain.User, error)
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
