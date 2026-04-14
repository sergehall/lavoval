import { z } from 'zod';

const runtimeSkillCreatorSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  firstName: z.string().min(1),
  lastName: z.string().min(1),
});

export const skillRunStatusSchema = z.enum(['queued', 'running', 'completed', 'failed']);
export type SkillRunStatus = z.infer<typeof skillRunStatusSchema>;

export const skillRunSkillSchema = z.object({
  id: z.string().uuid(),
  slug: z.string().min(2),
  title: z.string().min(2),
  entrypoint: z.string().min(2),
  creator: runtimeSkillCreatorSchema,
});
export type SkillRunSkill = z.infer<typeof skillRunSkillSchema>;

export const skillRunMetaSchema = z.object({
  durationMs: z.number().int().nullable().optional(),
  hasOutput: z.boolean(),
  hasError: z.boolean(),
  inputKeysCount: z.number().int().nonnegative(),
  outputKeysCount: z.number().int().nonnegative(),
});
export type SkillRunMeta = z.infer<typeof skillRunMetaSchema>;

export const skillRunSchema = z.object({
  id: z.string().uuid(),
  skillId: z.string().uuid(),
  userId: z.string().uuid(),
  skill: skillRunSkillSchema,
  status: skillRunStatusSchema,
  input: z.record(z.string(), z.unknown()),
  output: z.record(z.string(), z.unknown()).optional(),
  meta: skillRunMetaSchema,
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
