import { z } from 'zod';
export {
  defaultSkillProvider,
  formatSkillConfig,
  parseSkillConfig,
  suggestedSkillEntrypoints,
} from '@lavoval/registry';
export type {
  SkillCreator,
  SkillDetail,
  SkillModule,
  SkillMutationRequest,
  SkillProvider,
  SkillSummary,
  SkillVisibility,
} from '@lavoval/registry';

export {
  runtimeRunRequestSchema,
  skillRunMetaSchema,
  skillRunSchema,
  skillRunSkillSchema,
  skillRunStatusSchema,
} from './runtime';
export type {
  RuntimeRunRequest,
  SkillRun,
  SkillRunMeta,
  SkillRunSkill,
  SkillRunStatus,
} from './runtime';

export const roleSchema = z.enum(['user', 'admin', 'root_owner']);
export type Role = z.infer<typeof roleSchema>;

export const accountStatusSchema = z.enum(['active', 'invited', 'suspended', 'blocked']);
export type AccountStatus = z.infer<typeof accountStatusSchema>;

// Full status enum — used for display / admin governance responses.
export const skillStatusSchema = z.enum([
  'draft',
  'pending_review',
  'published',
  'hidden',
  'archived',
  'rejected',
]);
export type SkillStatus = z.infer<typeof skillStatusSchema>;

// Author-writable statuses — used in skill create/edit mutation forms.
export const authorSkillStatusSchema = z.enum(['draft', 'published', 'archived']);
export type AuthorSkillStatus = z.infer<typeof authorSkillStatusSchema>;

export const skillAccessTypeSchema = z.enum(['free', 'paid', 'invite_only']);
export type SkillAccessType = z.infer<typeof skillAccessTypeSchema>;

export const enrollmentStatusSchema = z.enum(['assigned', 'in_progress', 'completed']);
export type EnrollmentStatus = z.infer<typeof enrollmentStatusSchema>;

export const sessionUserSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  role: roleSchema,
  firstName: z.string().optional(),
  lastName: z.string().optional(),
  avatarUrl: z.string().nullable().optional(),
});
export type SessionUser = z.infer<typeof sessionUserSchema>;

export const authResponseSchema = z.object({
  accessToken: z.string(),
  refreshToken: z.string(),
  user: sessionUserSchema
});
export type AuthResponse = z.infer<typeof authResponseSchema>;

export const registerResponseSchema = z.object({
  email: z.string().email(),
  verificationRequired: z.boolean()
});
export type RegisterResponse = z.infer<typeof registerResponseSchema>;

export const verificationResponseSchema = z.object({
  email: z.string().email(),
  alreadyVerified: z.boolean()
});
export type VerificationResponse = z.infer<typeof verificationResponseSchema>;

export const availabilityStatusSchema = z.enum(['open', 'limited', 'closed']);
export type AvailabilityStatus = z.infer<typeof availabilityStatusSchema>;

const httpUrl = z.string().url().max(2048).refine(
  (v) => /^https?:\/\//i.test(v),
  { message: 'URL must start with http:// or https://' }
);

export const profileSchema = z.object({
  userId: z.string().uuid(),
  role: roleSchema,
  firstName: z.string().min(1),
  lastName: z.string().min(1),
  bio: z.string().max(500).nullable(),
  timezone: z.string().default('UTC'),
  username: z.string().nullable(),
  avatarUrl: z.string().nullable(),
  location: z.string().nullable(),
  skills: z.array(z.string()).nullable(),
  languages: z.array(z.string()).nullable(),
  websiteUrl: z.string().nullable(),
  linkedinUrl: z.string().nullable(),
  githubUrl: z.string().nullable(),
  twitterUrl: z.string().nullable(),
  availabilityStatus: availabilityStatusSchema.default('open'),
  isPublicProfile: z.boolean().default(true),
  showAvatar: z.boolean().default(true),
  showBio: z.boolean().default(true),
  showLocation: z.boolean().default(true),
  showSkills: z.boolean().default(true),
  showLanguages: z.boolean().default(true),
  showAvailabilityStatus: z.boolean().default(true),
  showWebsiteUrl: z.boolean().default(true),
  showLinkedinUrl: z.boolean().default(true),
  showGithubUrl: z.boolean().default(true),
  showTwitterUrl: z.boolean().default(true),
});
export type Profile = z.infer<typeof profileSchema>;

