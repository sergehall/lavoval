import 'server-only';

import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import { apiPaths } from '@lavoval/sdk';
import type {
  AuthResponse,
  LoginRequest,
  Profile,
  ProfileUpdateRequest,
  RegisterRequest,
  SkillDetail,
  SkillMutationRequest,
  SkillSummary,
} from '@lavoval/contracts';
import type { RuntimeRunRequest, SkillRun } from '@lavoval/contracts/runtime';
import { env } from '@/shared/config/env';
import type { ApiEnvelope, SessionState, UsersListItem } from './types';

const ACCESS_COOKIE = 'csl_access_token';
const REFRESH_COOKIE = 'csl_refresh_token';
const SESSION_COOKIE = 'csl_session';

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${env.apiUrl}${path}`, {
    ...init,
    headers: {
      'content-type': 'application/json',
      ...(init?.headers ?? {}),
    },
    cache: 'no-store',
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => ({ error: { message: 'Unknown error' } }));
    throw new ApiError(payload?.error?.message ?? 'Request failed', response.status);
  }

  return response.json() as Promise<T>;
}

export async function login(payload: LoginRequest) {
  return request<ApiEnvelope<AuthResponse>>(apiPaths.auth.login(), {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function register(payload: RegisterRequest) {
  return request<ApiEnvelope<AuthResponse>>(apiPaths.auth.register(), {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export async function logout(token: string) {
  return request<ApiEnvelope<{ success: boolean }>>(apiPaths.auth.logout(), {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
}

export async function fetchProfile(token: string) {
  return request<ApiEnvelope<Profile>>(apiPaths.me.profile(), {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function updateProfile(token: string, payload: ProfileUpdateRequest) {
  return request<ApiEnvelope<Profile>>(apiPaths.me.updateProfile(), {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  });
}

export async function fetchSkills(token?: string) {
  return request<ApiEnvelope<SkillSummary[]>>(apiPaths.skills.list(), {
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  });
}

export async function fetchMySkills(token: string) {
  return request<ApiEnvelope<SkillSummary[]>>(apiPaths.me.skills(), {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function fetchSkillById(id: string, token?: string) {
  return request<ApiEnvelope<SkillDetail>>(apiPaths.skills.detail(id), {
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  });
}

export async function fetchMySkillById(token: string, id: string) {
  return request<ApiEnvelope<SkillDetail>>(apiPaths.me.skill(id), {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function fetchAdminSkills(token: string) {
  return request<ApiEnvelope<SkillSummary[]>>(apiPaths.admin.skills(), {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function fetchAdminSkillById(token: string, id: string) {
  return request<ApiEnvelope<SkillDetail>>(apiPaths.admin.skill(id), {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function createAdminSkill(token: string, payload: SkillMutationRequest) {
  return request<ApiEnvelope<SkillDetail>>(apiPaths.admin.skills(), {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  });
}

export async function createMySkill(token: string, payload: SkillMutationRequest) {
  return request<ApiEnvelope<SkillDetail>>(apiPaths.me.skills(), {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  });
}

export async function updateAdminSkill(token: string, id: string, payload: SkillMutationRequest) {
  return request<ApiEnvelope<SkillDetail>>(apiPaths.admin.skill(id), {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  });
}

export async function updateMySkill(token: string, id: string, payload: SkillMutationRequest) {
  return request<ApiEnvelope<SkillDetail>>(apiPaths.me.skill(id), {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  });
}

export async function deleteAdminSkill(token: string, id: string) {
  return request<ApiEnvelope<{ success: boolean }>>(apiPaths.admin.skill(id), {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function deleteMySkill(token: string, id: string) {
  return request<ApiEnvelope<{ success: boolean }>>(apiPaths.me.skill(id), {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function fetchAdminUsers(token: string) {
  return request<ApiEnvelope<UsersListItem[]>>(apiPaths.admin.users(), {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function runSkill(token: string, payload: RuntimeRunRequest) {
  return request<ApiEnvelope<SkillRun>>(apiPaths.runtime.run(), {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  });
}

export async function fetchSkillRuns(token: string) {
  return request<ApiEnvelope<SkillRun[]>>(apiPaths.runtime.runs(), {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function fetchSkillRunById(token: string, runID: string) {
  return request<ApiEnvelope<SkillRun>>(apiPaths.runtime.runDetail(runID), {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function fetchAdminRuns(token: string) {
  return request<ApiEnvelope<SkillRun[]>>(apiPaths.admin.runs(), {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function fetchAdminRunById(token: string, runID: string) {
  return request<ApiEnvelope<SkillRun>>(apiPaths.admin.runDetail(runID), {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export async function persistSession(authResponse: AuthResponse) {
  const store = await cookies();
  const session: SessionState = {
    user: authResponse.user,
    accessToken: authResponse.accessToken,
    refreshToken: authResponse.refreshToken,
  };

  const cookieOptions = {
    httpOnly: true,
    sameSite: 'lax' as const,
    secure: false,
    path: '/',
  };

  store.set(ACCESS_COOKIE, authResponse.accessToken, cookieOptions);
  store.set(REFRESH_COOKIE, authResponse.refreshToken, cookieOptions);
  store.set(SESSION_COOKIE, JSON.stringify(session), cookieOptions);
}

export async function clearSession() {
  const store = await cookies();
  store.delete(ACCESS_COOKIE);
  store.delete(REFRESH_COOKIE);
  store.delete(SESSION_COOKIE);
}

export async function getSession() {
  const store = await cookies();
  const session = store.get(SESSION_COOKIE)?.value;
  if (!session) {
    return null;
  }

  try {
    return JSON.parse(session) as SessionState;
  } catch {
    return null;
  }
}

export async function requireSession() {
  const session = await getSession();
  if (!session) {
    redirect('/login');
  }
  return session;
}

export async function requireAdminSession() {
  const session = await requireSession();
  if (session.user.role !== 'admin') {
    redirect('/account');
  }
  return session;
}

export async function withValidSession<T>(handler: (session: SessionState) => Promise<T>) {
  const session = await requireSession();

  try {
    return await handler(session);
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      redirect('/login');
    }

    throw error;
  }
}
