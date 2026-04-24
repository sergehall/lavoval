ALTER TABLE lavoval_mail_jobs
  ADD COLUMN IF NOT EXISTS idempotency_key TEXT,
  ADD COLUMN IF NOT EXISTS provider_message_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS lavoval_idx_mail_jobs_active_idempotency
  ON lavoval_mail_jobs(idempotency_key)
  WHERE idempotency_key IS NOT NULL AND status IN ('queued', 'retrying', 'processing');

CREATE INDEX IF NOT EXISTS lavoval_idx_mail_jobs_status_error_code
  ON lavoval_mail_jobs(status, last_error_code);

CREATE INDEX IF NOT EXISTS lavoval_idx_mail_jobs_provider_sent_at
  ON lavoval_mail_jobs(provider, sent_at DESC);
