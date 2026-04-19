package domain

import "time"

type Role string

type AccountStatus string

type SkillStatus string

type SkillAccessType string

type Visibility string

type EnrollmentStatus string

type SkillRunStatus string

type AvailabilityStatus string

type OAuthProvider string
type MailJobStatus string
type MailSuppressionKind string

const (
	RoleUser      Role = "user"
	RoleAdmin     Role = "admin"
	RoleRootOwner Role = "root_owner"

	AccountStatusActive    AccountStatus = "active"
	AccountStatusInvited   AccountStatus = "invited"
	AccountStatusSuspended AccountStatus = "suspended"
	AccountStatusBlocked   AccountStatus = "blocked"

	AvailabilityOpen    AvailabilityStatus = "open"
	AvailabilityLimited AvailabilityStatus = "limited"
	AvailabilityClosed  AvailabilityStatus = "closed"

	OAuthProviderGoogle OAuthProvider = "google"
	OAuthProviderGitHub OAuthProvider = "github"

	SkillStatusDraft         SkillStatus = "draft"
	SkillStatusPendingReview SkillStatus = "pending_review"
	SkillStatusPublished     SkillStatus = "published"
	SkillStatusHidden        SkillStatus = "hidden"
	SkillStatusArchived      SkillStatus = "archived"
	SkillStatusRejected      SkillStatus = "rejected"

	AccessTypeFree       SkillAccessType = "free"
	AccessTypePaid       SkillAccessType = "paid"
	AccessTypeInviteOnly SkillAccessType = "invite_only"

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

func RoleRank(role Role) int {
	switch role {
	case RoleRootOwner:
		return 3
	case RoleAdmin:
		return 2
	case RoleUser:
		return 1
	default:
		return 0
	}
}

func RoleAtLeast(role Role, required Role) bool {
	return RoleRank(role) >= RoleRank(required)
}

func CanAccessAdmin(role Role) bool {
	return RoleAtLeast(role, RoleAdmin)
}

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
	SessionVersion                int           `json:"sessionVersion"`
	SuspensionReason              *string       `json:"suspensionReason,omitempty"`
	SuspendedAt                   *time.Time    `json:"suspendedAt,omitempty"`
	SuspendedBy                   *string       `json:"suspendedBy,omitempty"`
	BlockReason                   *string       `json:"blockReason,omitempty"`
	BlockedAt                     *time.Time    `json:"blockedAt,omitempty"`
	BlockedBy                     *string       `json:"blockedBy,omitempty"`
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
	ShowAvatar         bool               `json:"showAvatar"`
	ShowBio            bool               `json:"showBio"`
	ShowLocation       bool               `json:"showLocation"`
	ShowSkills         bool               `json:"showSkills"`
	ShowLanguages      bool               `json:"showLanguages"`
	ShowAvailability   bool               `json:"showAvailabilityStatus"`
	ShowWebsiteURL     bool               `json:"showWebsiteUrl"`
	ShowLinkedInURL    bool               `json:"showLinkedinUrl"`
	ShowGitHubURL      bool               `json:"showGithubUrl"`
	ShowTwitterURL     bool               `json:"showTwitterUrl"`
	CreatedAt          time.Time          `json:"createdAt"`
	UpdatedAt          time.Time          `json:"updatedAt"`
	DeletedAt          *time.Time         `json:"deletedAt,omitempty"`
}

