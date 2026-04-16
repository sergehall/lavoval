-- Extend skill status enum with governance values
DO $$
DECLARE rec RECORD;
BEGIN
  FOR rec IN
    SELECT c.conname
    FROM pg_constraint c
    JOIN pg_class t ON t.oid = c.conrelid
    JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(c.conkey)
    WHERE t.relname = 'skills'
      AND c.contype = 'c'
      AND a.attname = 'status'
  LOOP
    EXECUTE format('ALTER TABLE skills DROP CONSTRAINT %I', rec.conname);
  END LOOP;
END $$;

ALTER TABLE skills
  ADD CONSTRAINT skills_status_check
  CHECK (status IN ('draft', 'pending_review', 'published', 'hidden', 'archived', 'rejected'));

-- Moderation fields
ALTER TABLE skills
  ADD COLUMN IF NOT EXISTS moderation_reason TEXT,
  ADD COLUMN IF NOT EXISTS moderated_by      UUID REFERENCES users(id),
  ADD COLUMN IF NOT EXISTS moderated_at      TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS is_featured       BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS is_verified       BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_skills_featured ON skills(is_featured) WHERE deleted_at IS NULL AND is_featured = TRUE;
