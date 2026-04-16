'use server';

import { redirect } from 'next/navigation';
import { z } from 'zod';
import {
  forgotPasswordRequestSchema,
  loginRequestSchema,
  mfaCompleteSignInRequestSchema,
  mfaDisableRequestSchema,
  mfaRegenerateRecoveryCodesRequestSchema,
  mfaVerifyEnrollmentRequestSchema,
  registerRequestSchema,
  resendVerificationRequestSchema,
  resetPasswordRequestSchema,
} from '@lavoval/contracts';
import {
  ApiError,
  clearSession,
  completeMFASignIn,
  disableMFA,
  enrollMFA,
  cancelMFAEnrollment,
  fetchMFAStatus,
  forgotPassword,
  login,
  logout,
  persistSession,
  register,
  regenerateMFARecoveryCodes,
  resendVerification,
  requireSession,
  resetPassword,
  withValidSession,
  verifyMFAEnrollment,
  verifyEmail,
} from '@/shared/api/server-client';
import { canAccessAdmin } from '@/shared/lib/rbac';

export type AuthFormState = {
  error: string | null;
  email: string;
  needsVerification?: boolean;
  mfaRequired?: boolean;
  challengeId?: string | null;
  recoveryMode?: boolean;
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

export type ForgotPasswordState = {
  error: string | null;
  success: string | null;
  email: string;
};

export type ResetPasswordState = {
  error: string | null;
  success: string | null;
  email: string;
};

export type MFAState = {
  error: string | null;
  success: string | null;
  successTitle?: string | null;
  enabled: boolean;
  pendingEnrollment: boolean;
  enrolledAt: string | null;
  secret: string | null;
  provisionUrl: string | null;
  recoveryCodes?: string[];
};

export async function loginAction(_previousState: AuthFormState, formData: FormData) {
  const email = String(formData.get('email') ?? '');
  const challengeId = String(formData.get('challengeId') ?? '');
  const recoveryMode = String(formData.get('recoveryMode') ?? '') === 'true';

  if (challengeId) {
    const parsed = mfaCompleteSignInRequestSchema.safeParse({
      challengeId,
      code: recoveryMode ? undefined : String(formData.get('code') ?? ''),
      recoveryCode: recoveryMode ? String(formData.get('recoveryCode') ?? '') : undefined,
    });

    if (!parsed.success) {
      return {
        error: recoveryMode
          ? 'Enter a recovery code to finish signing in.'
          : 'Enter the current 6-digit authenticator code.',
        email,
        needsVerification: false,
        mfaRequired: true,
        challengeId,
        recoveryMode,
      };
    }

    try {
      const response = await completeMFASignIn(parsed.data);
      await persistSession(response.data);
      redirect(canAccessAdmin(response.data.user.role) ? '/admin' : '/account');
    } catch (error) {
      if (error instanceof ApiError) {
        return {
          error:
            error.code === 'mfa_challenge_expired'
              ? 'This code prompt expired. Start sign-in again.'
              : error.code === 'mfa_recovery_code_invalid'
                ? 'That recovery code was not accepted.'
                : error.code === 'mfa_code_invalid'
                  ? 'That authenticator code was not accepted.'
                  : error.message || 'Could not complete sign in right now.',
          email,
          needsVerification: false,
          mfaRequired: true,
          challengeId,
          recoveryMode,
        };
      }

      return {
        error: 'Could not complete sign in right now.',
        email,
        needsVerification: false,
        mfaRequired: true,
        challengeId,
        recoveryMode,
      };
    }
  }

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
    redirect(canAccessAdmin(response.data.user.role) ? '/admin' : '/account');
  } catch (error) {
    if (error instanceof ApiError) {
      return {
        error:
          error.code === 'mfa_required'
            ? 'Enter the code from your authenticator app to finish signing in.'
            : error.status === 401
              ? 'Email or password is incorrect.'
              : error.status === 403
                ? 'Please confirm your email before signing in.'
                : error.message || 'Could not sign in right now. Please try again.',
        email,
        needsVerification: error.status === 403,
        mfaRequired: error.code === 'mfa_required',
        challengeId: typeof error.meta?.challengeId === 'string' ? error.meta.challengeId : null,
        recoveryMode: false,
      };
    }

    return {
      error: 'Could not sign in right now. Please try again.',
      email,
      needsVerification: false,
      mfaRequired: false,
      challengeId: null,
      recoveryMode: false,
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

export async function forgotPasswordAction(
  _previousState: ForgotPasswordState,
  formData: FormData,
) {
  const email = String(formData.get('email') ?? '');
  const parsed = forgotPasswordRequestSchema.safeParse({
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
    const response = await forgotPassword(parsed.data);
    return {
      error: null,
      success: `If ${response.data.email} belongs to a Lavoval account, a reset link is on its way.`,
      email: response.data.email,
    };
  } catch (error) {
    if (error instanceof ApiError) {
      return {
        error: error.message || 'Could not start password recovery right now.',
        success: null,
        email,
      };
    }

    return {
      error: 'Could not start password recovery right now.',
      success: null,
      email,
    };
  }
}

export async function resetPasswordAction(_previousState: ResetPasswordState, formData: FormData) {
  const token = String(formData.get('token') ?? '');
  const newPassword = String(formData.get('newPassword') ?? '');
  const email = String(formData.get('email') ?? '');

  const parsed = resetPasswordRequestSchema.safeParse({
    token,
    newPassword,
  });

  if (!parsed.success) {
    const passwordIssue = parsed.error.issues.find((issue) => issue.path[0] === 'newPassword');
    return {
      error: passwordIssue
        ? 'Password must be at least 12 characters.'
        : 'This reset link is invalid.',
      success: null,
      email,
    };
  }

  try {
    const response = await resetPassword(parsed.data);
    return {
      error: null,
      success: `Password updated for ${response.data.email}. You can sign in with your new password now.`,
      email: response.data.email,
    };
  } catch (error) {
    if (error instanceof ApiError) {
      return {
        error:
          error.status === 410
            ? 'This reset link expired. Request a fresh one below.'
            : error.status === 400
              ? 'This reset link is invalid. Request a fresh one below.'
              : error.message || 'Could not reset your password right now.',
        success: null,
        email,
      };
    }

    return {
      error: 'Could not reset your password right now.',
      success: null,
      email,
    };
  }
}

export async function loadMFAState(): Promise<MFAState> {
  return withValidSession(async (session) => {
    const response = await fetchMFAStatus(session.accessToken);

    return {
      error: null,
      success: null,
      successTitle: null,
      enabled: response.data.enabled,
      pendingEnrollment: response.data.pendingEnrollment,
      enrolledAt: response.data.enrolledAt ?? null,
      secret: null,
      provisionUrl: null,
      recoveryCodes: response.data.recoveryCodes ?? [],
    };
  });
}

export async function mfaSettingsAction(previousState: MFAState, formData: FormData) {
  const session = await requireSession();
  const intent = String(formData.get('intent') ?? '');

  if (intent === 'enroll') {
    try {
      const response = await enrollMFA(session.accessToken);
      const status = await fetchMFAStatus(session.accessToken);

      return {
        error: null,
        success:
          'Authenticator setup started. Add the secret below to your authenticator app, then confirm with a 6-digit code.',
        successTitle: 'Setup started',
        enabled: status.data.enabled,
        pendingEnrollment: status.data.pendingEnrollment,
        enrolledAt: status.data.enrolledAt ?? null,
        secret: response.data.secret,
        provisionUrl: response.data.provisionUrl,
        recoveryCodes: [],
      };
    } catch (error) {
      return {
        ...previousState,
        error: error instanceof ApiError ? error.message : 'Could not start MFA setup right now.',
        success: null,
        successTitle: null,
      };
    }
  }

  if (intent === 'verify') {
    const parsed = mfaVerifyEnrollmentRequestSchema.safeParse({
      code: String(formData.get('code') ?? ''),
    });

    if (!parsed.success) {
      return {
        ...previousState,
        error: 'Enter the 6-digit code from your authenticator app.',
        success: null,
      };
    }

    try {
      const response = await verifyMFAEnrollment(session.accessToken, parsed.data);

      return {
        error: null,
        success: 'Multi-factor authentication is now active for your account.',
        successTitle: 'MFA enabled',
        enabled: response.data.enabled,
        pendingEnrollment: response.data.pendingEnrollment,
        enrolledAt: response.data.enrolledAt ?? null,
        secret: null,
        provisionUrl: null,
        recoveryCodes: response.data.recoveryCodes ?? [],
      };
    } catch (error) {
      return {
        ...previousState,
        error:
          error instanceof ApiError && error.status === 400
            ? 'That authenticator code did not match. Check the current code and try again.'
            : error instanceof ApiError
              ? error.message
              : 'Could not verify MFA right now.',
        success: null,
        successTitle: null,
      };
    }
  }

  if (intent === 'cancel') {
    try {
      const response = await cancelMFAEnrollment(session.accessToken);

      return {
        error: null,
        success:
          'Authenticator setup has been cancelled. You can start again whenever you are ready.',
        successTitle: 'Setup cancelled',
        enabled: response.data.enabled,
        pendingEnrollment: response.data.pendingEnrollment,
        enrolledAt: response.data.enrolledAt ?? null,
        secret: null,
        provisionUrl: null,
        recoveryCodes: [],
      };
    } catch (error) {
      return {
        ...previousState,
        error: error instanceof ApiError ? error.message : 'Could not cancel MFA setup right now.',
        success: null,
        successTitle: null,
      };
    }
  }

  if (intent === 'disable') {
    const parsed = mfaDisableRequestSchema.safeParse({
      password: String(formData.get('password') ?? ''),
      code: String(formData.get('code') ?? ''),
    });

    if (!parsed.success) {
      return {
        ...previousState,
        error: 'Enter your current password and a valid 6-digit authenticator code.',
        success: null,
      };
    }

    try {
      const response = await disableMFA(session.accessToken, parsed.data);

      return {
        error: null,
        success: 'Multi-factor authentication has been turned off for this account.',
        successTitle: 'MFA disabled',
        enabled: response.data.enabled,
        pendingEnrollment: response.data.pendingEnrollment,
        enrolledAt: response.data.enrolledAt ?? null,
        secret: null,
        provisionUrl: null,
        recoveryCodes: [],
      };
    } catch (error) {
      return {
        ...previousState,
        error:
          error instanceof ApiError && error.status === 401
            ? 'Your current password is incorrect.'
            : error instanceof ApiError && error.status === 400
              ? 'That authenticator code did not match.'
              : error instanceof ApiError
                ? error.message
                : 'Could not disable MFA right now.',
        success: null,
        successTitle: null,
      };
    }
  }

  if (intent === 'regenerate') {
    const parsed = mfaRegenerateRecoveryCodesRequestSchema.safeParse({
      password: String(formData.get('password') ?? ''),
      code: String(formData.get('code') ?? ''),
    });

    if (!parsed.success) {
      return {
        ...previousState,
        error: 'Enter your current password and a valid 6-digit authenticator code.',
        success: null,
        successTitle: null,
      };
    }

    try {
      const response = await regenerateMFARecoveryCodes(session.accessToken, parsed.data);

      return {
        error: null,
        success:
          'Your previous recovery codes are now invalid. Store this fresh set somewhere safe before you leave this page.',
        successTitle: 'Recovery codes regenerated',
        enabled: response.data.enabled,
        pendingEnrollment: response.data.pendingEnrollment,
        enrolledAt: response.data.enrolledAt ?? null,
        secret: null,
        provisionUrl: null,
        recoveryCodes: response.data.recoveryCodes ?? [],
      };
    } catch (error) {
      return {
        ...previousState,
        error:
          error instanceof ApiError && error.status === 401
            ? 'Your current password is incorrect.'
            : error instanceof ApiError && error.status === 400
              ? 'That authenticator code did not match.'
              : error instanceof ApiError
                ? error.message
                : 'Could not regenerate recovery codes right now.',
        success: null,
        successTitle: null,
      };
    }
  }

  return {
    ...previousState,
    error: 'Unsupported MFA action.',
    success: null,
    successTitle: null,
  };
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
