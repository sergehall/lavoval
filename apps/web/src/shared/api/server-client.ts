import 'server-only';

import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import { ApiClientError, createApiClient } from '@lavoval/sdk';
import type {
  AdminAuditLog,
  AdminUserRoleUpdateRequest,
  AdminUserStatusUpdateRequest,
  AdminSkillGovernanceRequest,
  AdminSkillPricingRequest,
  AdminStats,
  AdminUserUpdateRequest,
  AccountSecuritySummary,
  AuthResponse,
  Category,
  Subcategory,
  Tag,
  Agent,
  SkillAgentCompatibility,
  SkillReview,
  Collection,
  CollectionInput,
  CreateReviewInput,
  RunFeedbackInput,
  SkillFilterParams,
  EnrollmentAssignRequest,
  EnrollmentDetail,
  EnrollmentUpdateRequest,
  ModuleMutationRequest,
  PublicProfile,
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
export type { Category, Subcategory, Tag, Agent, SkillAgentCompatibility, SkillReview, Collection };
import type { RuntimeRunRequest, SkillRun } from '@lavoval/contracts/runtime';
import { env } from '@/shared/config/env';
import {
  ACCESS_COOKIE,
  getSessionFromCookies,
  REFRESH_COOKIE,
  SESSION_COOKIE,
} from '@/shared/lib/auth-session';
import type { SessionState } from '@/shared/lib/auth-session';
import { signInHref } from '@/shared/lib/auth-navigation';
import { canAccessAdmin } from '@/shared/lib/rbac';
import type {
  ApiEnvelope,
  AdminMailEventFilter,
  AdminMailJobFilter,
  MailEvent,
  MailCleanupResult,
  MailCleanupRun,
  MailJob,
  MailOperationalSnapshot,
  MailRetentionSnapshot,
  MailSuppression,
  UsersListItem,
  UserWithProfile,
} from './types';
export type { AdminAuditLog, AdminStats };
export type { SessionState } from '@/shared/lib/auth-session';

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

async function fetchPublicJson<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${env.apiUrl}${path}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      ...(init?.headers ?? {}),
    },
    cache: 'no-store',
  });

  let payload: unknown = null;
  try {
    payload = await response.json();
  } catch {
    payload = null;
  }

  if (!response.ok) {
    const errorPayload = payload as { error?: { code?: string; message?: string } } | null;
    throw new ApiError(
      errorPayload?.error?.message ?? `Request failed with status ${response.status}`,
      response.status,
      errorPayload?.error?.code,
    );
  }

  return payload as T;
}

async function fetchAdminJson<T>(token: string, path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${env.apiUrl}${path}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      Authorization: `Bearer ${token}`,
      ...(init?.headers ?? {}),
    },
    cache: 'no-store',
  });

  let payload: unknown = null;
  try {
    payload = await response.json();
  } catch {
    payload = null;
  }

  if (!response.ok) {
    const errorPayload = payload as { error?: { code?: string; message?: string } } | null;
    throw new ApiError(
      errorPayload?.error?.message ?? `Request failed with status ${response.status}`,
      response.status,
      errorPayload?.error?.code,
    );
  }

  return payload as T;
}

function withSearchParams(path: string, params: Record<string, string | number | undefined>) {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value == null || value === '') continue;
    query.set(key, String(value));
  }

  const encoded = query.toString();
  if (!encoded) {
    return path;
  }
  return `${path}?${encoded}`;
}

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

