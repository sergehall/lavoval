import { z } from 'zod';

export const roleSchema = z.enum(['user', 'admin']);
export type Role = z.infer<typeof roleSchema>;

export const accountStatusSchema = z.enum(['active', 'invited', 'suspended']);
export type AccountStatus = z.infer<typeof accountStatusSchema>;

export const skillStatusSchema = z.enum(['draft', 'published', 'archived']);
export type SkillStatus = z.infer<typeof skillStatusSchema>;

export const skillRunStatusSchema = z.enum(['queued', 'running', 'completed', 'failed']);
export type SkillRunStatus = z.infer<typeof skillRunStatusSchema>;

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

export const profileSchema = z.object({
  userId: z.string().uuid(),
  firstName: z.string().min(1),
  lastName: z.string().min(1),
  bio: z.string().max(500).nullable(),
  timezone: z.string().default('UTC')
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
export type SkillModule = z.infer<typeof skillModuleSchema>;

export const skillCreatorSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  firstName: z.string().min(1),
  lastName: z.string().min(1)
});
export type SkillCreator = z.infer<typeof skillCreatorSchema>;

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
export type SkillSummary = z.infer<typeof skillSummarySchema>;

export const skillDetailSchema = skillSummarySchema.extend({
  description: z.string(),
  modules: z.array(skillModuleSchema)
});
export type SkillDetail = z.infer<typeof skillDetailSchema>;

export const skillRunSchema = z.object({
  id: z.string().uuid(),
  skillId: z.string().uuid(),
  userId: z.string().uuid(),
  status: skillRunStatusSchema,
  input: z.record(z.string(), z.unknown()),
  output: z.record(z.string(), z.unknown()).optional(),
  errorMessage: z.string().nullable().optional(),
  startedAt: z.string().nullable().optional(),
  finishedAt: z.string().nullable().optional(),
  createdAt: z.string(),
});
export type SkillRun = z.infer<typeof skillRunSchema>;

export const runtimeRunRequestSchema = z.object({
  skillId: z.string().uuid(),
  input: z.record(z.string(), z.unknown()),
});
export type RuntimeRunRequest = z.infer<typeof runtimeRunRequestSchema>;

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

export const profileUpdateSchema = z.object({
  firstName: z.string().min(2),
  lastName: z.string().min(2),
  bio: z.string().max(500).nullable(),
  timezone: z.string().min(2)
});
export type ProfileUpdateRequest = z.infer<typeof profileUpdateSchema>;

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
export type SkillMutationRequest = z.infer<typeof skillMutationSchema>;

export const apiEnvelopeSchema = <T extends z.ZodTypeAny>(schema: T) => z.object({
  data: schema,
  meta: z.record(z.string(), z.unknown()).optional()
});
