import type {
  AuthResponse,
  Profile,
  SkillDetail,
  SkillSummary,
  SessionUser,
} from '@lavoval/contracts';
import type { SkillRun } from '@lavoval/contracts/runtime';

export type ApiEnvelope<T> = {
  data: T;
  meta?: Record<string, unknown>;
};

export type AuthTokens = Pick<AuthResponse, 'accessToken' | 'refreshToken'>;
export type SessionState = {
  user: SessionUser;
  accessToken: string;
  refreshToken: string;
};

export type DashboardSummary = {
  profile: Profile;
  skills: SkillSummary[];
};

export type UsersListItem = SessionUser & {
  status: string;
  createdAt: string;
};

export type UserWithProfile = {
  user: UsersListItem;
  profile: import('@lavoval/contracts').Profile;
};

export type SkillPayload = SkillDetail;
export type SkillRunPayload = SkillRun;