export const publicProfileSkillSchema = z.object({
  id: z.string().uuid(),
  slug: z.string().min(2),
  title: z.string().min(2),
  summary: z.string().min(2),
  skillType: z.string(),
  difficulty: z.string(),
  avgRating: z.number().default(0),
  runsCount: z.number().int().default(0),
  updatedAt: z.string(),
});
export type PublicProfileSkill = z.infer<typeof publicProfileSkillSchema>;

export const publicProfileSchema = z.object({
  userId: z.string().uuid(),
  firstName: z.string().min(1),
  lastName: z.string().min(1),
  fullName: z.string().min(1),
  username: z.string().nullable(),
  avatarUrl: z.string().nullable(),
  bio: z.string().nullable(),
  location: z.string().nullable(),
  skills: z.array(z.string()).nullable(),
  languages: z.array(z.string()).nullable(),
  websiteUrl: z.string().nullable(),
  linkedinUrl: z.string().nullable(),
  githubUrl: z.string().nullable(),
  twitterUrl: z.string().nullable(),
  availabilityStatus: availabilityStatusSchema.nullable(),
  publicSkills: z.array(publicProfileSkillSchema),
});
export type PublicProfile = z.infer<typeof publicProfileSchema>;

export const skillModuleSchema = z.object({
  id: z.string().uuid(),
  skillId: z.string().uuid(),
  slug: z.string().min(2),
  title: z.string().min(2),
  summary: z.string().min(2),
  content: z.string().min(2),
  position: z.number().int().nonnegative(),
  status: skillStatusSchema,
  createdAt: z.string(),
  updatedAt: z.string()
});

export const skillJsonSchemaSchema = z.record(z.string(), z.unknown());

export const skillCurrentVersionSchema = z.object({
  id: z.string().uuid(),
  skillId: z.string().uuid(),
  versionNo: z.number().int().positive(),
  isCurrent: z.boolean(),
  changelog: z.string().nullable().optional(),
  contentMd: z.string(),
  promptTemplate: z.string().nullable().optional(),
  systemInstructions: z.string().nullable().optional(),
  inputSchema: skillJsonSchemaSchema.nullable().optional(),
  outputSchema: skillJsonSchemaSchema.nullable().optional(),
  errorSchema: skillJsonSchemaSchema.nullable().optional(),
  createdBy: z.string().uuid(),
  createdAt: z.string(),
});

export const skillCreatorSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  firstName: z.string().min(1),
  lastName: z.string().min(1)
});

export const skillSummarySchema = z.object({
  id: z.string().uuid(),
  slug: z.string().min(2),
  title: z.string().min(2),
  summary: z.string().min(2),
  provider: z.string().min(2),
  entrypoint: z.string().min(2),
  config: z.record(z.string(), z.unknown()),
  status: skillStatusSchema,
  visibility: z.enum(['public', 'private']),
  creator: skillCreatorSchema,
  createdAt: z.string(),
  updatedAt: z.string(),
  modulesCount: z.number().int().nonnegative(),
  // marketplace fields
  categoryId: z.string().uuid().nullable().optional(),
  subcategoryId: z.string().uuid().nullable().optional(),
  skillType: z.string().default('workflow'),
  difficulty: z.string().default('middle'),
  coverUrl: z.string().nullable().optional(),
  iconUrl: z.string().nullable().optional(),
  isAgentReady: z.boolean().default(false),
  recommendedAgentId: z.string().uuid().nullable().optional(),
  estimatedTimeMinutes: z.number().int().nullable().optional(),
  languageCode: z.string().default('en'),
  successRate: z.number().default(0),
  avgRating: z.number().default(0),
  runsCount: z.number().int().default(0),
  savesCount: z.number().int().default(0),
  forksCount: z.number().int().default(0),
  publishedAt: z.string().nullable().optional(),
  tags: z.array(z.object({
    id: z.string().uuid(),
    slug: z.string(),
    name: z.string(),
    kind: z.string(),
    createdAt: z.string(),
  })).optional(),
});

export const skillDetailSchema = skillSummarySchema.extend({
  description: z.string(),
  modules: z.array(skillModuleSchema),
  currentVersion: skillCurrentVersionSchema.nullable().optional(),
});

