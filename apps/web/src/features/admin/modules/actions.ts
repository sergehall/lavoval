'use server';

import { revalidatePath } from 'next/cache';
import { moduleMutationSchema } from '@lavoval/contracts';
import {
  createAdminModule,
  deleteAdminModule,
  requireSession,
  updateAdminModule,
} from '@/shared/api/server-client';

function parseModuleForm(formData: FormData) {
  const rawPosition = formData.get('position');
  return moduleMutationSchema.parse({
    slug: formData.get('slug'),
    title: formData.get('title'),
    summary: formData.get('summary'),
    content: formData.get('content'),
    position: rawPosition ? parseInt(rawPosition as string, 10) : undefined,
    status: formData.get('status'),
  });
}

export async function createModuleAction(skillID: string, formData: FormData) {
  const session = await requireSession();
  const payload = parseModuleForm(formData);
  await createAdminModule(session.accessToken, skillID, payload);
  revalidatePath(`/admin/skills/${skillID}`);
}

export async function updateModuleAction(skillID: string, moduleID: string, formData: FormData) {
  const session = await requireSession();
  const payload = parseModuleForm(formData);
  await updateAdminModule(session.accessToken, skillID, moduleID, payload);
  revalidatePath(`/admin/skills/${skillID}`);
}

export async function deleteModuleAction(skillID: string, moduleID: string) {
  const session = await requireSession();
  await deleteAdminModule(session.accessToken, skillID, moduleID);
  revalidatePath(`/admin/skills/${skillID}`);
}
