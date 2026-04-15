'use server';

import { revalidatePath } from 'next/cache';
import { enrollmentAssignSchema, enrollmentUpdateSchema } from '@lavoval/contracts';
import {
  assignSkillToUser,
  requireSession,
  updateAdminEnrollment,
} from '@/shared/api/server-client';

export async function assignSkillAction(formData: FormData) {
  const session = await requireSession();

  const payload = enrollmentAssignSchema.parse({
    userId: formData.get('userId'),
    skillId: formData.get('skillId'),
  });

  await assignSkillToUser(session.accessToken, payload);
  revalidatePath('/admin/enrollments');
}

export async function updateEnrollmentAction(enrollmentID: string, formData: FormData) {
  const session = await requireSession();

  const raw = formData.get('progressPercent');
  const payload = enrollmentUpdateSchema.parse({
    status: formData.get('status'),
    progressPercent: raw ? parseInt(raw as string, 10) : 0,
  });

  await updateAdminEnrollment(session.accessToken, enrollmentID, payload);
  revalidatePath('/admin/enrollments');
}
