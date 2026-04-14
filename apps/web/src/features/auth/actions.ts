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

export type RegisterFormState = {
  error: string | null;
  email: string;
  registered: boolean;
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

export async function registerAction(_previousState: RegisterFormState, formData: FormData) {
  const email = String(formData.get('email') ?? '');
  const parsed = registerRequestSchema.safeParse({
    email: formData.get('email'),
    password: formData.get('password'),
    firstName: formData.get('firstName'),
    lastName: formData.get('lastName'),
  });

  if (!parsed.success) {
    return {
      error: formatRegisterValidationError(parsed.error),
      email,
      registered: false,
    };
  }

  try {
    await register(parsed.data);

    return {
      error: null,
      email: parsed.data.email,
      registered: true,
    };
  } catch (error) {
    if (error instanceof ApiError) {
      return {
        error:
          error.status === 409
            ? 'An account with this email already exists.'
            : error.message || 'Could not create your account right now. Please try again.',
        email,
        registered: false,
      };
    }

    return {
      error: 'Could not create your account right now. Please try again.',
      email,
      registered: false,
    };
  }
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

function formatRegisterValidationError(error: z.ZodError) {
  const firstIssue = error.issues[0];
  if (!firstIssue) {
    return 'Please review the registration details and try again.';
  }

  if (firstIssue.path[0] === 'email') {
    return 'Enter a valid email address.';
  }

  if (firstIssue.path[0] === 'password') {
    return 'Password must be at least 12 characters.';
  }

  if (firstIssue.path[0] === 'firstName' || firstIssue.path[0] === 'lastName') {
    return 'First and last name must be at least 2 characters.';
  }

  return firstIssue.message;
}
