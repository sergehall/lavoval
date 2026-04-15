CREATE TABLE IF NOT EXISTS mail_jobs (
  id UUID PRIMARY KEY,
  message_type TEXT NOT NULL,
  recipient_email CITEXT NOT NULL,
  payload JSONB NOT NULL,
  status TEXT NOT NULL,
  attempts INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 4,
  next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  leased_until TIMESTAMPTZ,
  last_error TEXT,
  last_error_code TEXT,
  provider TEXT,
  sent_at TIMESTAMPTZ,
  dead_lettered_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mail_jobs_dispatch_ready
  ON mail_jobs(status, next_attempt_at, leased_until, created_at);

CREATE INDEX IF NOT EXISTS idx_mail_jobs_recipient_created
  ON mail_jobs(recipient_email, created_at DESC);