type PublicProfileSkill struct {
	ID         string    `json:"id"`
	Slug       string    `json:"slug"`
	Title      string    `json:"title"`
	Summary    string    `json:"summary"`
	SkillType  string    `json:"skillType"`
	Difficulty string    `json:"difficulty"`
	AvgRating  float64   `json:"avgRating"`
	RunsCount  int       `json:"runsCount"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type PublicProfile struct {
	UserID             string               `json:"userId"`
	FirstName          string               `json:"firstName"`
	LastName           string               `json:"lastName"`
	FullName           string               `json:"fullName"`
	Username           *string              `json:"username"`
	AvatarURL          *string              `json:"avatarUrl"`
	Bio                *string              `json:"bio"`
	Location           *string              `json:"location"`
	Skills             []string             `json:"skills"`
	Languages          []string             `json:"languages"`
	WebsiteURL         *string              `json:"websiteUrl"`
	LinkedInURL        *string              `json:"linkedinUrl"`
	GitHubURL          *string              `json:"githubUrl"`
	TwitterURL         *string              `json:"twitterUrl"`
	AvailabilityStatus *AvailabilityStatus  `json:"availabilityStatus"`
	PublicSkills       []PublicProfileSkill `json:"publicSkills"`
}

type Skill struct {
	ID               string          `json:"id"`
	Slug             string          `json:"slug"`
	Title            string          `json:"title"`
	Summary          string          `json:"summary"`
	Description      string          `json:"description"`
	Provider         string          `json:"provider"`
	Entrypoint       string          `json:"entrypoint"`
	Config           map[string]any  `json:"config"`
	Status           SkillStatus     `json:"status"`
	Visibility       Visibility      `json:"visibility"`
	PriceCents       int             `json:"priceCents"`
	Currency         string          `json:"currency"`
	AccessType       SkillAccessType `json:"accessType"`
	IsFeatured       bool            `json:"isFeatured"`
	IsVerified       bool            `json:"isVerified"`
	ModerationReason *string         `json:"moderationReason,omitempty"`
	ModeratedBy      *string         `json:"moderatedBy,omitempty"`
	ModeratedAt      *time.Time      `json:"moderatedAt,omitempty"`
	CreatedBy        string          `json:"createdBy"`
	Creator          Creator         `json:"creator"`
	Modules          []Module        `json:"modules,omitempty"`
	ModulesCount     int             `json:"modulesCount"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
	// marketplace fields (migration 022)
	CategoryID           *string       `json:"categoryId,omitempty"`
	SubcategoryID        *string       `json:"subcategoryId,omitempty"`
	SkillType            string        `json:"skillType"`
	Difficulty           string        `json:"difficulty"`
	CoverURL             *string       `json:"coverUrl,omitempty"`
	IconURL              *string       `json:"iconUrl,omitempty"`
	IsAgentReady         bool          `json:"isAgentReady"`
	RecommendedAgentID   *string       `json:"recommendedAgentId,omitempty"`
	EstimatedTimeMinutes *int          `json:"estimatedTimeMinutes,omitempty"`
	LanguageCode         string        `json:"languageCode"`
	SuccessRate          float64       `json:"successRate"`
	AvgRating            float64       `json:"avgRating"`
	RunsCount            int           `json:"runsCount"`
	SavesCount           int           `json:"savesCount"`
	ForksCount           int           `json:"forksCount"`
	PublishedAt          *time.Time    `json:"publishedAt,omitempty"`
	Tags                 []Tag         `json:"tags,omitempty"`
	CurrentVersion       *SkillVersion `json:"currentVersion,omitempty"`
}

