'use server';

import { redirect } from 'next/navigation';
import { revalidatePath } from 'next/cache';
import { requireSession, runSkill } from '@/shared/api/server-client';

export async function runSkillAction(skillID: string, formData: FormData) {
  const session = await requireSession();
  const text = String(formData.get('text') ?? '').trim();

  const response = await runSkill(session.accessToken, {
    skillId: skillID,
    input: text ? { text } : {},
  });

  revalidatePath('/account/runs');
  redirect(`/account/runs/${response.data.id}`);
}
