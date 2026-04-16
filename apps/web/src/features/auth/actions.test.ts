import { beforeEach, describe, expect, it, vi } from 'vitest';

const {
  requireSessionMock,
  loginMock,
  completeMFASignInMock,
  persistSessionMock,
  fetchMFAStatusMock,
  enrollMFAMock,
  verifyMFAEnrollmentMock,
  disableMFAMock,
  regenerateMFARecoveryCodesMock,
  verifyEmailMock,
  resetPasswordMock,
  redirectMock,
} = vi.hoisted(() => ({
  requireSessionMock: vi.fn(),
  loginMock: vi.fn(),
  completeMFASignInMock: vi.fn(),
  persistSessionMock: vi.fn(),
  fetchMFAStatusMock: vi.fn(),
  enrollMFAMock: vi.fn(),
  verifyMFAEnrollmentMock: vi.fn(),
  disableMFAMock: vi.fn(),
  regenerateMFARecoveryCodesMock: vi.fn(),
  verifyEmailMock: vi.fn(),
  resetPasswordMock: vi.fn(),
  redirectMock: vi.fn((path: string) => {
    throw new Error(`NEXT_REDIRECT:${path}`);
  }),
}));

vi.mock('next/navigation', () => ({
  redirect: redirectMock,
}));

vi.mock('@/shared/api/server-client', () => {
  class ApiError extends Error {
    status: number;
    code?: string;
    meta?: Record<string, unknown>;

    constructor(message: string, status: number, code?: string, meta?: Record<string, unknown>) {
      super(message);
      this.name = 'ApiError';
      this.status = status;
      this.code = code;
      this.meta = meta;
    }
  }

  return {
    ApiError,
    clearSession: vi.fn(),
    completeMFASignIn: completeMFASignInMock,
    disableMFA: disableMFAMock,
    enrollMFA: enrollMFAMock,
    cancelMFAEnrollment: vi.fn(),
    fetchMFAStatus: fetchMFAStatusMock,
    forgotPassword: vi.fn(),
    login: loginMock,
    logout: vi.fn(),
    persistSession: persistSessionMock,
    register: vi.fn(),
    regenerateMFARecoveryCodes: regenerateMFARecoveryCodesMock,
    resendVerification: vi.fn(),
    requireSession: requireSessionMock,
    resetPassword: resetPasswordMock,
    verifyMFAEnrollment: verifyMFAEnrollmentMock,
    verifyEmail: verifyEmailMock,
    withValidSession: vi.fn(),
  };
});

import { ApiError } from '@/shared/api/server-client';
import {
  loginAction,
  mfaSettingsAction,
  resetPasswordAction,
  verifyEmailAction,
  type MFAState,
  type ResetPasswordState,
} from './actions';

function makeFormData(values: Record<string, string>) {
  const formData = new FormData();
  for (const [key, value] of Object.entries(values)) {
    formData.set(key, value);
  }
  return formData;
}

