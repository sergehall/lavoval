CREATE TABLE IF NOT EXISTS lavoval_skill_runs (
  id UUID PRIMARY KEY,
  skill_id UUID NOT NULL REFERENCES lavoval_skills(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES lavoval_users(id) ON DELETE CASCADE,
  status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'failed')),
  input_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  output_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  error_message TEXT,
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS lavoval_idx_skill_runs_user_created_at
  ON lavoval_skill_runs(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS lavoval_idx_skill_runs_skill_created_at
  ON lavoval_skill_runs(skill_id, created_at DESC);