export const skillMutationSchema = z.object({
  slug: z.string().min(2),
  title: z.string().min(3),
  summary: z.string().min(10),
  description: z.string().min(20),
  provider: z.string().min(2),
  entrypoint: z.string().min(2),
  config: z.record(z.string(), z.unknown()),
  status: authorSkillStatusSchema,
  visibility: z.enum(['public', 'private']),
  // marketplace fields
  categoryId: z.string().uuid().nullable().optional(),
  subcategoryId: z.string().uuid().nullable().optional(),
  skillType: z.string().optional(),
  difficulty: z.string().optional(),
  coverUrl: z.string().nullable().optional(),
  isAgentReady: z.boolean().optional(),
  estimatedTimeMinutes: z.number().int().nullable().optional(),
  languageCode: z.string().optional(),
  tagIds: z.array(z.string().uuid()).optional(),
  inputSchema: skillJsonSchemaSchema.optional(),
  outputSchema: skillJsonSchemaSchema.optional(),
  errorSchema: skillJsonSchemaSchema.optional(),
  promptTemplate: z.string().optional(),
  systemInstructions: z.string().optional(),
  changelog: z.string().optional(),
});

export const moduleMutationSchema = z.object({
  slug: z.string().min(2).max(100),
  title: z.string().min(2).max(200),
  summary: z.string().min(2).max(500),
  content: z.string().min(10),
  position: z.number().int().min(0).optional(),
  status: skillStatusSchema,
});
export type ModuleMutationRequest = z.infer<typeof moduleMutationSchema>;

export const loginRequestSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8)
});
export type LoginRequest = z.infer<typeof loginRequestSchema>;

export const registerRequestSchema = z.object({
  email: z.string().email(),
  password: z.string().min(12),
  firstName: z.string().min(2),
  lastName: z.string().min(2)
});
export type RegisterRequest = z.infer<typeof registerRequestSchema>;

export const verifyEmailRequestSchema = z.object({
  token: z.string().min(24)
});
export type VerifyEmailRequest = z.infer<typeof verifyEmailRequestSchema>;

export const resendVerificationRequestSchema = z.object({
  email: z.string().email()
});
export type ResendVerificationRequest = z.infer<typeof resendVerificationRequestSchema>;

export const forgotPasswordRequestSchema = z.object({
  email: z.string().email()
});
export type ForgotPasswordRequest = z.infer<typeof forgotPasswordRequestSchema>;

export const forgotPasswordResponseSchema = z.object({
  email: z.string().email(),
  sent: z.boolean()
});
export type ForgotPasswordResponse = z.infer<typeof forgotPasswordResponseSchema>;

export const resetPasswordRequestSchema = z.object({
  token: z.string().min(24),
  newPassword: z.string().min(12)
});
export type ResetPasswordRequest = z.infer<typeof resetPasswordRequestSchema>;

export const resetPasswordResponseSchema = z.object({
  email: z.string().email(),
  reset: z.boolean()
});
export type ResetPasswordResponse = z.infer<typeof resetPasswordResponseSchema>;

export const mfaStatusResponseSchema = z.object({
  enabled: z.boolean(),
  pendingEnrollment: z.boolean(),
  enrolledAt: z.string().nullable().optional(),
  recoveryCodes: z.array(z.string()).optional(),
});
export type MFAStatusResponse = z.infer<typeof mfaStatusResponseSchema>;

export const mfaEnrollResponseSchema = z.object({
  secret: z.string().min(16),
  provisionUrl: z.string().min(1),
});
export type MFAEnrollResponse = z.infer<typeof mfaEnrollResponseSchema>;

export const oauthProviderSchema = z.enum(['google', 'github']);
export type OAuthProvider = z.infer<typeof oauthProviderSchema>;

export const accountSecurityProviderSchema = z.object({
  provider: oauthProviderSchema,
  connected: z.boolean(),
  connectedAt: z.string(),
});
export type AccountSecurityProvider = z.infer<typeof accountSecurityProviderSchema>;

export const accountSecuritySummarySchema = z.object({
  hasPassword: z.boolean(),
  passwordUpdatedAt: z.string().nullable().optional(),
  providers: z.array(accountSecurityProviderSchema),
  mfaEnabled: z.boolean(),
  mfaEnrolledAt: z.string().nullable().optional(),
});
export type AccountSecuritySummary = z.infer<typeof accountSecuritySummarySchema>;

