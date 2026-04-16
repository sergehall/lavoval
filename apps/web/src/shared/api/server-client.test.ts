import { beforeEach, describe, expect, it, vi } from 'vitest';

function createAccessToken(
  overrides: Partial<{ uid: string; email: string; role: string; exp: number }> = {},
) {
  const payload = {
    uid: '11111111-1111-4111-8111-111111111111',
    email: 'serge@example.com',
    role: 'admin',
    exp: Math.floor(Date.now() / 1000) + 60 * 60,
    ...overrides,
  };

  const encode = (value: unknown) => Buffer.from(JSON.stringify(value)).toString('base64url');
  return `${encode({ alg: 'HS256', typ: 'JWT' })}.${encode(payload)}.signature`;
}

const { apiClientMock, cookieStore, cookiesMock, redirectMock } = vi.hoisted(() => {
  const apiClientMock = {
    auth: {
      login: vi.fn(),
      mfaCompleteSignIn: vi.fn(),
      completeGoogleOAuth: vi.fn(),
      completeGitHubOAuth: vi.fn(),
      register: vi.fn(),
      verifyEmail: vi.fn(),
      resendVerification: vi.fn(),
      forgotPassword: vi.fn(),
      resetPassword: vi.fn(),
      mfaStatus: vi.fn(),
      mfaEnroll: vi.fn(),
      mfaCancelEnrollment: vi.fn(),
      mfaVerifyEnrollment: vi.fn(),
      mfaDisable: vi.fn(),
      mfaRegenerateRecoveryCodes: vi.fn(),
      logout: vi.fn(),
    },
    me: {
      profile: vi.fn(),
      security: vi.fn(),
      updateProfile: vi.fn(),
      skills: vi.fn(),
      skill: vi.fn(),
      createSkill: vi.fn(),
      updateSkill: vi.fn(),
      deleteSkill: vi.fn(),
    },
    skills: {
      list: vi.fn(),
      detail: vi.fn(),
    },
    admin: new Proxy(
      {},
      {
        get: () => vi.fn(),
      },
    ),
    runtime: new Proxy(
      {},
      {
        get: () => vi.fn(),
      },
    ),
  };

  const cookieStore = {
    get: vi.fn(),
    set: vi.fn(),
    delete: vi.fn(),
  };

  return {
    apiClientMock,
    cookieStore,
    cookiesMock: vi.fn(async () => cookieStore),
    redirectMock: vi.fn((path: string) => {
      throw new Error(`NEXT_REDIRECT:${path}`);
    }),
  };
});

vi.mock('server-only', () => ({}));
vi.mock('next/headers', () => ({
  cookies: cookiesMock,
}));
vi.mock('next/navigation', () => ({
  redirect: redirectMock,
}));
vi.mock('@lavoval/sdk', async () => {
  class ApiClientError extends Error {
    status: number;
    code?: string;
    meta?: Record<string, unknown>;

    constructor(message: string, status = 500, code?: string, meta?: Record<string, unknown>) {
      super(message);
      this.status = status;
      this.code = code;
      this.meta = meta;
    }
  }

  return {
    ApiClientError,
    createApiClient: vi.fn(() => apiClientMock),
  };
});

import { ApiError, getSession, withValidSession } from './server-client';

describe('server session handling', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    const accessToken = createAccessToken();
    cookieStore.get.mockImplementation((name: string) => {
      if (name === 'csl_access_token') {
        return { value: accessToken };
      }

      if (name === 'csl_session') {
        return {
          value: JSON.stringify({
            user: {
              id: '11111111-1111-4111-8111-111111111111',
              email: 'serge@example.com',
              firstName: 'Serge',
              lastName: 'Hall',
              role: 'admin',
            },
            accessToken,
            refreshToken: 'refresh-token',
          }),
        };
      }

      return undefined;
    });

    apiClientMock.me.profile.mockResolvedValue({
      data: {
        userId: '11111111-1111-4111-8111-111111111111',
        role: 'admin',
        firstName: 'Serge',
        lastName: 'Hall',
        bio: null,
        timezone: 'UTC',
        username: null,
        avatarUrl: null,
        location: null,
        skills: null,
        languages: null,
        websiteUrl: null,
        linkedinUrl: null,
        githubUrl: null,
        twitterUrl: null,
        availabilityStatus: 'open',
        isPublicProfile: true,
      },
    });
  });

  it('returns the session only when the cookie payload matches a live access token', async () => {
    const session = await getSession();

    expect(session?.user.email).toBe('serge@example.com');
  });

  it('drops the session when the access token is expired', async () => {
    const expiredToken = createAccessToken({ exp: Math.floor(Date.now() / 1000) - 60 });

    cookieStore.get.mockImplementation((name: string) => {
      if (name === 'csl_access_token') {
        return { value: expiredToken };
      }

      if (name === 'csl_session') {
        return {
          value: JSON.stringify({
            user: {
              id: '11111111-1111-4111-8111-111111111111',
              email: 'serge@example.com',
              role: 'admin',
            },
            accessToken: expiredToken,
            refreshToken: 'refresh-token',
          }),
        };
      }

      return undefined;
    });

    await expect(getSession()).resolves.toBeNull();
  });

  it('returns the handler result when the session is valid', async () => {
    const result = await withValidSession(async (session) => session.user.email);

    expect(result).toBe('serge@example.com');
    expect(apiClientMock.me.profile).toHaveBeenCalledTimes(1);
    expect(redirectMock).not.toHaveBeenCalled();
  });

  it('redirects to login on backend 401', async () => {
    apiClientMock.me.profile.mockRejectedValueOnce(new ApiError('Unauthorized', 401));

    await expect(withValidSession(async () => 'never')).rejects.toThrow(
      'NEXT_REDIRECT:/?auth=sign-in',
    );

    expect(redirectMock).toHaveBeenCalledWith('/?auth=sign-in');
  });
});
