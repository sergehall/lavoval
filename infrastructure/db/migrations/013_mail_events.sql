CREATE TABLE IF NOT EXISTS lavoval_mail_events (
  id UUID PRIMARY KEY,
  job_id UUID NOT NULL REFERENCES lavoval_mail_jobs(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL,
  message_type TEXT NOT NULL,
  provider TEXT,
  provider_message_id TEXT,
  recipient_email CITEXT NOT NULL,
  error_code TEXT,
  attempt INT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS lavoval_idx_mail_events_job_created
  ON lavoval_mail_events(job_id, created_at DESC);

CREATE INDEX IF NOT EXISTS lavoval_idx_mail_events_type_created
  ON lavoval_mail_events(event_type, created_at DESC);

CREATE INDEX IF NOT EXISTS lavoval_idx_mail_events_error_created
  ON lavoval_mail_events(error_code, created_at DESC);
