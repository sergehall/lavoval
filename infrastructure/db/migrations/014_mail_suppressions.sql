CREATE TABLE IF NOT EXISTS mail_suppressions (
  id UUID PRIMARY KEY,
  kind TEXT NOT NULL CHECK (kind IN ('email', 'domain')),
  value CITEXT NOT NULL,
  reason TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_mail_suppressions_kind_value
  ON mail_suppressions(kind, value);
