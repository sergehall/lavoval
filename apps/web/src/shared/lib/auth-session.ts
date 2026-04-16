import { sessionUserSchema, type Role } from '@lavoval/contracts';
import { z } from 'zod';

const sessionStateSchema = z.object({
  user: sessionUserSchema,
  accessToken: z.string().min(1),
  refreshToken: z.string().min(1),
});

const accessTokenClaimsSchema = z.object({
  uid: z.string().min(1),
  email: z.string().email(),
  role: z.custom<Role>((value) => value === 'user' || value === 'admin' || value === 'root_owner'),
  exp: z.number().int().positive(),
});

export type SessionState = z.infer<typeof sessionStateSchema>;
export type AccessTokenClaims = z.infer<typeof accessTokenClaimsSchema>;

export const ACCESS_COOKIE = 'csl_access_token';
export const REFRESH_COOKIE = 'csl_refresh_token';
export const SESSION_COOKIE = 'csl_session';

function decodeBase64Url(value: string) {
  const normalized = value.replace(/-/g, '+').replace(/_/g, '/');
  const padding = (4 - (normalized.length % 4)) % 4;
  return atob(normalized.padEnd(normalized.length + padding, '='));
}

export function parseSessionCookie(value: string | undefined | null) {
  if (!value) {
    return null;
  }

  try {
    return sessionStateSchema.parse(JSON.parse(value));
  } catch {
    return null;
  }
}

export function parseAccessTokenClaims(token: string | undefined | null) {
  if (!token) {
    return null;
  }

  const parts = token.split('.');
  if (parts.length !== 3) {
    return null;
  }

  try {
    return accessTokenClaimsSchema.parse(JSON.parse(decodeBase64Url(parts[1] ?? '')));
  } catch {
    return null;
  }
}

export function isAccessTokenExpired(claims: AccessTokenClaims, now = Date.now()) {
  return claims.exp * 1000 <= now;
}

export function getSessionFromCookies({
  sessionCookie,
  accessTokenCookie,
  now = Date.now(),
}: {
  sessionCookie: string | undefined | null;
  accessTokenCookie?: string | undefined | null;
  now?: number;
}) {
  const session = parseSessionCookie(sessionCookie);
  if (!session) {
    return null;
  }

  const activeAccessToken = accessTokenCookie ?? session.accessToken;
  if (!activeAccessToken || session.accessToken !== activeAccessToken) {
    return null;
  }

  const claims = parseAccessTokenClaims(activeAccessToken);
  if (!claims || isAccessTokenExpired(claims, now)) {
    return null;
  }

  if (
    claims.uid !== session.user.id ||
    claims.email !== session.user.email ||
    claims.role !== session.user.role
  ) {
    return null;
  }

  return session;
}

export function getRoleFromAccessToken(token: string | undefined | null, now = Date.now()) {
  const claims = parseAccessTokenClaims(token);
  if (!claims || isAccessTokenExpired(claims, now)) {
    return null;
  }

  return claims.role;
}
