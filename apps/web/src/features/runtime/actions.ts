'use server';

import { redirect } from 'next/navigation';
import { revalidatePath } from 'next/cache';
import { ApiError, runSkill, withValidSession } from '@/shared/api/server-client';

export async function runSkillAction(skillID: string, formData: FormData) {
  const text = String(formData.get('text') ?? '').trim();
  const skillPath = `/skills/${skillID}`;

  return withValidSession(async (session) => {
    try {
      const response = await runSkill(session.accessToken, {
        skillId: skillID,
        input: text ? { text } : {},
      });

      revalidatePath('/account/runs');
      redirect(`/account/runs/${response.data.id}`);
    } catch (error) {
      if (error instanceof ApiError) {
        const params = new URLSearchParams();
        params.set('runError', error.code ?? 'skill_run_failed');
        params.set('runMessage', error.message);
        redirect(`${skillPath}?${params.toString()}` as any);
      }

      throw error;
    }
  });
}
