import 'server-only';

import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import { ApiClientError, createApiClient } from '@lavoval/sdk';
import type {
  AdminUserUpdateRequest,
  AccountSecuritySummary,
  AuthResponse,
  EnrollmentAssignRequest,
  EnrollmentDetail,
  EnrollmentUpdateRequest,
  ModuleMutationRequest,
  GitHubOAuthCompleteRequest,
  GoogleOAuthCompleteRequest,
  MFACompleteSignInRequest,
  MFADisableRequest,
  MFAEnrollResponse,
  MFARegenerateRecoveryCodesRequest,
  MFAStatusResponse,
  MFAVerifyEnrollmentRequest,
  ForgotPasswordRequest,
  LoginRequest,
  Profile,
  ProfileUpdateRequest,
  RegisterRequest,
  ResendVerificationRequest,
  ResetPasswordRequest,
  SkillDetail,
  SkillModule,
  SkillMutationRequest,
  SkillSummary,
  VerifyEmailRequest,
} from '@lavoval/contracts';
import type { RuntimeRunRequest, SkillRun } from '@lavoval/contracts/runtime';
import { env } from '@/shared/config/env';
import { signInHref } from '@/shared/lib/auth-navigation';
import type { ApiEnvelope, SessionState, UsersListItem, UserWithProfile } from './types';

const ACCESS_COOKIE = 'csl_access_token';
const REFRESH_COOKIE = 'csl_refresh_token';
const SESSION_COOKIE = 'csl_session';

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly code?: string,
    public readonly meta?: Record<string, unknown>,
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
    throw new ApiError(error.message, error.status, error.code, error.meta);
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

export async function completeMFASignIn(payload: MFACompleteSignInRequest) {
  try {
    return await apiClient.auth.mfaCompleteSignIn(payload);
  } catch (error) {
    mapApiError(error);
  }
}

export async function completeGoogleOAuth(payload: GoogleOAuthCompleteRequest) {
  try {
    return await apiClient.auth.completeGoogleOAuth(payload);
  } catch (error) {
    mapApiError(error);
  }
}

export async function completeGitHubOAuth(payload: GitHubOAuthCompleteRequest) {
  try {
    return await apiClient.auth.completeGitHubOAuth(payload);
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

export async function verifyEmail(payload: VerifyEmailRequest) {
  try {
    return await apiClient.auth.verifyEmail(payload);
  } catch (error) {
    mapApiError(error);
  }
}

export async function resendVerification(payload: ResendVerificationRequest) {
  try {
    return await apiClient.auth.resendVerification(payload);
  } catch (error) {
    mapApiError(error);
  }
}

export async function forgotPassword(payload: ForgotPasswordRequest) {
  try {
    return await apiClient.auth.forgotPassword(payload);
  } catch (error) {
    mapApiError(error);
  }
}

export async function resetPassword(payload: ResetPasswordRequest) {
  try {
    return await apiClient.auth.resetPassword(payload);
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchMFAStatus(token: string) {
  try {
    return await apiClient.auth.mfaStatus({ token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function enrollMFA(token: string) {
  try {
    return await apiClient.auth.mfaEnroll({ token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function cancelMFAEnrollment(token: string) {
  try {
    return await apiClient.auth.mfaCancelEnrollment({ token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function verifyMFAEnrollment(token: string, payload: MFAVerifyEnrollmentRequest) {
  try {
    return await apiClient.auth.mfaVerifyEnrollment(payload, { token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function disableMFA(token: string, payload: MFADisableRequest) {
  try {
    return await apiClient.auth.mfaDisable(payload, { token });
  } catch (error) {
    mapApiError(error);
  }
}

export async function regenerateMFARecoveryCodes(
  token: string,
  payload: MFARegenerateRecoveryCodesRequest,
) {
  try {
    return await apiClient.auth.mfaRegenerateRecoveryCodes(payload, { token });
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

export async function fetchAccountSecurity(token: string) {
  try {
    return await apiClient.me.security({ token });
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

export async function fetchAdminModules(token: string, skillID: string) {
  try {
    return (await apiClient.admin.modules(skillID, { token })) as ApiEnvelope<SkillModule[]>;
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

export async function createAdminModule(
  token: string,
  skillID: string,
  payload: ModuleMutationRequest,
) {
  try {
    return (await apiClient.admin.createModule(skillID, payload, { token })) as ApiEnvelope<SkillModule>;
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

export async function updateAdminModule(
  token: string,
  _skillID: string,
  moduleID: string,
  payload: ModuleMutationRequest,
) {
  try {
    return (await apiClient.admin.updateModule(moduleID, payload, { token })) as ApiEnvelope<SkillModule>;
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

export async function deleteAdminModule(token: string, _skillID: string, moduleID: string) {
  try {
    return (await apiClient.admin.deleteModule(moduleID, { token })) as ApiEnvelope<{ success: boolean }>;
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

export async function fetchAdminUser(token: string, id: string) {
  try {
    return (await apiClient.admin.user(id, { token })) as ApiEnvelope<UserWithProfile>;
  } catch (error) {
    mapApiError(error);
  }
}

export async function updateAdminUser(token: string, id: string, payload: AdminUserUpdateRequest) {
  try {
    return (await apiClient.admin.updateUser(id, payload, { token })) as ApiEnvelope<UsersListItem>;
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAdminEnrollments(token: string) {
  try {
    return (await apiClient.admin.enrollments({ token })) as ApiEnvelope<EnrollmentDetail[]>;
  } catch (error) {
    mapApiError(error);
  }
}

export async function assignSkillToUser(token: string, payload: EnrollmentAssignRequest) {
  try {
    return (await apiClient.admin.assignSkill(payload, { token })) as ApiEnvelope<EnrollmentDetail>;
  } catch (error) {
    mapApiError(error);
  }
}

export async function updateAdminEnrollment(
  token: string,
  id: string,
  payload: EnrollmentUpdateRequest,
) {
  try {
    return (await apiClient.admin.updateEnrollment(id, payload, {
      token,
    })) as ApiEnvelope<EnrollmentDetail>;
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
    redirect(signInHref);
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

async function clearSessionAndRedirectToLogin(): Promise<never> {
  await clearSession();
  redirect(signInHref);
}

export async function withValidSession<T>(handler: (session: SessionState) => Promise<T>) {
  const session = await requireSession();

  try {
    return await handler(session);
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      await clearSessionAndRedirectToLogin();
    }

    throw error;
  }
}
