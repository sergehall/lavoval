CREATE TABLE IF NOT EXISTS mail_cleanup_runs (
    id UUID PRIMARY KEY,
    mode TEXT NOT NULL CHECK (mode IN ('apply', 'dry_run', 'manual')),
    status TEXT NOT NULL CHECK (status IN ('success', 'failed')),
    dry_run BOOLEAN NOT NULL DEFAULT FALSE,
    candidate_jobs BIGINT NOT NULL DEFAULT 0,
    candidate_events BIGINT NOT NULL DEFAULT 0,
    deleted_jobs BIGINT NOT NULL DEFAULT 0,
    deleted_events BIGINT NOT NULL DEFAULT 0,
    error_message TEXT,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mail_cleanup_runs_created_at
    ON mail_cleanup_runs (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_mail_cleanup_runs_status_created_at
    ON mail_cleanup_runs (status, created_at DESC);
