'use server';

import { revalidatePath } from 'next/cache';
import { skillMutationSchema } from '@lavoval/contracts';
import {
  createMySkill,
  deleteMySkill,
  requireSession,
  updateMySkill,
} from '@/shared/api/server-client';

export async function createOwnSkillAction(formData: FormData) {
  const session = await requireSession();
  const payload = skillMutationSchema.parse({
    slug: formData.get('slug'),
    title: formData.get('title'),
    summary: formData.get('summary'),
    description: formData.get('description'),
    status: formData.get('status'),
    visibility: formData.get('visibility'),
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
    status: formData.get('status'),
    visibility: formData.get('visibility'),
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
