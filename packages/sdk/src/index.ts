import type {
  AuthResponse,
  AccountSecuritySummary,
  GitHubOAuthCompleteRequest,
  GoogleOAuthCompleteRequest,
  MFACompleteSignInRequest,
  MFADisableRequest,
  MFAEnrollResponse,
  MFARegenerateRecoveryCodesRequest,
  MFAStatusResponse,
  MFAVerifyEnrollmentRequest,
  ForgotPasswordRequest,
  ForgotPasswordResponse,
  LoginRequest,
  Profile,
  ProfileUpdateRequest,
  RegisterRequest,
  RegisterResponse,
  ResendVerificationRequest,
  ResetPasswordRequest,
  ResetPasswordResponse,
  SessionUser,
  VerificationResponse,
  VerifyEmailRequest,
} from '@lavoval/contracts';
import type { RuntimeRunRequest, SkillRun } from '@lavoval/contracts/runtime';
import type { SkillDetail, SkillMutationRequest, SkillSummary } from '@lavoval/registry';

export const DEFAULT_LAVOVAL_API_URL = 'http://localhost:8080';

export type ApiEnvelope<T> = {
  data: T;
  meta?: Record<string, unknown>;
};

export type ApiClientRequestOptions = {
  token?: string;
  headers?: HeadersInit;
};

export type ApiClientConfig = {
  baseUrl: string;
  fetchFn?: typeof fetch;
  defaultHeaders?: HeadersInit;
};

export type EnvLike = Record<string, string | undefined>;

export class ApiClientError extends Error {
  readonly status: number;
  readonly code?: string;
  readonly meta?: Record<string, unknown>;

  constructor(
    message: string,
    status: number,
    code?: string,
    meta?: Record<string, unknown>,
  ) {
    super(message);
    this.name = 'ApiClientError';
    this.status = status;
    this.code = code;
    this.meta = meta;
  }
}

export function resolveApiBaseUrl(explicit?: string, env: EnvLike = process.env) {
  return explicit ?? env.LAVOVAL_API_URL ?? env.NEXT_PUBLIC_API_URL ?? DEFAULT_LAVOVAL_API_URL;
}

export function resolveAccessToken(explicit?: string, env: EnvLike = process.env) {
  return explicit ?? env.LAVOVAL_ACCESS_TOKEN ?? env.ACCESS_TOKEN;
}

export const apiPaths = {
  auth: {
    login: () => '/api/v1/auth/login',
    register: () => '/api/v1/auth/register',
    verifyEmail: () => '/api/v1/auth/verify-email',
    resendVerification: () => '/api/v1/auth/resend-verification',
    forgotPassword: () => '/api/v1/auth/forgot-password',
    resetPassword: () => '/api/v1/auth/reset-password',
    googleOAuthStart: () => '/api/v1/auth/oauth/google/start',
    githubOAuthStart: () => '/api/v1/auth/oauth/github/start',
    googleOAuthComplete: () => '/api/v1/auth/oauth/google/complete',
    githubOAuthComplete: () => '/api/v1/auth/oauth/github/complete',
    mfaStatus: () => '/api/v1/auth/mfa/status',
    mfaEnroll: () => '/api/v1/auth/mfa/enroll',
    mfaCancelEnrollment: () => '/api/v1/auth/mfa/cancel-enrollment',
    mfaVerifyEnrollment: () => '/api/v1/auth/mfa/verify-enrollment',
    mfaDisable: () => '/api/v1/auth/mfa/disable',
    mfaRegenerateRecoveryCodes: () => '/api/v1/auth/mfa/recovery-codes/regenerate',
    mfaCompleteSignIn: () => '/api/v1/auth/mfa/complete-sign-in',
    logout: () => '/api/v1/auth/logout',
  },
  me: {
    profile: () => '/api/v1/me',
    security: () => '/api/v1/me/security',
    updateProfile: () => '/api/v1/me/profile',
    skills: () => '/api/v1/me/skills',
    skill: (id: string) => `/api/v1/me/skills/${id}`,
  },
  skills: {
    list: () => '/api/v1/skills',
    detail: (id: string) => `/api/v1/skills/${id}`,
  },
  runtime: {
    run: () => '/api/v1/runtime/run',
    runs: () => '/api/v1/runtime/runs',
    runDetail: (id: string) => `/api/v1/runtime/runs/${id}`,
  },
  admin: {
    users: () => '/api/v1/admin/users',
    skills: () => '/api/v1/admin/skills',
    skill: (id: string) => `/api/v1/admin/skills/${id}`,
    runs: () => '/api/v1/admin/runs',
    runDetail: (id: string) => `/api/v1/admin/runs/${id}`,
  },
} as const;

