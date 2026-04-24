INSERT INTO lavoval_users (id, email, password_hash, role, status)
VALUES (
  '93af53e2-0c31-4a38-a5fa-bc1c3ff6fb7f',
  'admin@lavoval.local',
  '$2a$10$h8y5nbQzNziiTL0KWs/o2e.YN9x6mJQm3RWdltcGbGj9LIcYsEMJm',
  'root_owner',
  'active'
)
ON CONFLICT (email) DO NOTHING;

INSERT INTO lavoval_profiles (user_id, first_name, last_name, bio, timezone)
VALUES (
  '93af53e2-0c31-4a38-a5fa-bc1c3ff6fb7f',
  'Platform',
  'Admin',
  'Default administrative account for local development.',
  'UTC'
)
ON CONFLICT (user_id) DO NOTHING;
