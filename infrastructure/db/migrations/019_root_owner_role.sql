ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
  ADD CONSTRAINT users_role_check
  CHECK (role IN ('user', 'admin', 'root_owner'));

UPDATE users
SET role = 'root_owner', updated_at = NOW()
WHERE email = 'admin@lavoval.local' AND deleted_at IS NULL;
