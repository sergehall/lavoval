ALTER TABLE mail_jobs
  ADD COLUMN IF NOT EXISTS idempotency_key TEXT,
  ADD COLUMN IF NOT EXISTS provider_message_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mail_jobs_active_idempotency
  ON mail_jobs(idempotency_key)
  WHERE idempotency_key IS NOT NULL AND status IN ('queued', 'retrying', 'processing');

CREATE INDEX IF NOT EXISTS idx_mail_jobs_status_error_code
  ON mail_jobs(status, last_error_code);

CREATE INDEX IF NOT EXISTS idx_mail_jobs_provider_sent_at
  ON mail_jobs(provider, sent_at DESC);
