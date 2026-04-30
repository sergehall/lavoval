import type { NextRequest } from 'next/server';
import { NextResponse } from 'next/server';
import type { Role } from '@lavoval/contracts';
import {
  ACCESS_COOKIE,
  getRoleFromAccessToken,
  REFRESH_COOKIE,
  SESSION_COOKIE,
} from '@/shared/lib/auth-session';
import { signInHref } from '@/shared/lib/auth-navigation';
import { canAccessAdmin } from '@/shared/lib/rbac';

function createNonce() {
  return Buffer.from(crypto.randomUUID()).toString('base64');
}

function createContentSecurityPolicy(nonce: string) {
  const devScriptSource = process.env.NODE_ENV === 'production' ? '' : " 'unsafe-eval'";

  return [
    "default-src 'self'",
    "base-uri 'self'",
    "connect-src 'self' http://localhost:* http://127.0.0.1:* https://api.lavoval.com ws://localhost:* ws://127.0.0.1:*",
    "font-src 'self' data:",
    "form-action 'self'",
    "frame-ancestors 'none'",
    "img-src 'self' data: blob: https://avatars.githubusercontent.com",
    "object-src 'none'",
    `script-src 'self' 'nonce-${nonce}' 'strict-dynamic'${devScriptSource}`,
    `style-src 'self' 'nonce-${nonce}'`,
  ].join('; ');
}

function applySecurityHeaders(response: NextResponse, contentSecurityPolicy: string) {
  response.headers.set('Content-Security-Policy', contentSecurityPolicy);
  response.headers.set('X-Content-Type-Options', 'nosniff');
  response.headers.set('X-Frame-Options', 'DENY');
  return response;
}

function clearAuthCookies(response: NextResponse) {
  for (const name of [ACCESS_COOKIE, REFRESH_COOKIE, SESSION_COOKIE]) {
    response.cookies.delete(name);
  }
}

export function proxy(request: NextRequest) {
  const role = getRoleFromAccessToken(request.cookies.get(ACCESS_COOKIE)?.value) as Role | null;
  const { pathname } = request.nextUrl;
  const nonce = createNonce();
  const contentSecurityPolicy = createContentSecurityPolicy(nonce);
  const requestHeaders = new Headers(request.headers);
  requestHeaders.set('x-nonce', nonce);
  requestHeaders.set('Content-Security-Policy', contentSecurityPolicy);

  if (pathname.startsWith('/account') && !role) {
    const response = NextResponse.redirect(new URL(signInHref, request.url));
    clearAuthCookies(response);
    return applySecurityHeaders(response, contentSecurityPolicy);
  }

  if (pathname.startsWith('/admin') && !canAccessAdmin(role)) {
    const response = NextResponse.redirect(new URL(role ? '/account' : signInHref, request.url));
    if (!role) {
      clearAuthCookies(response);
    }
    return applySecurityHeaders(response, contentSecurityPolicy);
  }

  return applySecurityHeaders(
    NextResponse.next({
      request: {
        headers: requestHeaders,
      },
    }),
    contentSecurityPolicy,
  );
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
};
