'use server';

import { redirect } from 'next/navigation';
import { adminUserUpdateSchema } from '@lavoval/contracts';
import { requireAdminSession, updateAdminUser } from '@/shared/api/server-client';

export async function updateUserAction(userID: string, formData: FormData) {
  const session = await requireAdminSession();

  const rawReason = formData.get('reason');
  const payload = adminUserUpdateSchema.parse({
    role: formData.get('role'),
    status: formData.get('status'),
    reason: rawReason && String(rawReason).trim() ? String(rawReason).trim() : undefined,
  });

  await updateAdminUser(session.accessToken, userID, payload);
  redirect(`/admin/users/${userID}`);
}
