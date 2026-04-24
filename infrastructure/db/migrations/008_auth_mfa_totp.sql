ALTER TABLE lavoval_users
  ADD COLUMN IF NOT EXISTS mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS mfa_totp_secret_encrypted TEXT,
  ADD COLUMN IF NOT EXISTS mfa_pending_totp_secret_encrypted TEXT,
  ADD COLUMN IF NOT EXISTS mfa_enrolled_at TIMESTAMPTZ;

ALTER TABLE lavoval_users
  DROP CONSTRAINT IF EXISTS chk_users_mfa_enabled_requires_secret;

ALTER TABLE lavoval_users
  ADD CONSTRAINT chk_users_mfa_enabled_requires_secret
  CHECK (NOT mfa_enabled OR mfa_totp_secret_encrypted IS NOT NULL);