type SkillVersion struct {
	ID                 string         `json:"id"`
	SkillID            string         `json:"skillId"`
	VersionNo          int            `json:"versionNo"`
	IsCurrent          bool           `json:"isCurrent"`
	Changelog          *string        `json:"changelog,omitempty"`
	ContentMD          string         `json:"contentMd"`
	PromptTemplate     *string        `json:"promptTemplate,omitempty"`
	SystemInstructions *string        `json:"systemInstructions,omitempty"`
	InputSchema        map[string]any `json:"inputSchema,omitempty"`
	OutputSchema       map[string]any `json:"outputSchema,omitempty"`
	ErrorSchema        map[string]any `json:"errorSchema,omitempty"`
	CreatedBy          string         `json:"createdBy"`
	CreatedAt          time.Time      `json:"createdAt"`
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

// SkillAccess records explicit access grants for paid / invite-only skills.
type SkillAccess struct {
	ID         string          `json:"id"`
	SkillID    string          `json:"skillId"`
	UserID     string          `json:"userId"`
	AccessType SkillAccessType `json:"accessType"`
	GrantedBy  *string         `json:"grantedBy,omitempty"`
	ExpiresAt  *time.Time      `json:"expiresAt,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
}

// AdminAuditLog records every governance action taken by an admin.
type AdminAuditLog struct {
	ID           string    `json:"id"`
	EntityType   string    `json:"entityType"`
	EntityID     string    `json:"entityId"`
	Action       string    `json:"action"`
	OldValueJSON *string   `json:"oldValueJson,omitempty"`
	NewValueJSON *string   `json:"newValueJson,omitempty"`
	Reason       *string   `json:"reason,omitempty"`
	ActorID      string    `json:"actorId"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ── Catalog ──────────────────────────────────────────────────────────────────

type Category struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Icon        *string   `json:"icon,omitempty"`
	SortOrder   int       `json:"sortOrder"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Subcategory struct {
	ID         string    `json:"id"`
	CategoryID string    `json:"categoryId"`
	Slug       string    `json:"slug"`
	Name       string    `json:"name"`
	SortOrder  int       `json:"sortOrder"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Tag struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`
	CreatedAt time.Time `json:"createdAt"`
}

// ── Agents ───────────────────────────────────────────────────────────────────

type Agent struct {
	ID                 string         `json:"id"`
	Slug               string         `json:"slug"`
	Name               string         `json:"name"`
	Provider           string         `json:"provider"`
	ModelName          string         `json:"modelName"`
	Description        *string        `json:"description,omitempty"`
	SupportsText       bool           `json:"supportsText"`
	SupportsCode       bool           `json:"supportsCode"`
	SupportsTools      bool           `json:"supportsTools"`
	SupportsWeb        bool           `json:"supportsWeb"`
	SupportsFiles      bool           `json:"supportsFiles"`
	SupportsMultimodal bool           `json:"supportsMultimodal"`
	SupportsJSONOutput bool           `json:"supportsJsonOutput"`
	MaxContextTokens   *int           `json:"maxContextTokens,omitempty"`
	PricingJSON        map[string]any `json:"pricing,omitempty"`
	Status             string         `json:"status"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
}

type SkillAgentCompatibility struct {
	SkillID            string  `json:"skillId"`
	AgentID            string  `json:"agentId"`
	Agent              Agent   `json:"agent"`
	CompatibilityScore float64 `json:"compatibilityScore"`
	SuccessRate        float64 `json:"successRate"`
	AvgRating          float64 `json:"avgRating"`
	RunsCount          int     `json:"runsCount"`
	TestedBySystem     bool    `json:"testedBySystem"`
	TestedByUsers      bool    `json:"testedByUsers"`
	Notes              *string `json:"notes,omitempty"`
}

// ── Social ───────────────────────────────────────────────────────────────────

type SkillReview struct {
	ID         string    `json:"id"`
	SkillID    string    `json:"skillId"`
	UserID     string    `json:"userId"`
	RunID      *string   `json:"runId,omitempty"`
	Rating     int       `json:"rating"`
	ReviewText *string   `json:"reviewText,omitempty"`
	Reviewer   Creator   `json:"reviewer"`
	CreatedAt  time.Time `json:"createdAt"`
}

type SavedSkill struct {
	UserID    string    `json:"userId"`
	SkillID   string    `json:"skillId"`
	CreatedAt time.Time `json:"createdAt"`
}

type Collection struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"ownerId"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Visibility  string    `json:"visibility"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CollectionItem struct {
	CollectionID string    `json:"collectionId"`
	SkillID      string    `json:"skillId"`
	SortOrder    int       `json:"sortOrder"`
	AddedAt      time.Time `json:"addedAt"`
}

type RunFeedback struct {
	RunID           string  `json:"runId"`
	UserID          string  `json:"userId"`
	Rating          int     `json:"rating"`
	UsefulnessScore *int    `json:"usefulnessScore,omitempty"`
	WouldUseAgain   *bool   `json:"wouldUseAgain,omitempty"`
	Comment         *string `json:"comment,omitempty"`
}

// ── Skill filter ─────────────────────────────────────────────────────────────

type SkillFilter struct {
	Query         string
	CategoryID    string
	SubcategoryID string
	TagSlugs      []string
	Difficulty    string
	SkillType     string
	IsAgentReady  *bool
	Sort          string // popular | new | rating | runs
}

// ── AdminStats is the top-level admin dashboard snapshot.
type AdminStats struct {
	Users  AdminUserStats  `json:"users"`
	Skills AdminSkillStats `json:"skills"`
}

type AdminUserStats struct {
	Total      int64 `json:"total"`
	Active     int64 `json:"active"`
	Suspended  int64 `json:"suspended"`
	Blocked    int64 `json:"blocked"`
	NewLast7d  int64 `json:"new7d"`
	NewLast30d int64 `json:"new30d"`
}

type AdminSkillStats struct {
	Total         int64 `json:"total"`
	Published     int64 `json:"published"`
	PendingReview int64 `json:"pendingReview"`
	Hidden        int64 `json:"hidden"`
	Free          int64 `json:"free"`
	Paid          int64 `json:"paid"`
	NewLast7d     int64 `json:"new7d"`
}
