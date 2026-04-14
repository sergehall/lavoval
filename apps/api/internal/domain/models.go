package domain

import "time"

type Role string

type AccountStatus string

type SkillStatus string

type Visibility string

type EnrollmentStatus string

type SkillRunStatus string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"

	AccountStatusActive    AccountStatus = "active"
	AccountStatusInvited   AccountStatus = "invited"
	AccountStatusSuspended AccountStatus = "suspended"

	SkillStatusDraft     SkillStatus = "draft"
	SkillStatusPublished SkillStatus = "published"
	SkillStatusArchived  SkillStatus = "archived"

	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"

	EnrollmentStatusAssigned   EnrollmentStatus = "assigned"
	EnrollmentStatusInProgress EnrollmentStatus = "in_progress"
	EnrollmentStatusCompleted  EnrollmentStatus = "completed"

	SkillRunStatusQueued    SkillRunStatus = "queued"
	SkillRunStatusRunning   SkillRunStatus = "running"
	SkillRunStatusCompleted SkillRunStatus = "completed"
	SkillRunStatusFailed    SkillRunStatus = "failed"
)

type User struct {
	ID           string        `json:"id"`
	Email        string        `json:"email"`
	PasswordHash string        `json:"-"`
	Role         Role          `json:"role"`
	Status       AccountStatus `json:"status"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

type Profile struct {
	UserID    string     `json:"userId"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Bio       *string    `json:"bio"`
	Timezone  string     `json:"timezone"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

type Skill struct {
	ID           string         `json:"id"`
	Slug         string         `json:"slug"`
	Title        string         `json:"title"`
	Summary      string         `json:"summary"`
	Description  string         `json:"description"`
	Provider     string         `json:"provider"`
	Entrypoint   string         `json:"entrypoint"`
	Config       map[string]any `json:"config"`
	Status       SkillStatus    `json:"status"`
	Visibility   Visibility     `json:"visibility"`
	CreatedBy    string         `json:"createdBy"`
	Creator      Creator        `json:"creator"`
	Modules      []Module       `json:"modules,omitempty"`
	ModulesCount int            `json:"modulesCount"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

type Creator struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Module struct {
	ID        string      `json:"id"`
	SkillID   string      `json:"skillId"`
	Slug      string      `json:"slug"`
	Title     string      `json:"title"`
	Summary   string      `json:"summary"`
	Content   string      `json:"content"`
	Position  int         `json:"position"`
	Status    SkillStatus `json:"status"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

type Enrollment struct {
	ID              string           `json:"id"`
	UserID          string           `json:"userId"`
	SkillID         string           `json:"skillId"`
	Status          EnrollmentStatus `json:"status"`
	ProgressPercent int              `json:"progressPercent"`
	AssignedAt      time.Time        `json:"assignedAt"`
	CompletedAt     *time.Time       `json:"completedAt,omitempty"`
}

type SkillRun struct {
	ID           string         `json:"id"`
	SkillID      string         `json:"skillId"`
	UserID       string         `json:"userId"`
	Skill        SkillRunSkill  `json:"skill"`
	Status       SkillRunStatus `json:"status"`
	Input        map[string]any `json:"input"`
	Output       map[string]any `json:"output,omitempty"`
	Meta         SkillRunMeta   `json:"meta"`
	ErrorMessage *string        `json:"errorMessage,omitempty"`
	StartedAt    *time.Time     `json:"startedAt,omitempty"`
	FinishedAt   *time.Time     `json:"finishedAt,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
}

type SkillRunSkill struct {
	ID         string  `json:"id"`
	Slug       string  `json:"slug"`
	Title      string  `json:"title"`
	Entrypoint string  `json:"entrypoint"`
	Creator    Creator `json:"creator"`
}

type SkillRunMeta struct {
	DurationMs      *int64 `json:"durationMs,omitempty"`
	HasOutput       bool   `json:"hasOutput"`
	HasError        bool   `json:"hasError"`
	InputKeysCount  int    `json:"inputKeysCount"`
	OutputKeysCount int    `json:"outputKeysCount"`
}
