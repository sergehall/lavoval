import type { NextRequest } from 'next/server';
import { NextResponse } from 'next/server';
import { signInHref } from '@/shared/lib/auth-navigation';

const sessionCookieName = 'csl_session';

function readRole(request: NextRequest) {
  const session = request.cookies.get(sessionCookieName)?.value;
  if (!session) {
    return null;
  }

  try {
    const parsed = JSON.parse(session) as { user?: { role?: string } };
    return parsed.user?.role ?? null;
  } catch {
    return null;
  }
}

export function proxy(request: NextRequest) {
  const role = readRole(request);
  const { pathname } = request.nextUrl;

  if (pathname.startsWith('/account') && !role) {
    return NextResponse.redirect(new URL(signInHref, request.url));
  }

  if (pathname.startsWith('/admin') && role !== 'admin') {
    return NextResponse.redirect(new URL(role ? '/account' : signInHref, request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/account/:path*', '/admin/:path*'],
};
