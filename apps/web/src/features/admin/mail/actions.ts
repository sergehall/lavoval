'use server';

import { revalidatePath } from 'next/cache';
import {
  cleanupAdminMailRetention,
  createAdminMailSuppression,
  deleteAdminMailSuppression,
  replayAdminMailJob,
  requeueAdminDeadLetter,
  requireAdminSession,
} from '@/shared/api/server-client';

export async function requeueDeadLetterAction(jobID: string) {
  const session = await requireAdminSession();
  await requeueAdminDeadLetter(session.accessToken, jobID);
  revalidatePath('/admin/mail');
}

export async function replayMailJobAction(jobID: string) {
  const session = await requireAdminSession();
  await replayAdminMailJob(session.accessToken, jobID);
  revalidatePath('/admin/mail');
}

export async function createMailSuppressionAction(formData: FormData) {
  const session = await requireAdminSession();
  await createAdminMailSuppression(session.accessToken, {
    kind: String(formData.get('kind')) as 'email' | 'domain',
    value: String(formData.get('value') ?? ''),
    reason: String(formData.get('reason') ?? ''),
  });
  revalidatePath('/admin/mail');
}

export async function deleteMailSuppressionAction(suppressionID: string) {
  const session = await requireAdminSession();
  await deleteAdminMailSuppression(session.accessToken, suppressionID);
  revalidatePath('/admin/mail');
}

export async function cleanupMailRetentionAction() {
  const session = await requireAdminSession();
  await cleanupAdminMailRetention(session.accessToken);
  revalidatePath('/admin/mail');
}
