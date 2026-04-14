'use server';

import { redirect } from 'next/navigation';
import { z } from 'zod';
import { loginRequestSchema, registerRequestSchema } from '@lavoval/contracts';
import {
  ApiError,
  clearSession,
  login,
  logout,
  persistSession,
  register,
  requireSession,
} from '@/shared/api/server-client';

export type AuthFormState = {
  error: string | null;
  email: string;
};

export async function loginAction(_previousState: AuthFormState, formData: FormData) {
  const email = String(formData.get('email') ?? '');
  const parsed = loginRequestSchema.safeParse({
    email: formData.get('email'),
    password: formData.get('password'),
  });

  if (!parsed.success) {
    return {
      error: formatAuthValidationError(parsed.error),
      email,
    };
  }

  try {
    const response = await login(parsed.data);
    await persistSession(response.data);
    redirect(response.data.user.role === 'admin' ? '/admin' : '/account');
  } catch (error) {
    if (error instanceof ApiError) {
      return {
        error:
          error.status === 401
            ? 'Email or password is incorrect.'
            : error.message || 'Could not sign in right now. Please try again.',
        email,
      };
    }

    return {
      error: 'Could not sign in right now. Please try again.',
      email,
    };
  }
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

function formatAuthValidationError(error: z.ZodError) {
  const firstIssue = error.issues[0];
  if (!firstIssue) {
    return 'Please check your credentials and try again.';
  }

  if (firstIssue.path[0] === 'email') {
    return 'Enter a valid email address.';
  }

  if (firstIssue.path[0] === 'password') {
    return 'Password must be at least 8 characters.';
  }

  return firstIssue.message;
}
