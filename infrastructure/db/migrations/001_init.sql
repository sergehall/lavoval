CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE IF NOT EXISTS lavoval_users (
  id UUID PRIMARY KEY,
  email CITEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('user', 'admin', 'root_owner')),
  status TEXT NOT NULL CHECK (status IN ('active', 'invited', 'suspended')),
  session_version INTEGER NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS lavoval_profiles (
  user_id UUID PRIMARY KEY REFERENCES lavoval_users(id) ON DELETE CASCADE,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  bio TEXT,
  timezone TEXT NOT NULL DEFAULT 'UTC',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS lavoval_skills (
  id UUID PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  summary TEXT NOT NULL,
  description TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('draft', 'published', 'archived')),
  visibility TEXT NOT NULL CHECK (visibility IN ('public', 'private')),
  created_by UUID NOT NULL REFERENCES lavoval_users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS lavoval_skill_modules (
  id UUID PRIMARY KEY,
  skill_id UUID NOT NULL REFERENCES lavoval_skills(id) ON DELETE CASCADE,
  slug TEXT NOT NULL,
  title TEXT NOT NULL,
  summary TEXT NOT NULL,
  content TEXT NOT NULL,
  position INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL CHECK (status IN ('draft', 'published', 'archived')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  UNIQUE (skill_id, slug)
);

CREATE TABLE IF NOT EXISTS lavoval_user_skill_enrollments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES lavoval_users(id) ON DELETE CASCADE,
  skill_id UUID NOT NULL REFERENCES lavoval_skills(id) ON DELETE CASCADE,
  status TEXT NOT NULL CHECK (status IN ('assigned', 'in_progress', 'completed')),
  progress_percent INTEGER NOT NULL DEFAULT 0 CHECK (progress_percent >= 0 AND progress_percent <= 100),
  assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  completed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, skill_id)
);

CREATE INDEX IF NOT EXISTS lavoval_idx_users_role ON lavoval_users(role) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS lavoval_idx_skills_status ON lavoval_skills(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS lavoval_idx_skill_modules_skill_position ON lavoval_skill_modules(skill_id, position) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS lavoval_idx_enrollments_user_status ON lavoval_user_skill_enrollments(user_id, status);
