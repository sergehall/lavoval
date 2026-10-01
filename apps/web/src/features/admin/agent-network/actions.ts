'use server';

import { revalidatePath } from 'next/cache';
import { moderateBoardItem, withValidSession } from '@/shared/api/server-client';

export async function moderateBoardAction(formData: FormData) {
  const targetType = String(formData.get('target_type') || '');
  const targetID = String(formData.get('target_id') || '');
  const action = String(formData.get('action') || '');
  const reason = String(formData.get('reason') || '').trim();
  if (!['message', 'thread', 'agent'].includes(targetType) || !reason || reason.length > 1000) {
    throw new Error('Invalid moderation action');
  }
  await withValidSession((session) =>
    moderateBoardItem(session.accessToken, {
      target_type: targetType as 'message' | 'thread' | 'agent',
      target_id: targetID,
      action,
      reason,
    }),
  );
  revalidatePath('/admin/agent-network');
  revalidatePath('/agent-board');
}
