package service

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

type AdminService struct {
	users       repository.UserStore
	profiles    repository.ProfileStore
	skills      repository.SkillStore
	enrollments repository.EnrollmentStore
	modules     repository.ModuleStore
	mailJobs    repository.MailJobStore
	mailEvents  repository.MailEventStore
}

func NewAdminService(users repository.UserStore, profiles repository.ProfileStore, skills repository.SkillStore, enrollments repository.EnrollmentStore, modules repository.ModuleStore, mailJobs repository.MailJobStore, mailEvents repository.MailEventStore) *AdminService {
	return &AdminService{
		users:       users,
		profiles:    profiles,
		skills:      skills,
		enrollments: enrollments,
		modules:     modules,
		mailJobs:    mailJobs,
		mailEvents:  mailEvents,
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

func (s *AdminService) ListDeadLetters(ctx context.Context, limit int) ([]domain.MailJob, error) {
	if s.mailJobs == nil {
		return nil, fmt.Errorf("mail jobs store is not configured")
	}

	items, err := s.mailJobs.ListDeadLetters(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list dead letters: %w", err)
	}
	return items, nil
}

func (s *AdminService) ListMailEvents(ctx context.Context, limit int) ([]domain.MailEvent, error) {
	if s.mailEvents == nil {
		return nil, fmt.Errorf("mail events store is not configured")
	}

	events, err := s.mailEvents.ListRecent(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list mail events: %w", err)
	}
	return events, nil
}

func (s *AdminService) ListMailEventsByJob(ctx context.Context, jobID string, limit int) ([]domain.MailEvent, error) {
	if s.mailEvents == nil {
		return nil, fmt.Errorf("mail events store is not configured")
	}

	events, err := s.mailEvents.ListByJobID(ctx, jobID, limit)
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