function mergeHeaders(...sets: Array<HeadersInit | undefined>) {
  const headers = new Headers();

  for (const set of sets) {
    if (!set) {
      continue;
    }
    const normalized = new Headers(set);
    for (const [key, value] of normalized.entries()) {
      headers.set(key, value);
    }
  }

  return headers;
}

type UsersListItem = SessionUser & {
  status: string;
  createdAt: string;
};

export function createApiClient(config: ApiClientConfig) {
  const fetchFn = config.fetchFn ?? fetch;

  async function request<T>(path: string, init?: RequestInit, options?: ApiClientRequestOptions) {
    const headers = mergeHeaders(
      { 'content-type': 'application/json' },
      config.defaultHeaders,
      options?.headers,
      init?.headers,
      options?.token ? { Authorization: `Bearer ${options.token}` } : undefined,
    );

    const response = await fetchFn(`${config.baseUrl}${path}`, {
      ...init,
      headers,
      cache: 'no-store',
    });

    if (!response.ok) {
      const payload = await response.json().catch(() => ({ error: { message: 'Unknown error' } }));
      throw new ApiClientError(
        payload?.error?.message ?? 'Request failed',
        response.status,
        payload?.error?.code,
        payload?.error?.meta,
      );
    }

    return response.json() as Promise<ApiEnvelope<T>>;
  }

  return {
    request,
    auth: {
      login(payload: LoginRequest) {
        return request<AuthResponse>(apiPaths.auth.login(), {
          method: 'POST',
          body: JSON.stringify(payload),
        });
      },
      register(payload: RegisterRequest) {
        return request<RegisterResponse>(apiPaths.auth.register(), {
          method: 'POST',
          body: JSON.stringify(payload),
        });
      },
      verifyEmail(payload: VerifyEmailRequest) {
        return request<VerificationResponse>(apiPaths.auth.verifyEmail(), {
          method: 'POST',
          body: JSON.stringify(payload),
        });
      },
      resendVerification(payload: ResendVerificationRequest) {
        return request<RegisterResponse>(apiPaths.auth.resendVerification(), {
          method: 'POST',
          body: JSON.stringify(payload),
        });
      },
      forgotPassword(payload: ForgotPasswordRequest) {
        return request<ForgotPasswordResponse>(apiPaths.auth.forgotPassword(), {
          method: 'POST',
          body: JSON.stringify(payload),
        });
      },
      resetPassword(payload: ResetPasswordRequest) {
        return request<ResetPasswordResponse>(apiPaths.auth.resetPassword(), {
          method: 'POST',
          body: JSON.stringify(payload),
        });
      },
      completeGoogleOAuth(payload: GoogleOAuthCompleteRequest) {
        return request<AuthResponse>(apiPaths.auth.googleOAuthComplete(), {
          method: 'POST',
          body: JSON.stringify(payload),
        });
      },
      completeGitHubOAuth(payload: GitHubOAuthCompleteRequest) {
        return request<AuthResponse>(apiPaths.auth.githubOAuthComplete(), {
          method: 'POST',
          body: JSON.stringify(payload),
        });
      },
      mfaStatus(options: ApiClientRequestOptions) {
        return request<MFAStatusResponse>(apiPaths.auth.mfaStatus(), undefined, options);
      },
      mfaEnroll(options: ApiClientRequestOptions) {
        return request<MFAEnrollResponse>(apiPaths.auth.mfaEnroll(), { method: 'POST' }, options);
      },
      mfaCancelEnrollment(options: ApiClientRequestOptions) {
        return request<MFAStatusResponse>(apiPaths.auth.mfaCancelEnrollment(), { method: 'POST' }, options);
      },
      mfaVerifyEnrollment(payload: MFAVerifyEnrollmentRequest, options: ApiClientRequestOptions) {
        return request<MFAStatusResponse>(apiPaths.auth.mfaVerifyEnrollment(), {
          method: 'POST',
          body: JSON.stringify(payload),
        }, options);
      },
      mfaDisable(payload: MFADisableRequest, options: ApiClientRequestOptions) {
        return request<MFAStatusResponse>(apiPaths.auth.mfaDisable(), {
          method: 'POST',
          body: JSON.stringify(payload),
        }, options);
      },
      mfaRegenerateRecoveryCodes(payload: MFARegenerateRecoveryCodesRequest, options: ApiClientRequestOptions) {
        return request<MFAStatusResponse>(apiPaths.auth.mfaRegenerateRecoveryCodes(), {
          method: 'POST',
          body: JSON.stringify(payload),
        }, options);
      },
      mfaCompleteSignIn(payload: MFACompleteSignInRequest) {
        return request<AuthResponse>(apiPaths.auth.mfaCompleteSignIn(), {
          method: 'POST',
          body: JSON.stringify(payload),
        });
      },
      logout(options: ApiClientRequestOptions) {
        return request<{ success: boolean }>(apiPaths.auth.logout(), { method: 'POST' }, options);
      },
    },
    me: {
      profile(options: ApiClientRequestOptions) {
        return request<Profile>(apiPaths.me.profile(), undefined, options);
      },
      security(options: ApiClientRequestOptions) {
        return request<AccountSecuritySummary>(apiPaths.me.security(), undefined, options);
      },
      updateProfile(payload: ProfileUpdateRequest, options: ApiClientRequestOptions) {
        return request<Profile>(
          apiPaths.me.updateProfile(),
          { method: 'PATCH', body: JSON.stringify(payload) },
          options,
        );
      },
      skills(options: ApiClientRequestOptions) {
        return request<SkillSummary[]>(apiPaths.me.skills(), undefined, options);
      },
      skill(id: string, options: ApiClientRequestOptions) {
        return request<SkillDetail>(apiPaths.me.skill(id), undefined, options);
      },
      createSkill(payload: SkillMutationRequest, options: ApiClientRequestOptions) {
        return request<SkillDetail>(
          apiPaths.me.skills(),
          { method: 'POST', body: JSON.stringify(payload) },
          options,
        );
      },
      updateSkill(id: string, payload: SkillMutationRequest, options: ApiClientRequestOptions) {
        return request<SkillDetail>(
          apiPaths.me.skill(id),
          { method: 'PATCH', body: JSON.stringify(payload) },
          options,
        );
      },
      deleteSkill(id: string, options: ApiClientRequestOptions) {
        return request<{ success: boolean }>(
          apiPaths.me.skill(id),
          { method: 'DELETE' },
          options,
        );
      },
    },
    skills: {
      list(options?: ApiClientRequestOptions) {
        return request<SkillSummary[]>(apiPaths.skills.list(), undefined, options);
      },
      detail(id: string, options?: ApiClientRequestOptions) {
        return request<SkillDetail>(apiPaths.skills.detail(id), undefined, options);
      },
    },
    runtime: {
      run(payload: RuntimeRunRequest, options: ApiClientRequestOptions) {
        return request<SkillRun>(
          apiPaths.runtime.run(),
          { method: 'POST', body: JSON.stringify(payload) },
          options,
        );
      },
      runs(options: ApiClientRequestOptions) {
        return request<SkillRun[]>(apiPaths.runtime.runs(), undefined, options);
      },
      runDetail(id: string, options: ApiClientRequestOptions) {
        return request<SkillRun>(apiPaths.runtime.runDetail(id), undefined, options);
      },
    },
    admin: {
      users(options: ApiClientRequestOptions) {
        return request<UsersListItem[]>(apiPaths.admin.users(), undefined, options);
      },
      skills(options: ApiClientRequestOptions) {
        return request<SkillSummary[]>(apiPaths.admin.skills(), undefined, options);
      },
      skill(id: string, options: ApiClientRequestOptions) {
        return request<SkillDetail>(apiPaths.admin.skill(id), undefined, options);
      },
      createSkill(payload: SkillMutationRequest, options: ApiClientRequestOptions) {
        return request<SkillDetail>(
          apiPaths.admin.skills(),
          { method: 'POST', body: JSON.stringify(payload) },
          options,
        );
      },
      updateSkill(id: string, payload: SkillMutationRequest, options: ApiClientRequestOptions) {
        return request<SkillDetail>(
          apiPaths.admin.skill(id),
          { method: 'PATCH', body: JSON.stringify(payload) },
          options,
        );
      },
      deleteSkill(id: string, options: ApiClientRequestOptions) {
        return request<{ success: boolean }>(
          apiPaths.admin.skill(id),
          { method: 'DELETE' },
          options,
        );
      },
      runs(options: ApiClientRequestOptions) {
        return request<SkillRun[]>(apiPaths.admin.runs(), undefined, options);
      },
      runDetail(id: string, options: ApiClientRequestOptions) {
        return request<SkillRun>(apiPaths.admin.runDetail(id), undefined, options);
      },
    },
  };
}
