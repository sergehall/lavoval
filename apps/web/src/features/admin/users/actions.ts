'use server';

import { redirect } from 'next/navigation';
import { adminUserRoleUpdateSchema, adminUserStatusUpdateSchema } from '@lavoval/contracts';
import {
  requireAdminSession,
  updateAdminUserRole,
  updateAdminUserStatus,
} from '@/shared/api/server-client';
import { isRootOwner } from '@/shared/lib/rbac';

export async function updateUserStatusAction(userID: string, formData: FormData) {
  const session = await requireAdminSession();

  const rawReason = formData.get('reason');
  const payload = adminUserStatusUpdateSchema.parse({
    status: formData.get('status'),
    reason: rawReason && String(rawReason).trim() ? String(rawReason).trim() : undefined,
  });

  await updateAdminUserStatus(session.accessToken, userID, payload);
  redirect(`/admin/users/${userID}`);
}

export async function updateUserRoleAction(userID: string, formData: FormData) {
  const session = await requireAdminSession();
  if (!isRootOwner(session.user.role)) {
    throw new Error('Only root_owner can change elevated roles.');
  }

  const rawReason = formData.get('reason');
  const payload = adminUserRoleUpdateSchema.parse({
    role: formData.get('role'),
    reason: rawReason && String(rawReason).trim() ? String(rawReason).trim() : '',
  });

  await updateAdminUserRole(session.accessToken, userID, payload);
  redirect(`/admin/users/${userID}`);
}