export const enrollmentDetailSchema = z.object({
  id: z.string().uuid(),
  userId: z.string().uuid(),
  skillId: z.string().uuid(),
  status: enrollmentStatusSchema,
  progressPercent: z.number().int().min(0).max(100),
  assignedAt: z.string(),
  completedAt: z.string().nullable().optional(),
  userEmail: z.string().email(),
  skillTitle: z.string().min(1),
  skillSlug: z.string().min(1),
});
export type EnrollmentDetail = z.infer<typeof enrollmentDetailSchema>;

export const enrollmentAssignSchema = z.object({
  userId: z.string().uuid(),
  skillId: z.string().uuid(),
});
export type EnrollmentAssignRequest = z.infer<typeof enrollmentAssignSchema>;

export const enrollmentUpdateSchema = z.object({
  status: enrollmentStatusSchema,
  progressPercent: z.number().int().min(0).max(100),
});
export type EnrollmentUpdateRequest = z.infer<typeof enrollmentUpdateSchema>;

export const mfaVerifyEnrollmentRequestSchema = z.object({
  code: z.string().length(6),
});
export type MFAVerifyEnrollmentRequest = z.infer<typeof mfaVerifyEnrollmentRequestSchema>;

export const mfaDisableRequestSchema = z.object({
  password: z.string().min(8),
  code: z.string().length(6),
});
export type MFADisableRequest = z.infer<typeof mfaDisableRequestSchema>;

export const mfaRegenerateRecoveryCodesRequestSchema = z.object({
  password: z.string().min(8),
  code: z.string().length(6),
});
export type MFARegenerateRecoveryCodesRequest = z.infer<typeof mfaRegenerateRecoveryCodesRequestSchema>;

export const mfaCompleteSignInRequestSchema = z.object({
  challengeId: z.string().uuid(),
  code: z.string().length(6).optional(),
  recoveryCode: z.string().min(8).optional(),
}).refine((value) => Boolean(value.code || value.recoveryCode), {
  message: 'Provide an authenticator code or a recovery code.',
});
export type MFACompleteSignInRequest = z.infer<typeof mfaCompleteSignInRequestSchema>;

export const googleOAuthCompleteRequestSchema = z.object({
  code: z.string().min(8),
  state: z.string().min(16),
});
export type GoogleOAuthCompleteRequest = z.infer<typeof googleOAuthCompleteRequestSchema>;

export const githubOAuthCompleteRequestSchema = z.object({
  code: z.string().min(8),
  state: z.string().min(16),
});
export type GitHubOAuthCompleteRequest = z.infer<typeof githubOAuthCompleteRequestSchema>;

export const profileUpdateSchema = z.object({
  firstName: z.string().min(2).max(100),
  lastName: z.string().min(2).max(100),
  bio: z.string().max(500).nullable(),
  timezone: z.string().min(2).max(100),
  username: z
    .string()
    .min(3)
    .max(30)
    .regex(/^[a-zA-Z0-9_-]+$/, 'Only letters, digits, - and _ are allowed')
    .nullable()
    .optional(),
  avatarUrl: httpUrl.nullable().optional(),
  websiteUrl: httpUrl.nullable().optional(),
  linkedinUrl: httpUrl
    .refine((v) => /linkedin\.com\//i.test(v), { message: 'Must be a LinkedIn URL' })
    .nullable()
    .optional(),
  githubUrl: httpUrl
    .refine((v) => /github\.com\//i.test(v), { message: 'Must be a GitHub URL' })
    .nullable()
    .optional(),
  twitterUrl: httpUrl
    .refine((v) => /(twitter\.com|x\.com)\//i.test(v), { message: 'Must be a Twitter / X URL' })
    .nullable()
    .optional(),
  location: z.string().min(2).max(100).nullable().optional(),
  skills: z
    .array(z.string().min(1).max(50))
    .max(20)
    .nullable()
    .optional(),
  languages: z
    .array(z.string().min(2).max(10))
    .max(10)
    .nullable()
    .optional(),
  availabilityStatus: availabilityStatusSchema.optional(),
  isPublicProfile: z.boolean().optional(),
  showAvatar: z.boolean().optional(),
  showBio: z.boolean().optional(),
  showLocation: z.boolean().optional(),
  showSkills: z.boolean().optional(),
  showLanguages: z.boolean().optional(),
  showAvailabilityStatus: z.boolean().optional(),
  showWebsiteUrl: z.boolean().optional(),
  showLinkedinUrl: z.boolean().optional(),
  showGithubUrl: z.boolean().optional(),
  showTwitterUrl: z.boolean().optional(),
});
export type ProfileUpdateRequest = z.infer<typeof profileUpdateSchema>;

export const adminUserUpdateSchema = z.object({
  role: roleSchema,
  status: accountStatusSchema,
  reason: z.string().min(3).max(500).optional(),
}).superRefine((value, ctx) => {
  if ((value.status === 'suspended' || value.status === 'blocked') && !value.reason) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ['reason'],
      message: 'Reason is required when suspending or blocking a user.',
    });
  }
});
export type AdminUserUpdateRequest = z.infer<typeof adminUserUpdateSchema>;

