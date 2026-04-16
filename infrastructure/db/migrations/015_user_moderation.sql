-- Drop the existing check constraint on users.status so we can add 'blocked'
DO $$
DECLARE rec RECORD;
BEGIN
  FOR rec IN
    SELECT c.conname
    FROM pg_constraint c
    JOIN pg_class t ON t.oid = c.conrelid
    JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(c.conkey)
    WHERE t.relname = 'users'
      AND c.contype = 'c'
      AND a.attname = 'status'
  LOOP
    EXECUTE format('ALTER TABLE users DROP CONSTRAINT %I', rec.conname);
  END LOOP;
END $$;

ALTER TABLE users
  ADD CONSTRAINT users_status_check
  CHECK (status IN ('active', 'invited', 'suspended', 'blocked'));

-- Suspension audit fields
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS suspension_reason TEXT,
  ADD COLUMN IF NOT EXISTS suspended_at      TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS suspended_by      UUID REFERENCES users(id),
  ADD COLUMN IF NOT EXISTS block_reason      TEXT,
  ADD COLUMN IF NOT EXISTS blocked_at        TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS blocked_by        UUID REFERENCES users(id);
