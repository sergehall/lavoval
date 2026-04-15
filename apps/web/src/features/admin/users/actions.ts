'use server';

import { redirect } from 'next/navigation';
import { adminUserUpdateSchema } from '@lavoval/contracts';
import { requireSession, updateAdminUser } from '@/shared/api/server-client';

export async function updateUserAction(userID: string, formData: FormData) {
  const session = await requireSession();

  const payload = adminUserUpdateSchema.parse({
    role: formData.get('role'),
    status: formData.get('status'),
  });

  await updateAdminUser(session.accessToken, userID, payload);
  redirect(`/admin/users/${userID}`);
}