export const adminUserStatusUpdateSchema = z.object({
  status: accountStatusSchema,
  reason: z.string().min(3).max(500).optional(),
}).superRefine((value, ctx) => {
  if ((value.status === 'suspended' || value.status === 'blocked') && !value.reason) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ['reason'],
      message: 'Reason is required when suspending or blocking a user.',
    });
  }
});
export type AdminUserStatusUpdateRequest = z.infer<typeof adminUserStatusUpdateSchema>;

export const adminUserRoleUpdateSchema = z.object({
  role: roleSchema,
  reason: z.string().min(3).max(500),
});
export type AdminUserRoleUpdateRequest = z.infer<typeof adminUserRoleUpdateSchema>;

export const adminSkillGovernanceSchema = z.object({
  status: skillStatusSchema,
  reason: z.string().min(3).max(500).optional(),
  featured: z.boolean().optional(),
  verified: z.boolean().optional(),
}).superRefine((value, ctx) => {
  if ((value.status === 'hidden' || value.status === 'rejected') && !value.reason) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ['reason'],
      message: 'Reason is required when hiding or rejecting a skill.',
    });
  }
});
export type AdminSkillGovernanceRequest = z.infer<typeof adminSkillGovernanceSchema>;

export const adminSkillPricingSchema = z.object({
  priceCents: z.number().int().min(0),
  currency: z.string().trim().length(3).transform((value) => value.toUpperCase()).default('USD'),
  accessType: skillAccessTypeSchema,
}).superRefine((value, ctx) => {
  if (value.accessType === 'paid' && value.priceCents <= 0) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ['priceCents'],
      message: 'Paid skills must have a price greater than zero.',
    });
  }
});
export type AdminSkillPricingRequest = z.infer<typeof adminSkillPricingSchema>;

export const adminAuditLogSchema = z.object({
  id: z.string().uuid(),
  entityType: z.string(),
  entityId: z.string(),
  action: z.string(),
  oldValueJson: z.record(z.string(), z.unknown()).nullable().optional(),
  newValueJson: z.record(z.string(), z.unknown()).nullable().optional(),
  reason: z.string().nullable().optional(),
  actorId: z.string().uuid().nullable().optional(),
  createdAt: z.string(),
});
export type AdminAuditLog = z.infer<typeof adminAuditLogSchema>;

export const adminUserStatsSchema = z.object({
  total: z.number().int(),
  active: z.number().int(),
  suspended: z.number().int(),
  blocked: z.number().int(),
  new7d: z.number().int(),
  new30d: z.number().int(),
});
export type AdminUserStats = z.infer<typeof adminUserStatsSchema>;

export const adminSkillStatsSchema = z.object({
  total: z.number().int(),
  published: z.number().int(),
  pendingReview: z.number().int(),
  hidden: z.number().int(),
  free: z.number().int(),
  paid: z.number().int(),
  new7d: z.number().int(),
});
export type AdminSkillStats = z.infer<typeof adminSkillStatsSchema>;

export const adminStatsSchema = z.object({
  users: adminUserStatsSchema,
  skills: adminSkillStatsSchema,
});
export type AdminStats = z.infer<typeof adminStatsSchema>;

// ── Marketplace enums ─────────────────────────────────────────────────────────

export const skillTypeSchema = z.enum(['guide', 'workflow', 'prompt-pack', 'agent-ready', 'service']);
export type SkillType = z.infer<typeof skillTypeSchema>;

export const difficultySchema = z.enum(['junior', 'middle', 'senior']);
export type Difficulty = z.infer<typeof difficultySchema>;

// ── Catalog: categories, subcategories, tags ──────────────────────────────────

export const categorySchema = z.object({
  id: z.string().uuid(),
  slug: z.string().min(1),
  name: z.string().min(1),
  description: z.string().nullable().optional(),
  icon: z.string().nullable().optional(),
  sortOrder: z.number().int(),
  createdAt: z.string(),
});
export type Category = z.infer<typeof categorySchema>;

