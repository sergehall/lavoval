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
  nextMock: vi.fn((..._args: unknown[]) => ({ type: 'next', headers: new Headers() })),
  redirectMock: vi.fn((url: URL) => {
    const deleted: string[] = [];

    return {
      type: 'redirect',
      url: url.toString(),
      headers: new Headers(),
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

import { config, proxy } from './proxy';

function createRequest(pathname: string, accessToken?: string) {
  return {
    url: `https://lavoval.com${pathname}`,
    nextUrl: { pathname },
    headers: new Headers(),
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

function expectNonceBasedScriptPolicy(contentSecurityPolicy: string | null) {
  expect(contentSecurityPolicy).toContain("frame-ancestors 'none'");
  expect(contentSecurityPolicy).toMatch(/script-src[^;]*'nonce-[^']+'/);
  expect(contentSecurityPolicy).toContain("'strict-dynamic'");
  expect(contentSecurityPolicy).not.toContain("'unsafe-inline'");
  expect(contentSecurityPolicy?.match(/script-src[^;]*/)?.[0]).not.toContain('data:');
}

function expectProductionOnlySecureSources(contentSecurityPolicy: string | null) {
  const connectSource = contentSecurityPolicy?.match(/connect-src[^;]*/)?.[0];

  expect(connectSource).toBe("connect-src 'self' https://api.lavoval.com");
  expect(contentSecurityPolicy).not.toContain('http://');
  expect(contentSecurityPolicy).not.toContain('ws://');
  expect(contentSecurityPolicy).not.toContain("'unsafe-eval'");
}

describe('proxy auth gating', () => {
  it('redirects expired account access to sign-in and clears auth cookies', () => {
    const expiredToken = createAccessToken({ exp: Math.floor(Date.now() / 1000) - 60 });

    const response = proxy(createRequest('/account', expiredToken) as never) as unknown as {
      headers: Headers;
      url: string;
      deleted: string[];
    };

    expect(response.url).toBe('https://lavoval.com/?auth=sign-in');
    expect(response.headers.get('X-Content-Type-Options')).toBe('nosniff');
    expect(response.headers.get('X-Frame-Options')).toBe('DENY');
    expectNonceBasedScriptPolicy(response.headers.get('Content-Security-Policy'));
    expectProductionOnlySecureSources(response.headers.get('Content-Security-Policy'));
    expect(response.deleted).toEqual(
      expect.arrayContaining(['csl_access_token', 'csl_refresh_token', 'csl_session']),
    );
  });

  it('redirects non-admin users away from admin routes without clearing live auth', () => {
    const userToken = createAccessToken({ role: 'user' });

    const response = proxy(createRequest('/admin', userToken) as never) as unknown as {
      headers: Headers;
      url: string;
      deleted: string[];
    };

    expect(response.url).toBe('https://lavoval.com/account');
    expect(response.headers.get('Content-Security-Policy')).toContain("default-src 'self'");
    expect(response.deleted).toEqual([]);
  });

  it('allows valid admin access through', () => {
    const adminToken = createAccessToken();

    const response = proxy(createRequest('/admin', adminToken) as never) as unknown as {
      headers: Headers;
      type: string;
    };

    expect(response.type).toBe('next');
    expect(response.headers.get('X-Content-Type-Options')).toBe('nosniff');
    expect(response.headers.get('X-Frame-Options')).toBe('DENY');
    expectNonceBasedScriptPolicy(response.headers.get('Content-Security-Policy'));
    expectProductionOnlySecureSources(response.headers.get('Content-Security-Policy'));
    expect(nextMock).toHaveBeenCalled();
  });

  it('passes nonce and CSP through request headers for Next.js rendering', () => {
    proxy(createRequest('/skills') as never);

    expect(nextMock).toHaveBeenCalledWith({
      request: {
        headers: expect.any(Headers),
      },
    });

    const nextArgs = nextMock.mock.calls.at(-1)?.[0] as { request: { headers: Headers } };
    const nonce = nextArgs.request.headers.get('x-nonce');
    const contentSecurityPolicy = nextArgs.request.headers.get('Content-Security-Policy');

    expect(nonce).toBeTruthy();
    expect(contentSecurityPolicy).toContain(`'nonce-${nonce}'`);
    expectNonceBasedScriptPolicy(contentSecurityPolicy);
    expectProductionOnlySecureSources(contentSecurityPolicy);
  });

  it('matches public routes so security headers are present outside gated areas', () => {
    expect(config.matcher).toEqual(['/((?!_next/static|_next/image|favicon.ico).*)']);
  });
});
