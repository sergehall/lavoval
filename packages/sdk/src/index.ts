export const apiPaths = {
  auth: {
    login: () => '/api/v1/auth/login',
    register: () => '/api/v1/auth/register',
    logout: () => '/api/v1/auth/logout',
  },
  me: {
    profile: () => '/api/v1/me',
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
