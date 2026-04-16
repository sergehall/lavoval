import { beforeEach, describe, expect, it, vi } from 'vitest';

const { cookieStore, cookiesMock, redirectMock } = vi.hoisted(() => {
  const cookieStore = {
    get: vi.fn(),
    set: vi.fn(),
    delete: vi.fn(),
  };

  return {
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

import { ApiError, withValidSession } from './server-client';

describe('withValidSession', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cookieStore.get.mockImplementation((name: string) => {
      if (name !== 'csl_session') {
        return undefined;
      }

      return {
        value: JSON.stringify({
          user: {
            id: 'user-1',
            email: 'serge@example.com',
            firstName: 'Serge',
            lastName: 'Hall',
            role: 'admin',
          },
          accessToken: 'access-token',
          refreshToken: 'refresh-token',
        }),
      };
    });
  });

  it('returns the handler result when the session is valid', async () => {
    const result = await withValidSession(async (session) => session.user.email);

    expect(result).toBe('serge@example.com');
    expect(cookieStore.delete).not.toHaveBeenCalled();
    expect(redirectMock).not.toHaveBeenCalled();
  });

  it('redirects to login on 401 without mutating cookies during render', async () => {
    await expect(
      withValidSession(async () => {
        throw new ApiError('Unauthorized', 401);
      }),
    ).rejects.toThrow('NEXT_REDIRECT:/?auth=sign-in');

    expect(cookieStore.delete).not.toHaveBeenCalled();
    expect(redirectMock).toHaveBeenCalledWith('/?auth=sign-in');
  });
});
