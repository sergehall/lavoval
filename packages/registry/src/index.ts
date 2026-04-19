export type SkillStatus = 'draft' | 'published' | 'archived';
export type SkillVisibility = 'public' | 'private';
export type SkillProvider = 'internal';

export const defaultSkillProvider: SkillProvider = 'internal';
export const suggestedSkillEntrypoints = ['echo', 'text-summary-mock', 'keyword-extract-mock'] as const;

export type SkillModule = {
  id: string;
  skillId: string;
  slug: string;
  title: string;
  summary: string;
  content: string;
  position: number;
  status: SkillStatus;
  createdAt: string;
  updatedAt: string;
};

export type SkillCreator = {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
};

export type SkillSummary = {
  id: string;
  slug: string;
  title: string;
  summary: string;
  provider: string;
  entrypoint: string;
  config: Record<string, unknown>;
  status: SkillStatus;
  visibility: SkillVisibility;
  creator: SkillCreator;
  createdAt: string;
  updatedAt: string;
  modulesCount: number;
};

export type SkillSchema = Record<string, unknown>;

export type SkillCurrentVersion = {
  id: string;
  skillId: string;
  versionNo: number;
  isCurrent: boolean;
  changelog?: string | null;
  contentMd: string;
  promptTemplate?: string | null;
  systemInstructions?: string | null;
  inputSchema?: SkillSchema | null;
  outputSchema?: SkillSchema | null;
  errorSchema?: SkillSchema | null;
  createdBy: string;
  createdAt: string;
};

export type SkillDetail = SkillSummary & {
  description: string;
  modules: SkillModule[];
  currentVersion?: SkillCurrentVersion | null;
};

export type SkillMutationRequest = {
  slug: string;
  title: string;
  summary: string;
  description: string;
  provider: string;
  entrypoint: string;
  config: Record<string, unknown>;
  status: SkillStatus;
  visibility: SkillVisibility;
  inputSchema?: SkillSchema;
  outputSchema?: SkillSchema;
  errorSchema?: SkillSchema;
  promptTemplate?: string;
  systemInstructions?: string;
  changelog?: string;
};

export function parseSkillConfig(raw: FormDataEntryValue | string | null | undefined) {
  const value = String(raw ?? '').trim();
  if (!value) {
    return {};
  }

  return JSON.parse(value) as Record<string, unknown>;
}

export function formatSkillConfig(config: Record<string, unknown> | null | undefined) {
  return JSON.stringify(config ?? {}, null, 2);
}

export function buildSkillSearchText(skill: Pick<SkillSummary, 'title' | 'slug' | 'summary' | 'creator'>) {
  return [
    skill.title,
    skill.slug,
    skill.summary,
    skill.creator.firstName,
    skill.creator.lastName,
    skill.creator.email,
  ]
    .join(' ')
    .toLowerCase();
}
