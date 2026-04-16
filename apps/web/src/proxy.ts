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

function clearAuthCookies(response: NextResponse) {
  for (const name of [ACCESS_COOKIE, REFRESH_COOKIE, SESSION_COOKIE]) {
    response.cookies.delete(name);
  }
}

export function proxy(request: NextRequest) {
  const role = getRoleFromAccessToken(request.cookies.get(ACCESS_COOKIE)?.value) as Role | null;
  const { pathname } = request.nextUrl;

  if (pathname.startsWith('/account') && !role) {
    const response = NextResponse.redirect(new URL(signInHref, request.url));
    clearAuthCookies(response);
    return response;
  }

  if (pathname.startsWith('/admin') && !canAccessAdmin(role)) {
    const response = NextResponse.redirect(new URL(role ? '/account' : signInHref, request.url));
    if (!role) {
      clearAuthCookies(response);
    }
    return response;
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/account/:path*', '/admin/:path*'],
};
