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

export const roleSchema = z.enum(['user', 'admin']);
export type Role = z.infer<typeof roleSchema>;

export const accountStatusSchema = z.enum(['active', 'invited', 'suspended']);
export type AccountStatus = z.infer<typeof accountStatusSchema>;

export const skillStatusSchema = z.enum(['draft', 'published', 'archived']);
export type SkillStatus = z.infer<typeof skillStatusSchema>;

export const enrollmentStatusSchema = z.enum(['assigned', 'in_progress', 'completed']);
export type EnrollmentStatus = z.infer<typeof enrollmentStatusSchema>;

export const sessionUserSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  role: roleSchema,
  firstName: z.string().optional(),
  lastName: z.string().optional()
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
});
export type Profile = z.infer<typeof profileSchema>;

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
  modulesCount: z.number().int().nonnegative()
});

export const skillDetailSchema = skillSummarySchema.extend({
  description: z.string(),
  modules: z.array(skillModuleSchema)
});

export const skillMutationSchema = z.object({
  slug: z.string().min(2),
  title: z.string().min(3),
  summary: z.string().min(10),
  description: z.string().min(20),
  provider: z.string().min(2),
  entrypoint: z.string().min(2),
  config: z.record(z.string(), z.unknown()),
  status: skillStatusSchema,
  visibility: z.enum(['public', 'private'])
});

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
});
export type ProfileUpdateRequest = z.infer<typeof profileUpdateSchema>;

export const adminUserUpdateSchema = z.object({
  role: roleSchema,
  status: accountStatusSchema,
});
export type AdminUserUpdateRequest = z.infer<typeof adminUserUpdateSchema>;

export const apiEnvelopeSchema = <T extends z.ZodTypeAny>(schema: T) => z.object({
  data: schema,
  meta: z.record(z.string(), z.unknown()).optional()
});
