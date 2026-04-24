CREATE TABLE IF NOT EXISTS lavoval_mfa_recovery_codes (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES lavoval_users(id) ON DELETE CASCADE,
  code_hash TEXT NOT NULL UNIQUE,
  consumed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS lavoval_idx_mfa_recovery_codes_user_active
  ON lavoval_mfa_recovery_codes(user_id, created_at DESC)
  WHERE consumed_at IS NULL;

CREATE TABLE IF NOT EXISTS lavoval_auth_sign_in_challenges (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES lavoval_users(id) ON DELETE CASCADE,
  expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS lavoval_idx_auth_sign_in_challenges_user_active
  ON lavoval_auth_sign_in_challenges(user_id, created_at DESC)
  WHERE consumed_at IS NULL;
