import 'server-only';

import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import { ApiClientError, createApiClient } from '@lavoval/sdk';
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

const apiClient = createApiClient({
  baseUrl: env.apiUrl,
  fetchFn: fetch,
});

function mapApiError(error: unknown): never {
  if (error instanceof ApiClientError) {
    throw new ApiError(error.message, error.status);
  }

  throw error;
}

export async function login(payload: LoginRequest) {
  try {
    return await apiClient.auth.login(payload);
  } catch (error) {
    mapApiError(error);
  }
}

export async function register(payload: RegisterRequest) {
  try {
    return await apiClient.auth.register(payload);
  } catch (error) {
    mapApiError(error);
  }
}

export async function logout(token: string) {
  try {
    return await apiClient.auth.logout({ token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchProfile(token: string) {
  try {
    return await apiClient.me.profile({ token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function updateProfile(token: string, payload: ProfileUpdateRequest) {
  try {
    return await apiClient.me.updateProfile(payload, { token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchSkills(token?: string) {
  try {
    return await apiClient.skills.list(token ? { token } : undefined);
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchMySkills(token: string) {
  try {
    return await apiClient.me.skills({ token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchSkillById(id: string, token?: string) {
  try {
    return await apiClient.skills.detail(id, token ? { token } : undefined);
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchMySkillById(token: string, id: string) {
  try {
    return await apiClient.me.skill(id, { token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAdminSkills(token: string) {
  try {
    return await apiClient.admin.skills({ token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAdminSkillById(token: string, id: string) {
  try {
    return await apiClient.admin.skill(id, { token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function createAdminSkill(token: string, payload: SkillMutationRequest) {
  try {
    return await apiClient.admin.createSkill(payload, { token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function createMySkill(token: string, payload: SkillMutationRequest) {
  try {
    return await apiClient.me.createSkill(payload, { token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function updateAdminSkill(token: string, id: string, payload: SkillMutationRequest) {
  try {
    return await apiClient.admin.updateSkill(id, payload, { token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function updateMySkill(token: string, id: string, payload: SkillMutationRequest) {
  try {
    return await apiClient.me.updateSkill(id, payload, { token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function deleteAdminSkill(token: string, id: string) {
  try {
    return await apiClient.admin.deleteSkill(id, { token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function deleteMySkill(token: string, id: string) {
  try {
    return await apiClient.me.deleteSkill(id, { token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAdminUsers(token: string) {
  try {
    return (await apiClient.admin.users({ token })) as ApiEnvelope<UsersListItem[]>;
  } catch (error) {
    mapApiError(error);
  }
}

export async function runSkill(token: string, payload: RuntimeRunRequest) {
  try {
    return await apiClient.runtime.run(payload, { token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchSkillRuns(token: string) {
  try {
    return await apiClient.runtime.runs({ token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchSkillRunById(token: string, runID: string) {
  try {
    return await apiClient.runtime.runDetail(runID, { token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAdminRuns(token: string) {
  try {
    return await apiClient.admin.runs({ token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAdminRunById(token: string, runID: string) {
  try {
    return await apiClient.admin.runDetail(runID, { token });
  } catch (error) {
    mapApiError(error);
  }
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
