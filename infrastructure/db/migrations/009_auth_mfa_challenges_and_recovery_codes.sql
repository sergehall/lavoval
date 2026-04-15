CREATE TABLE IF NOT EXISTS mfa_recovery_codes (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  code_hash TEXT NOT NULL UNIQUE,
  consumed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mfa_recovery_codes_user_active
  ON mfa_recovery_codes(user_id, created_at DESC)
  WHERE consumed_at IS NULL;

CREATE TABLE IF NOT EXISTS auth_sign_in_challenges (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_sign_in_challenges_user_active
  ON auth_sign_in_challenges(user_id, created_at DESC)
  WHERE consumed_at IS NULL;
