'use server';

import { revalidatePath } from 'next/cache';
import { profileUpdateSchema } from '@lavoval/contracts';
import { requireSession, updateProfile } from '@/shared/api/server-client';

export async function updateProfileAction(formData: FormData) {
  const session = await requireSession();
  const payload = profileUpdateSchema.parse({
    firstName: formData.get('firstName'),
    lastName: formData.get('lastName'),
    bio: formData.get('bio') || null,
    timezone: formData.get('timezone'),
  });

  await updateProfile(session.accessToken, payload);
  revalidatePath('/account/profile');
}
