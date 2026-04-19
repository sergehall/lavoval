'use server';

import { revalidatePath } from 'next/cache';
import { skillMutationSchema } from '@lavoval/contracts';
import { parseSkillConfig } from '@lavoval/registry';
import {
  createMySkill,
  deleteMySkill,
  requireSession,
  updateMySkill,
} from '@/shared/api/server-client';

function parseJsonObjectField(formData: FormData, field: string) {
  const raw = String(formData.get(field) ?? '').trim();
  if (!raw) {
    return undefined;
  }

  return JSON.parse(raw) as Record<string, unknown>;
}

function parseMarketplaceFields(formData: FormData) {
  const tagIds = formData.getAll('tagIds').map(String).filter(Boolean);
  const estimatedRaw = formData.get('estimatedTimeMinutes');
  const estimatedTimeMinutes = estimatedRaw ? Number(estimatedRaw) : undefined;
  const isAgentReady = formData.get('isAgentReady') === 'true';

  return {
    categoryId: (formData.get('categoryId') as string) || undefined,
    skillType: (formData.get('skillType') as string) || undefined,
    difficulty: (formData.get('difficulty') as string) || undefined,
    languageCode: (formData.get('languageCode') as string) || undefined,
    isAgentReady,
    estimatedTimeMinutes:
      estimatedTimeMinutes && !isNaN(estimatedTimeMinutes) ? estimatedTimeMinutes : undefined,
    tagIds: tagIds.length > 0 ? tagIds : undefined,
    inputSchema: parseJsonObjectField(formData, 'inputSchema'),
    outputSchema: parseJsonObjectField(formData, 'outputSchema'),
    errorSchema: parseJsonObjectField(formData, 'errorSchema'),
    promptTemplate: String(formData.get('promptTemplate') ?? ''),
    systemInstructions: String(formData.get('systemInstructions') ?? ''),
    changelog: String(formData.get('changelog') ?? ''),
  };
}

export async function createOwnSkillAction(formData: FormData) {
  const session = await requireSession();
  const payload = skillMutationSchema.parse({
    slug: formData.get('slug'),
    title: formData.get('title'),
    summary: formData.get('summary'),
    description: formData.get('description'),
    provider: formData.get('provider'),
    entrypoint: formData.get('entrypoint'),
    config: parseSkillConfig(formData.get('config')),
    status: formData.get('status'),
    visibility: formData.get('visibility'),
    ...parseMarketplaceFields(formData),
  });

  await createMySkill(session.accessToken, payload);
  revalidatePath('/account');
  revalidatePath('/account/my-skills');
  revalidatePath('/skills');
}

export async function updateOwnSkillAction(skillID: string, formData: FormData) {
  const session = await requireSession();
  const payload = skillMutationSchema.parse({
    slug: formData.get('slug'),
    title: formData.get('title'),
    summary: formData.get('summary'),
    description: formData.get('description'),
    provider: formData.get('provider'),
    entrypoint: formData.get('entrypoint'),
    config: parseSkillConfig(formData.get('config')),
    status: formData.get('status'),
    visibility: formData.get('visibility'),
    ...parseMarketplaceFields(formData),
  });

  await updateMySkill(session.accessToken, skillID, payload);
  revalidatePath('/account');
  revalidatePath('/account/my-skills');
  revalidatePath(`/account/my-skills/${skillID}`);
  revalidatePath('/skills');
}

export async function deleteOwnSkillAction(skillID: string) {
  const session = await requireSession();
  await deleteMySkill(session.accessToken, skillID);
  revalidatePath('/account');
  revalidatePath('/account/my-skills');
  revalidatePath('/skills');
}
