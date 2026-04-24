CREATE TABLE IF NOT EXISTS lavoval_admin_audit_logs (
  id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  entity_type    TEXT        NOT NULL,
  entity_id      UUID        NOT NULL,
  action         TEXT        NOT NULL,
  old_value_json JSONB,
  new_value_json JSONB,
  reason         TEXT,
  actor_id       UUID        NOT NULL REFERENCES lavoval_users(id),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS lavoval_idx_audit_logs_entity   ON lavoval_admin_audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS lavoval_idx_audit_logs_actor    ON lavoval_admin_audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS lavoval_idx_audit_logs_created  ON lavoval_admin_audit_logs(created_at DESC);
