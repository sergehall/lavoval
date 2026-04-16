import type {
  AdminAuditLog,
  AdminSkillStats,
  AdminStats,
  AdminUserStats,
  AccountStatus,
  AuthResponse,
  Profile,
  SkillDetail,
  SkillSummary,
  SessionUser,
} from '@lavoval/contracts';
import type { SkillRun } from '@lavoval/contracts/runtime';

export type ApiEnvelope<T> = {
  data: T;
  meta?: Record<string, unknown>;
};

export type AuthTokens = Pick<AuthResponse, 'accessToken' | 'refreshToken'>;
export type SessionState = {
  user: SessionUser;
  accessToken: string;
  refreshToken: string;
};

export type DashboardSummary = {
  profile: Profile;
  skills: SkillSummary[];
};

export type UsersListItem = SessionUser & {
  status: AccountStatus;
  createdAt: string;
  emailVerifiedAt?: string | null;
  suspensionReason?: string | null;
  suspendedAt?: string | null;
  suspendedBy?: string | null;
  blockReason?: string | null;
  blockedAt?: string | null;
  blockedBy?: string | null;
};

export type UserWithProfile = {
  user: UsersListItem;
  profile: import('@lavoval/contracts').Profile;
};

export type SkillPayload = SkillDetail;
export type SkillRunPayload = SkillRun;

export type MailJobStatus = 'queued' | 'retrying' | 'processing' | 'sent' | 'dead_letter';

export type MailOperationalSnapshot = {
  countsByStatus: Record<MailJobStatus, number>;
  deadLettersByErrorCode: Record<string, number>;
  oldestReadyAgeSeconds: number;
};

export type MailRetentionSnapshot = {
  jobsRetention: string;
  eventsRetention: string;
  cleanupBatchSize: number;
  cleanupInterval: string;
  cleanupDryRun: boolean;
  autoCleanupEnabled: boolean;
  jobsCutoff?: string | null;
  eventsCutoff?: string | null;
  eligibleJobs: number;
  eligibleEvents: number;
  jobsRetentionActive: boolean;
  eventsRetentionActive: boolean;
  latestCleanupRun?: MailCleanupRun | null;
  alerts: MailRetentionAlert[];
  webhookAlertingEnabled: boolean;
};

export type MailCleanupResult = {
  jobsDeleted: number;
  eventsDeleted: number;
  remainingJobs: number;
  remainingEvents: number;
  completedAt: string;
};

export type MailCleanupRun = {
  id: string;
  mode: string;
  status: string;
  dryRun: boolean;
  candidateJobs: number;
  candidateEvents: number;
  deletedJobs: number;
  deletedEvents: number;
  errorMessage?: string | null;
  durationMs: number;
  createdAt: string;
};

export type MailRetentionAlert = {
  severity: string;
  code: string;
  message: string;
};

export type AdminMailJobFilter = {
  query?: string;
  messageType?: string;
  provider?: string;
  errorCode?: string;
  limit?: number;
};

export type AdminMailEventFilter = {
  query?: string;
  jobId?: string;
  eventType?: string;
  messageType?: string;
  provider?: string;
  errorCode?: string;
  limit?: number;
};

export type MailJob = {
  id: string;
  messageType: string;
  recipientEmail: string;
  idempotencyKey?: string | null;
  status: MailJobStatus;
  attempts: number;
  maxAttempts: number;
  nextAttemptAt: string;
  leasedUntil?: string | null;
  lastError?: string | null;
  lastErrorCode?: string | null;
  provider?: string | null;
  providerMessageId?: string | null;
  sentAt?: string | null;
  deadLetteredAt?: string | null;
  createdAt: string;
  updatedAt: string;
};

export type MailEvent = {
  id: string;
  jobId: string;
  eventType: string;
  messageType: string;
  provider?: string | null;
  providerMessageId?: string | null;
  recipientEmail: string;
  errorCode?: string | null;
  attempt?: number | null;
  createdAt: string;
};

export type { AdminAuditLog, AdminUserStats, AdminSkillStats, AdminStats };

export type MailSuppressionKind = 'email' | 'domain';

export type MailSuppression = {
  id: string;
  kind: MailSuppressionKind;
  value: string;
  reason: string;
  createdAt: string;
};
