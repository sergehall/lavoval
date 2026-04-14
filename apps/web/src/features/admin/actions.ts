'use server';

import { revalidatePath } from 'next/cache';
import { skillMutationSchema } from '@lavoval/contracts';
import { parseSkillConfig } from '@lavoval/registry';
import {
  createAdminSkill,
  deleteAdminSkill,
  requireAdminSession,
  updateAdminSkill,
} from '@/shared/api/server-client';

export async function createSkillAction(formData: FormData) {
  const session = await requireAdminSession();
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
  });

  await createAdminSkill(session.accessToken, payload);
  revalidatePath('/admin/skills');
}

export async function updateSkillAction(skillID: string, formData: FormData) {
  const session = await requireAdminSession();
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
  });

  await updateAdminSkill(session.accessToken, skillID, payload);
  revalidatePath('/admin/skills');
  revalidatePath(`/admin/skills/${skillID}`);
}

export async function deleteSkillAction(skillID: string) {
  const session = await requireAdminSession();
  await deleteAdminSkill(session.accessToken, skillID);
  revalidatePath('/admin/skills');
}
