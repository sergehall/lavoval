CREATE TABLE IF NOT EXISTS lavoval_oauth_states (
  id UUID PRIMARY KEY,
  provider TEXT NOT NULL CHECK (provider IN ('google', 'github')),
  state_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS lavoval_idx_oauth_states_provider_expires_at
  ON lavoval_oauth_states (provider, expires_at);

CREATE TABLE IF NOT EXISTS lavoval_oauth_identities (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES lavoval_users(id) ON DELETE CASCADE,
  provider TEXT NOT NULL CHECK (provider IN ('google', 'github')),
  provider_user_id TEXT NOT NULL,
  email CITEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (provider, provider_user_id),
  UNIQUE (user_id, provider)
);

CREATE INDEX IF NOT EXISTS lavoval_idx_oauth_identities_user_id
  ON lavoval_oauth_identities (user_id);
