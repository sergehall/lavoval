package domain

import "time"

type Role string

type AccountStatus string

type SkillStatus string

type Visibility string

type EnrollmentStatus string

type SkillRunStatus string

type AvailabilityStatus string

type OAuthProvider string
type MailJobStatus string
type MailSuppressionKind string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"

	AccountStatusActive    AccountStatus = "active"
	AccountStatusInvited   AccountStatus = "invited"
	AccountStatusSuspended AccountStatus = "suspended"

	AvailabilityOpen    AvailabilityStatus = "open"
	AvailabilityLimited AvailabilityStatus = "limited"
	AvailabilityClosed  AvailabilityStatus = "closed"

	OAuthProviderGoogle OAuthProvider = "google"
	OAuthProviderGitHub OAuthProvider = "github"

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

	MailJobStatusQueued     MailJobStatus = "queued"
	MailJobStatusRetrying   MailJobStatus = "retrying"
	MailJobStatusProcessing MailJobStatus = "processing"
	MailJobStatusSent       MailJobStatus = "sent"
	MailJobStatusDeadLetter MailJobStatus = "dead_letter"

	MailSuppressionKindEmail  MailSuppressionKind = "email"
	MailSuppressionKindDomain MailSuppressionKind = "domain"
)

type User struct {
	ID                            string        `json:"id"`
	Email                         string        `json:"email"`
	PasswordHash                  string        `json:"-"`
	Role                          Role          `json:"role"`
	Status                        AccountStatus `json:"status"`
	EmailVerifiedAt               *time.Time    `json:"emailVerifiedAt,omitempty"`
	MFAEnabled                    bool          `json:"mfaEnabled"`
	MFATOTPSecretEncrypted        *string       `json:"-"`
	MFAPendingTOTPSecretEncrypted *string       `json:"-"`
	MFAEnrolledAt                 *time.Time    `json:"mfaEnrolledAt,omitempty"`
	CreatedAt                     time.Time     `json:"createdAt"`
	UpdatedAt                     time.Time     `json:"updatedAt"`
}

