'use server';

import { redirect } from 'next/navigation';
import { z } from 'zod';
import {
  loginRequestSchema,
  registerRequestSchema,
  resendVerificationRequestSchema,
} from '@lavoval/contracts';
import {
  ApiError,
  clearSession,
  login,
  logout,
  persistSession,
  register,
  resendVerification,
  requireSession,
  verifyEmail,
} from '@/shared/api/server-client';

export type AuthFormState = {
  error: string | null;
  email: string;
  needsVerification?: boolean;
};

export type RegisterFormState = {
  error: string | null;
  email: string;
  registered: boolean;
};

export type VerificationRequestState = {
  error: string | null;
  success: string | null;
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
            : error.status === 403
              ? 'Please confirm your email before signing in.'
              : error.message || 'Could not sign in right now. Please try again.',
        email,
        needsVerification: error.status === 403,
      };
    }

    return {
      error: 'Could not sign in right now. Please try again.',
      email,
      needsVerification: false,
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

export async function resendVerificationAction(
  _previousState: VerificationRequestState,
  formData: FormData,
) {
  const email = String(formData.get('email') ?? '');
  const parsed = resendVerificationRequestSchema.safeParse({
    email: formData.get('email'),
  });

  if (!parsed.success) {
    return {
      error: 'Enter a valid email address.',
      success: null,
      email,
    };
  }

  try {
    const response = await resendVerification(parsed.data);
    return {
      error: null,
      success: `We sent a fresh confirmation link to ${response.data.email}.`,
      email: response.data.email,
    };
  } catch (error) {
    if (error instanceof ApiError) {
      return {
        error:
          error.status === 404
            ? 'We could not find an account with that email.'
            : error.status === 409
              ? 'This email is already confirmed. You can sign in now.'
              : error.message || 'Could not resend the confirmation email right now.',
        success: null,
        email,
      };
    }

    return {
      error: 'Could not resend the confirmation email right now.',
      success: null,
      email,
    };
  }
}

export async function logoutAction() {
  const session = await requireSession();
  await logout(session.accessToken).catch(() => undefined);
  await clearSession();
  redirect('/');
}

export async function verifyEmailAction(token: string) {
  if (!token) {
    return {
      ok: false,
      title: 'Missing confirmation link',
      message: 'Use the latest email we sent you, or request a fresh confirmation link below.',
      email: '',
    };
  }

  try {
    const response = await verifyEmail({ token });
    return {
      ok: true,
      title: response.data.alreadyVerified ? 'Email already confirmed' : 'Email confirmed',
      message: response.data.alreadyVerified
        ? 'Your Lavoval account was already verified. You can sign in right away.'
        : 'Your email is now confirmed. You can sign in and start using Lavoval.',
      email: response.data.email,
    };
  } catch (error) {
    if (error instanceof ApiError) {
      return {
        ok: false,
        title: error.status === 410 ? 'Confirmation link expired' : 'Confirmation failed',
        message:
          error.status === 410
            ? 'That confirmation link expired. Request a fresh email below.'
            : error.message || 'We could not confirm this email right now.',
        email: '',
      };
    }

    return {
      ok: false,
      title: 'Confirmation failed',
      message: 'We could not confirm this email right now.',
      email: '',
    };
  }
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