export async function fetchPublicCreatorProfile(creatorID: string) {
  try {
    return await fetchPublicJson<{ data: PublicProfile }>(`/api/v1/creators/${creatorID}`);
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchSkills(filter?: SkillFilterParams, token?: string) {
  const params: Record<string, string | number | undefined> = {};
  if (filter?.q) params.q = filter.q;
  if (filter?.category) params.category = filter.category;
  if (filter?.subcategory) params.subcategory = filter.subcategory;
  if (filter?.tags) params.tags = filter.tags;
  if (filter?.difficulty) params.difficulty = filter.difficulty;
  if (filter?.skillType) params.skillType = filter.skillType;
  if (filter?.sort) params.sort = filter.sort;
  if (filter?.agentReady != null) params.agentReady = String(filter.agentReady);

  const path = withSearchParams('/api/v1/skills', params);

  try {
    return await fetchPublicJson<{ data: SkillSummary[] }>(path);
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchCategories() {
  try {
    return await fetchPublicJson<Category[]>('/api/v1/categories');
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchSubcategories(categoryID: string) {
  try {
    return await fetchPublicJson<Subcategory[]>(`/api/v1/categories/${categoryID}/subcategories`);
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchTags() {
  try {
    return await fetchPublicJson<Tag[]>('/api/v1/tags');
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAgents() {
  try {
    return await fetchPublicJson<Agent[]>('/api/v1/agents');
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAgentBySlug(slug: string) {
  try {
    return await fetchPublicJson<Agent>(`/api/v1/agents/${slug}`);
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchSkillReviews(skillID: string) {
  try {
    return await fetchPublicJson<SkillReview[]>(`/api/v1/skills/${skillID}/reviews`);
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchRecommendedAgents(skillID: string) {
  try {
    return await fetchPublicJson<SkillAgentCompatibility[]>(
      `/api/v1/skills/${skillID}/recommended-agents`,
    );
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchMyCollections(token: string) {
  try {
    return await fetchAdminJson<Collection[]>(token, '/api/v1/me/collections');
  } catch (error) {
    mapApiError(error);
  }
}

export async function createCollection(token: string, payload: CollectionInput) {
  try {
    return await fetchAdminJson<Collection>(token, '/api/v1/me/collections', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
  } catch (error) {
    mapApiError(error);
  }
}

export async function saveSkill(token: string, skillID: string) {
  try {
    return await fetchAdminJson<void>(token, `/api/v1/skills/${skillID}/save`, { method: 'POST' });
  } catch (error) {
    mapApiError(error);
  }
}

export async function unsaveSkill(token: string, skillID: string) {
  try {
    return await fetchAdminJson<void>(token, `/api/v1/skills/${skillID}/save`, {
      method: 'DELETE',
    });
  } catch (error) {
    mapApiError(error);
  }
}

export async function createReview(token: string, skillID: string, payload: CreateReviewInput) {
  try {
    return await fetchAdminJson<SkillReview>(token, `/api/v1/skills/${skillID}/reviews`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
  } catch (error) {
    mapApiError(error);
  }
}

export async function submitRunFeedback(token: string, runID: string, payload: RunFeedbackInput) {
  try {
    return await fetchAdminJson<void>(token, `/api/v1/runtime/runs/${runID}/feedback`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
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
    return (await apiClient.admin.createModule(skillID, payload, {
      token,
    })) as ApiEnvelope<SkillModule>;
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
    return (await apiClient.admin.updateModule(moduleID, payload, {
      token,
    })) as ApiEnvelope<SkillModule>;
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
    return (await apiClient.admin.deleteModule(moduleID, { token })) as ApiEnvelope<{
      success: boolean;
    }>;
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

export async function updateAdminUserStatus(
  token: string,
  id: string,
  payload: AdminUserStatusUpdateRequest,
) {
  try {
    return (await apiClient.admin.updateUserStatus(id, payload, {
      token,
    })) as ApiEnvelope<UsersListItem>;
  } catch (error) {
    mapApiError(error);
  }
}

export async function updateAdminUserRole(
  token: string,
  id: string,
  payload: AdminUserRoleUpdateRequest,
) {
  try {
    return (await apiClient.admin.updateUserRole(id, payload, {
      token,
    })) as ApiEnvelope<UsersListItem>;
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

export async function fetchAdminMailOperations(token: string) {
  try {
    return await fetchAdminJson<MailOperationalSnapshot>(token, '/api/v1/admin/mail/ops');
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAdminMailRetention(token: string) {
  try {
    return await fetchAdminJson<MailRetentionSnapshot>(token, '/api/v1/admin/mail/retention');
  } catch (error) {
    mapApiError(error);
  }
}

export async function cleanupAdminMailRetention(token: string) {
  try {
    return await fetchAdminJson<MailCleanupResult>(token, '/api/v1/admin/mail/cleanup', {
      method: 'POST',
    });
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAdminMailCleanupRuns(token: string, limit = 20) {
  try {
    return await fetchAdminJson<MailCleanupRun[]>(
      token,
      withSearchParams('/api/v1/admin/mail/cleanup-runs', { limit }),
    );
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAdminDeadLetters(token: string, filter: AdminMailJobFilter = {}) {
  try {
    return await fetchAdminJson<MailJob[]>(
      token,
      withSearchParams('/api/v1/admin/mail/dead-letters', filter),
    );
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAdminMailEvents(token: string, filter: AdminMailEventFilter = {}) {
  try {
    return await fetchAdminJson<MailEvent[]>(
      token,
      withSearchParams('/api/v1/admin/mail/events', filter),
    );
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAdminMailJobEvents(
  token: string,
  jobID: string,
  filter: AdminMailEventFilter = {},
) {
  try {
    return await fetchAdminJson<MailEvent[]>(
      token,
      withSearchParams(`/api/v1/admin/mail/jobs/${jobID}/events`, filter),
    );
  } catch (error) {
    mapApiError(error);
  }
}

export async function requeueAdminDeadLetter(token: string, jobID: string) {
  try {
    return await fetchAdminJson<MailJob>(
      token,
      `/api/v1/admin/mail/dead-letters/${jobID}/requeue`,
      {
        method: 'POST',
      },
    );
  } catch (error) {
    mapApiError(error);
  }
}

export async function replayAdminMailJob(token: string, jobID: string) {
  try {
    return await fetchAdminJson<MailJob>(token, `/api/v1/admin/mail/jobs/${jobID}/replay`, {
      method: 'POST',
    });
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAdminMailSuppressions(token: string) {
  try {
    return await fetchAdminJson<MailSuppression[]>(token, '/api/v1/admin/mail/suppressions');
  } catch (error) {
    mapApiError(error);
  }
}

export async function createAdminMailSuppression(
  token: string,
  payload: { kind: 'email' | 'domain'; value: string; reason: string },
) {
  try {
    return await fetchAdminJson<MailSuppression>(token, '/api/v1/admin/mail/suppressions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
  } catch (error) {
    mapApiError(error);
  }
}

export async function deleteAdminMailSuppression(token: string, suppressionID: string) {
  try {
    return await fetchAdminJson<{ success: boolean }>(
      token,
      `/api/v1/admin/mail/suppressions/${suppressionID}`,
      { method: 'DELETE' },
    );
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchAdminStats(token: string) {
  try {
    return await fetchAdminJson<AdminStats>(token, '/api/v1/admin/stats');
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchUserAuditLog(token: string, userID: string) {
  try {
    return await fetchAdminJson<AdminAuditLog[]>(token, `/api/v1/admin/users/${userID}/audit`);
  } catch (error) {
    mapApiError(error);
  }
}

export async function governAdminSkill(
  token: string,
  skillID: string,
  payload: AdminSkillGovernanceRequest,
) {
  try {
    return await fetchAdminJson<SkillDetail>(token, `/api/v1/admin/skills/${skillID}/governance`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
  } catch (error) {
    mapApiError(error);
  }
}

export async function updateAdminSkillPricing(
  token: string,
  skillID: string,
  payload: AdminSkillPricingRequest,
) {
  try {
    return await fetchAdminJson<SkillDetail>(token, `/api/v1/admin/skills/${skillID}/pricing`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
  } catch (error) {
    mapApiError(error);
  }
}

export async function fetchSkillAuditLog(token: string, skillID: string) {
  try {
    return await fetchAdminJson<AdminAuditLog[]>(token, `/api/v1/admin/skills/${skillID}/audit`);
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
  return getSessionFromCookies({
    sessionCookie: store.get(SESSION_COOKIE)?.value,
    accessTokenCookie: store.get(ACCESS_COOKIE)?.value,
  });
}

export async function getValidatedSession() {
  const session = await getSession();
  if (!session) {
    return null;
  }

  try {
    const { data: profile } = await fetchProfile(session.accessToken);
    return {
      ...session,
      user: {
        ...session.user,
        role: profile.role,
        firstName: profile.firstName,
        lastName: profile.lastName,
      },
    };
  } catch (error) {
    if (error instanceof ApiError && (error.status === 401 || error.status === 403)) {
      return null;
    }

    throw error;
  }
}

export async function requireSession() {
  const session = await getValidatedSession();
  if (!session) {
    redirect(signInHref);
  }
  return session;
}

export async function requireAdminSession() {
  const session = await requireSession();
  if (!canAccessAdmin(session.user.role)) {
    redirect('/account');
  }
  return session;
}

function redirectToLogin(): never {
  redirect(signInHref);
}

export async function withValidSession<T>(handler: (session: SessionState) => Promise<T>) {
  const session = await requireSession();

  try {
    return await handler(session);
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      redirectToLogin();
    }

    throw error;
  }
}