type EmailVerificationToken struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	TokenHash  string     `json:"-"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	ConsumedAt *time.Time `json:"consumedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type PasswordResetToken struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	TokenHash  string     `json:"-"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	ConsumedAt *time.Time `json:"consumedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type MailJob struct {
	ID                string        `json:"id"`
	MessageType       string        `json:"messageType"`
	RecipientEmail    string        `json:"recipientEmail"`
	IdempotencyKey    *string       `json:"idempotencyKey,omitempty"`
	Payload           []byte        `json:"-"`
	Status            MailJobStatus `json:"status"`
	Attempts          int           `json:"attempts"`
	MaxAttempts       int           `json:"maxAttempts"`
	NextAttemptAt     time.Time     `json:"nextAttemptAt"`
	LeasedUntil       *time.Time    `json:"leasedUntil,omitempty"`
	LastError         *string       `json:"lastError,omitempty"`
	LastErrorCode     *string       `json:"lastErrorCode,omitempty"`
	Provider          *string       `json:"provider,omitempty"`
	ProviderMessageID *string       `json:"providerMessageId,omitempty"`
	SentAt            *time.Time    `json:"sentAt,omitempty"`
	DeadLetteredAt    *time.Time    `json:"deadLetteredAt,omitempty"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

type MailOperationalSnapshot struct {
	CountsByStatus         map[MailJobStatus]int64 `json:"countsByStatus"`
	DeadLettersByErrorCode map[string]int64        `json:"deadLettersByErrorCode"`
	OldestReadyAgeSeconds  float64                 `json:"oldestReadyAgeSeconds"`
}

type MailEvent struct {
	ID                string    `json:"id"`
	JobID             string    `json:"jobId"`
	EventType         string    `json:"eventType"`
	MessageType       string    `json:"messageType"`
	Provider          *string   `json:"provider,omitempty"`
	ProviderMessageID *string   `json:"providerMessageId,omitempty"`
	RecipientEmail    string    `json:"recipientEmail"`
	ErrorCode         *string   `json:"errorCode,omitempty"`
	Attempt           *int      `json:"attempt,omitempty"`
	Metadata          []byte    `json:"-"`
	CreatedAt         time.Time `json:"createdAt"`
}

type MailJobFilter struct {
	Query       string `json:"query,omitempty"`
	MessageType string `json:"messageType,omitempty"`
	Provider    string `json:"provider,omitempty"`
	ErrorCode   string `json:"errorCode,omitempty"`
	Limit       int    `json:"limit,omitempty"`
}

type MailEventFilter struct {
	Query       string `json:"query,omitempty"`
	JobID       string `json:"jobId,omitempty"`
	EventType   string `json:"eventType,omitempty"`
	MessageType string `json:"messageType,omitempty"`
	Provider    string `json:"provider,omitempty"`
	ErrorCode   string `json:"errorCode,omitempty"`
	Limit       int    `json:"limit,omitempty"`
}

type MailSuppression struct {
	ID        string              `json:"id"`
	Kind      MailSuppressionKind `json:"kind"`
	Value     string              `json:"value"`
	Reason    string              `json:"reason"`
	CreatedAt time.Time           `json:"createdAt"`
}

type MailRetentionSnapshot struct {
	JobsRetention         string               `json:"jobsRetention"`
	EventsRetention       string               `json:"eventsRetention"`
	CleanupBatchSize      int                  `json:"cleanupBatchSize"`
	CleanupInterval       string               `json:"cleanupInterval"`
	CleanupDryRun         bool                 `json:"cleanupDryRun"`
	AutoCleanupEnabled    bool                 `json:"autoCleanupEnabled"`
	JobsCutoff            *time.Time           `json:"jobsCutoff,omitempty"`
	EventsCutoff          *time.Time           `json:"eventsCutoff,omitempty"`
	EligibleJobs          int64                `json:"eligibleJobs"`
	EligibleEvents        int64                `json:"eligibleEvents"`
	JobsRetentionActive   bool                 `json:"jobsRetentionActive"`
	EventsRetentionActive bool                 `json:"eventsRetentionActive"`
	LatestCleanupRun      *MailCleanupRun      `json:"latestCleanupRun,omitempty"`
	Alerts                []MailRetentionAlert `json:"alerts"`
}

type MailCleanupResult struct {
	JobsDeleted     int64     `json:"jobsDeleted"`
	EventsDeleted   int64     `json:"eventsDeleted"`
	RemainingJobs   int64     `json:"remainingJobs"`
	RemainingEvents int64     `json:"remainingEvents"`
	CompletedAt     time.Time `json:"completedAt"`
}

type MailCleanupRun struct {
	ID              string    `json:"id"`
	Mode            string    `json:"mode"`
	Status          string    `json:"status"`
	DryRun          bool      `json:"dryRun"`
	CandidateJobs   int64     `json:"candidateJobs"`
	CandidateEvents int64     `json:"candidateEvents"`
	DeletedJobs     int64     `json:"deletedJobs"`
	DeletedEvents   int64     `json:"deletedEvents"`
	ErrorMessage    *string   `json:"errorMessage,omitempty"`
	DurationMs      int64     `json:"durationMs"`
	CreatedAt       time.Time `json:"createdAt"`
}

type MailRetentionAlert struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type MFARecoveryCode struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	CodeHash   string     `json:"-"`
	ConsumedAt *time.Time `json:"consumedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type AuthSignInChallenge struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	ConsumedAt *time.Time `json:"consumedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type OAuthState struct {
	ID         string        `json:"id"`
	Provider   OAuthProvider `json:"provider"`
	StateHash  string        `json:"-"`
	ExpiresAt  time.Time     `json:"expiresAt"`
	ConsumedAt *time.Time    `json:"consumedAt,omitempty"`
	CreatedAt  time.Time     `json:"createdAt"`
}

type OAuthIdentity struct {
	ID             string        `json:"id"`
	UserID         string        `json:"userId"`
	Provider       OAuthProvider `json:"provider"`
	ProviderUserID string        `json:"providerUserId"`
	Email          string        `json:"email"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

type Profile struct {
	UserID             string             `json:"userId"`
	Role               Role               `json:"role"`
	FirstName          string             `json:"firstName"`
	LastName           string             `json:"lastName"`
	Bio                *string            `json:"bio"`
	Timezone           string             `json:"timezone"`
	Username           *string            `json:"username"`
	AvatarURL          *string            `json:"avatarUrl"`
	Location           *string            `json:"location"`
	Skills             []string           `json:"skills"`
	Languages          []string           `json:"languages"`
	WebsiteURL         *string            `json:"websiteUrl"`
	LinkedInURL        *string            `json:"linkedinUrl"`
	GitHubURL          *string            `json:"githubUrl"`
	TwitterURL         *string            `json:"twitterUrl"`
	AvailabilityStatus AvailabilityStatus `json:"availabilityStatus"`
	IsPublicProfile    bool               `json:"isPublicProfile"`
	CreatedAt          time.Time          `json:"createdAt"`
	UpdatedAt          time.Time          `json:"updatedAt"`
	DeletedAt          *time.Time         `json:"deletedAt,omitempty"`
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

// EnrollmentDetail enriches an Enrollment with denormalised display fields
// returned by admin list queries (JOIN with users + skills tables).
type EnrollmentDetail struct {
	Enrollment
	UserEmail  string `json:"userEmail"`
	SkillTitle string `json:"skillTitle"`
	SkillSlug  string `json:"skillSlug"`
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
