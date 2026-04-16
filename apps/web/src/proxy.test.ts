import { describe, expect, it, vi } from 'vitest';

function createAccessToken(overrides: Partial<{ role: string; exp: number }> = {}) {
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

const { nextMock, redirectMock } = vi.hoisted(() => ({
  nextMock: vi.fn(() => ({ type: 'next' })),
  redirectMock: vi.fn((url: URL) => {
    const deleted: string[] = [];

    return {
      type: 'redirect',
      url: url.toString(),
      cookies: {
        delete: vi.fn((name: string) => {
          deleted.push(name);
        }),
      },
      deleted,
    };
  }),
}));

vi.mock('next/server', () => ({
  NextResponse: {
    next: nextMock,
    redirect: redirectMock,
  },
}));

import { proxy } from './proxy';

function createRequest(pathname: string, accessToken?: string) {
  return {
    url: `https://lavoval.com${pathname}`,
    nextUrl: { pathname },
    cookies: {
      get: vi.fn((name: string) => {
        if (name === 'csl_access_token' && accessToken) {
          return { value: accessToken };
        }

        return undefined;
      }),
    },
  };
}

describe('proxy auth gating', () => {
  it('redirects expired account access to sign-in and clears auth cookies', () => {
    const expiredToken = createAccessToken({ exp: Math.floor(Date.now() / 1000) - 60 });

    const response = proxy(createRequest('/account', expiredToken) as never) as unknown as {
      url: string;
      deleted: string[];
    };

    expect(response.url).toBe('https://lavoval.com/?auth=sign-in');
    expect(response.deleted).toEqual(
      expect.arrayContaining(['csl_access_token', 'csl_refresh_token', 'csl_session']),
    );
  });

  it('redirects non-admin users away from admin routes without clearing live auth', () => {
    const userToken = createAccessToken({ role: 'user' });

    const response = proxy(createRequest('/admin', userToken) as never) as unknown as {
      url: string;
      deleted: string[];
    };

    expect(response.url).toBe('https://lavoval.com/account');
    expect(response.deleted).toEqual([]);
  });

  it('allows valid admin access through', () => {
    const adminToken = createAccessToken();

    const response = proxy(createRequest('/admin', adminToken) as never);

    expect(response).toEqual({ type: 'next' });
    expect(nextMock).toHaveBeenCalled();
  });
});