describe('auth actions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    requireSessionMock.mockResolvedValue({
      accessToken: 'access-token',
      refreshToken: 'refresh-token',
      user: {
        id: '11111111-1111-4111-8111-111111111111',
        email: 'serge@example.com',
        role: 'user',
      },
    });
  });

  it('returns MFA challenge state when password sign-in requires a second factor', async () => {
    loginMock.mockRejectedValueOnce(
      new ApiError('MFA required', 401, 'mfa_required', {
        challengeId: '11111111-1111-4111-8111-111111111111',
      }),
    );

    const result = await loginAction(
      { error: null, email: '' },
      makeFormData({
        email: 'serge@example.com',
        password: 'password123',
      }),
    );

    expect(result).toEqual({
      error: 'Enter the code from your authenticator app to finish signing in.',
      email: 'serge@example.com',
      needsVerification: false,
      mfaRequired: true,
      challengeId: '11111111-1111-4111-8111-111111111111',
      recoveryMode: false,
    });
    expect(persistSessionMock).not.toHaveBeenCalled();
    expect(redirectMock).not.toHaveBeenCalled();
  });

  it('validates missing recovery code before attempting MFA completion', async () => {
    const result = await loginAction(
      { error: null, email: '' },
      makeFormData({
        email: 'serge@example.com',
        challengeId: '11111111-1111-4111-8111-111111111111',
        recoveryMode: 'true',
      }),
    );

    expect(result).toEqual({
      error: 'Enter a recovery code to finish signing in.',
      email: 'serge@example.com',
      needsVerification: false,
      mfaRequired: true,
      challengeId: '11111111-1111-4111-8111-111111111111',
      recoveryMode: true,
    });
    expect(completeMFASignInMock).not.toHaveBeenCalled();
  });

  it('persists the session and redirects after successful password sign-in', async () => {
    loginMock.mockResolvedValueOnce({
      data: {
        accessToken: 'access-token',
        refreshToken: 'refresh-token',
        user: {
          id: '11111111-1111-4111-8111-111111111111',
          email: 'serge@example.com',
          role: 'admin',
        },
      },
    });
    persistSessionMock.mockResolvedValueOnce(undefined);

    await expect(
      loginAction(
        { error: null, email: '' },
        makeFormData({
          email: 'serge@example.com',
          password: 'password123',
        }),
      ),
    ).rejects.toThrow('NEXT_REDIRECT:/admin');

    expect(loginMock).toHaveBeenCalledWith({
      email: 'serge@example.com',
      password: 'password123',
    });
    expect(persistSessionMock).toHaveBeenCalledTimes(1);
    expect(redirectMock).toHaveBeenCalledWith('/admin');
  });

  it('persists the MFA session and redirects after successful challenge completion', async () => {
    completeMFASignInMock.mockResolvedValueOnce({
      data: {
        accessToken: 'access-token',
        refreshToken: 'refresh-token',
        user: {
          id: '11111111-1111-4111-8111-111111111111',
          email: 'serge@example.com',
          role: 'user',
        },
      },
    });
    persistSessionMock.mockResolvedValueOnce(undefined);

    await expect(
      loginAction(
        { error: null, email: '' },
        makeFormData({
          email: 'serge@example.com',
          challengeId: '11111111-1111-4111-8111-111111111111',
          code: '123456',
        }),
      ),
    ).rejects.toThrow('NEXT_REDIRECT:/account');

    expect(completeMFASignInMock).toHaveBeenCalledWith({
      challengeId: '11111111-1111-4111-8111-111111111111',
      code: '123456',
      recoveryCode: undefined,
    });
    expect(persistSessionMock).toHaveBeenCalledTimes(1);
    expect(redirectMock).toHaveBeenCalledWith('/account');
  });

  it('returns an expired state when email verification token is no longer valid', async () => {
    verifyEmailMock.mockRejectedValueOnce(
      new ApiError('That confirmation link expired. Request a fresh email below.', 410),
    );

    const result = await verifyEmailAction('token-token-token-token-1234');

    expect(result).toEqual({
      ok: false,
      title: 'Confirmation link expired',
      message: 'That confirmation link expired. Request a fresh email below.',
      email: '',
    });
  });

  it('maps invalid reset-password submissions to the password validation message', async () => {
    const result = await resetPasswordAction(
      { error: null, success: null, email: '' },
      makeFormData({
        token: 'token-token-token-token-1234',
        newPassword: 'short',
        email: 'serge@example.com',
      }),
    );

    expect(result).toEqual({
      error: 'Password must be at least 12 characters.',
      success: null,
      email: 'serge@example.com',
    });
    expect(resetPasswordMock).not.toHaveBeenCalled();
  });

  it('returns a friendly message when the reset token is rejected by the API', async () => {
    resetPasswordMock.mockRejectedValueOnce(
      new ApiError('This reset link is invalid. Request a fresh one below.', 400),
    );

    const result = await resetPasswordAction(
      { error: null, success: null, email: '' } satisfies ResetPasswordState,
      makeFormData({
        token: 'token-token-token-token-1234',
        newPassword: 'long-enough-password',
        email: 'serge@example.com',
      }),
    );

    expect(result).toEqual({
      error: 'This reset link is invalid. Request a fresh one below.',
      success: null,
      email: 'serge@example.com',
    });
  });

  it('loads the pending MFA state after starting enrollment', async () => {
    enrollMFAMock.mockResolvedValueOnce({
      data: {
        secret: 'BASE32SECRET123456',
        provisionUrl: 'otpauth://totp/Lavoval:serge@example.com',
      },
    });
    fetchMFAStatusMock.mockResolvedValueOnce({
      data: {
        enabled: false,
        pendingEnrollment: true,
        enrolledAt: null,
      },
    });

    const result = await mfaSettingsAction(
      {
        error: null,
        success: null,
        successTitle: null,
        enabled: false,
        pendingEnrollment: false,
        enrolledAt: null,
        secret: null,
        provisionUrl: null,
        recoveryCodes: [],
      },
      makeFormData({ intent: 'enroll' }),
    );

    expect(enrollMFAMock).toHaveBeenCalledWith('access-token');
    expect(fetchMFAStatusMock).toHaveBeenCalledWith('access-token');
    expect(result).toEqual({
      error: null,
      success:
        'Authenticator setup started. Add the secret below to your authenticator app, then confirm with a 6-digit code.',
      successTitle: 'Setup started',
      enabled: false,
      pendingEnrollment: true,
      enrolledAt: null,
      secret: 'BASE32SECRET123456',
      provisionUrl: 'otpauth://totp/Lavoval:serge@example.com',
      recoveryCodes: [],
    });
  });

  it('maps MFA verify validation failures before calling the API', async () => {
    const previousState: MFAState = {
      error: null,
      success: 'Setup started',
      successTitle: 'Setup started',
      enabled: false,
      pendingEnrollment: true,
      enrolledAt: null,
      secret: 'BASE32SECRET123456',
      provisionUrl: 'otpauth://totp/Lavoval:serge@example.com',
      recoveryCodes: [],
    };

    const result = await mfaSettingsAction(
      previousState,
      makeFormData({ intent: 'verify', code: '12' }),
    );

    expect(result).toEqual({
      ...previousState,
      error: 'Enter the 6-digit code from your authenticator app.',
      success: null,
    });
    expect(verifyMFAEnrollmentMock).not.toHaveBeenCalled();
  });

  it('maps MFA disable authentication failures to a user-friendly message', async () => {
    const previousState: MFAState = {
      error: null,
      success: null,
      successTitle: null,
      enabled: true,
      pendingEnrollment: false,
      enrolledAt: '2026-01-01T00:00:00Z',
      secret: null,
      provisionUrl: null,
      recoveryCodes: [],
    };
    disableMFAMock.mockRejectedValueOnce(new ApiError('Unauthorized', 401, 'invalid_credentials'));

    const result = await mfaSettingsAction(
      previousState,
      makeFormData({
        intent: 'disable',
        password: 'password123',
        code: '123456',
      }),
    );

    expect(disableMFAMock).toHaveBeenCalledWith('access-token', {
      password: 'password123',
      code: '123456',
    });
    expect(result).toEqual({
      ...previousState,
      error: 'Your current password is incorrect.',
      success: null,
      successTitle: null,
    });
  });

  it('returns fresh recovery codes after successful regeneration', async () => {
    regenerateMFARecoveryCodesMock.mockResolvedValueOnce({
      data: {
        enabled: true,
        pendingEnrollment: false,
        enrolledAt: '2026-01-01T00:00:00Z',
        recoveryCodes: ['code-one', 'code-two'],
      },
    });

    const result = await mfaSettingsAction(
      {
        error: null,
        success: null,
        successTitle: null,
        enabled: true,
        pendingEnrollment: false,
        enrolledAt: '2026-01-01T00:00:00Z',
        secret: null,
        provisionUrl: null,
        recoveryCodes: [],
      },
      makeFormData({
        intent: 'regenerate',
        password: 'password123',
        code: '123456',
      }),
    );

    expect(regenerateMFARecoveryCodesMock).toHaveBeenCalledWith('access-token', {
      password: 'password123',
      code: '123456',
    });
    expect(result).toEqual({
      error: null,
      success:
        'Your previous recovery codes are now invalid. Store this fresh set somewhere safe before you leave this page.',
      successTitle: 'Recovery codes regenerated',
      enabled: true,
      pendingEnrollment: false,
      enrolledAt: '2026-01-01T00:00:00Z',
      secret: null,
      provisionUrl: null,
      recoveryCodes: ['code-one', 'code-two'],
    });
  });
});
