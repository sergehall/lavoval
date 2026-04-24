CREATE TABLE IF NOT EXISTS lavoval_password_reset_tokens (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES lavoval_users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS lavoval_idx_password_reset_tokens_user_active
  ON lavoval_password_reset_tokens(user_id, created_at DESC)
  WHERE consumed_at IS NULL;

CREATE INDEX IF NOT EXISTS lavoval_idx_password_reset_tokens_active_lookup
  ON lavoval_password_reset_tokens(token_hash, expires_at)
  WHERE consumed_at IS NULL;