export const subcategorySchema = z.object({
  id: z.string().uuid(),
  categoryId: z.string().uuid(),
  slug: z.string().min(1),
  name: z.string().min(1),
  sortOrder: z.number().int(),
  createdAt: z.string(),
});
export type Subcategory = z.infer<typeof subcategorySchema>;

export const tagSchema = z.object({
  id: z.string().uuid(),
  slug: z.string().min(1),
  name: z.string().min(1),
  kind: z.string(),
  createdAt: z.string(),
});
export type Tag = z.infer<typeof tagSchema>;

// ── Agents ────────────────────────────────────────────────────────────────────

export const agentSchema = z.object({
  id: z.string().uuid(),
  slug: z.string().min(1),
  name: z.string().min(1),
  provider: z.string().min(1),
  modelName: z.string().min(1),
  description: z.string().nullable().optional(),
  supportsText: z.boolean(),
  supportsCode: z.boolean(),
  supportsTools: z.boolean(),
  supportsWeb: z.boolean(),
  supportsFiles: z.boolean(),
  supportsMultimodal: z.boolean(),
  supportsJsonOutput: z.boolean(),
  maxContextTokens: z.number().int().nullable().optional(),
  pricing: z.record(z.string(), z.unknown()).nullable().optional(),
  status: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
});
export type Agent = z.infer<typeof agentSchema>;

export const skillAgentCompatibilitySchema = z.object({
  skillId: z.string().uuid(),
  agentId: z.string().uuid(),
  agent: agentSchema,
  compatibilityScore: z.number(),
  successRate: z.number(),
  avgRating: z.number(),
  runsCount: z.number().int(),
  testedBySystem: z.boolean(),
  testedByUsers: z.boolean(),
  notes: z.string().nullable().optional(),
});
export type SkillAgentCompatibility = z.infer<typeof skillAgentCompatibilitySchema>;

// ── Social: reviews, collections ──────────────────────────────────────────────

export const skillReviewSchema = z.object({
  id: z.string().uuid(),
  skillId: z.string().uuid(),
  userId: z.string().uuid(),
  runId: z.string().uuid().nullable().optional(),
  rating: z.number().int().min(1).max(5),
  reviewText: z.string().nullable().optional(),
  reviewer: z.object({
    id: z.string().uuid(),
    email: z.string().email(),
    firstName: z.string(),
    lastName: z.string(),
  }),
  createdAt: z.string(),
});
export type SkillReview = z.infer<typeof skillReviewSchema>;

export const collectionSchema = z.object({
  id: z.string().uuid(),
  ownerId: z.string().uuid(),
  title: z.string().min(1),
  description: z.string().nullable().optional(),
  visibility: z.enum(['public', 'private']),
  createdAt: z.string(),
  updatedAt: z.string(),
});
export type Collection = z.infer<typeof collectionSchema>;

export const collectionInputSchema = z.object({
  title: z.string().min(1).max(200),
  description: z.string().max(1000).optional(),
  visibility: z.enum(['public', 'private']).default('public'),
});
export type CollectionInput = z.infer<typeof collectionInputSchema>;

export const createReviewSchema = z.object({
  rating: z.number().int().min(1).max(5),
  reviewText: z.string().max(2000).optional(),
  runId: z.string().uuid().optional(),
});
export type CreateReviewInput = z.infer<typeof createReviewSchema>;

export const runFeedbackSchema = z.object({
  rating: z.number().int().min(1).max(5),
  usefulnessScore: z.number().int().min(1).max(5).optional(),
  wouldUseAgain: z.boolean().optional(),
  comment: z.string().max(2000).optional(),
});
export type RunFeedbackInput = z.infer<typeof runFeedbackSchema>;

// ── Skill filter params ───────────────────────────────────────────────────────

export type SkillFilterParams = {
  q?: string;
  category?: string;
  subcategory?: string;
  tags?: string;
  difficulty?: Difficulty;
  skillType?: SkillType;
  agentReady?: boolean;
  sort?: 'popular' | 'new' | 'rating' | 'runs';
};

export const apiEnvelopeSchema = <T extends z.ZodTypeAny>(schema: T) => z.object({
  data: schema,
  meta: z.record(z.string(), z.unknown()).optional()
});
