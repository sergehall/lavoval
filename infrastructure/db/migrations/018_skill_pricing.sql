-- Pricing fields on skills
ALTER TABLE skills
  ADD COLUMN IF NOT EXISTS price_cents  INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS currency     CHAR(3) NOT NULL DEFAULT 'USD',
  ADD COLUMN IF NOT EXISTS access_type  TEXT    NOT NULL DEFAULT 'free'
    CHECK (access_type IN ('free', 'paid', 'invite_only'));

-- Explicit access grants (for paid / invite-only skills)
CREATE TABLE IF NOT EXISTS skill_access (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  skill_id    UUID        NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  user_id     UUID        NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
  access_type TEXT        NOT NULL CHECK (access_type IN ('owner', 'purchased', 'granted', 'invite')),
  granted_by  UUID        REFERENCES users(id),
  expires_at  TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (skill_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_skill_access_user  ON skill_access(user_id);
CREATE INDEX IF NOT EXISTS idx_skill_access_skill ON skill_access(skill_id);
