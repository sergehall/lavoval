'use server';

import { redirect } from 'next/navigation';
import { loginRequestSchema, registerRequestSchema } from '@lavoval/contracts';
import {
  clearSession,
  login,
  logout,
  persistSession,
  register,
  requireSession,
} from '@/shared/api/server-client';

export async function loginAction(formData: FormData) {
  const payload = loginRequestSchema.parse({
    email: formData.get('email'),
    password: formData.get('password'),
  });

  const response = await login(payload);
  await persistSession(response.data);
  redirect(response.data.user.role === 'admin' ? '/admin' : '/account');
}

export async function registerAction(formData: FormData) {
  const payload = registerRequestSchema.parse({
    email: formData.get('email'),
    password: formData.get('password'),
    firstName: formData.get('firstName'),
    lastName: formData.get('lastName'),
  });

  const response = await register(payload);
  await persistSession(response.data);
  redirect('/account');
}

export async function logoutAction() {
  const session = await requireSession();
  await logout(session.accessToken).catch(() => undefined);
  await clearSession();
  redirect('/login');
}
