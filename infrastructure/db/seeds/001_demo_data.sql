INSERT INTO users (id, email, password_hash, role, status)
VALUES (
  'ce43b778-b766-46ae-b619-1bd2a2ddf7eb',
  'user@lavoval.local',
  '$2a$10$h8y5nbQzNziiTL0KWs/o2e.YN9x6mJQm3RWdltcGbGj9LIcYsEMJm',
  'user',
  'active'
)
ON CONFLICT (email) DO NOTHING;

INSERT INTO profiles (user_id, first_name, last_name, bio, timezone)
VALUES (
  'ce43b778-b766-46ae-b619-1bd2a2ddf7eb',
  'Demo',
  'User',
  'Exploring the platform foundations and catalog flows.',
  'UTC'
)
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO skills (id, slug, title, summary, description, status, visibility, created_by)
VALUES (
  '5f9fd5c8-dc0a-4e38-aad4-fa4e3d0eecb8',
  'platform-design-foundations',
  'Platform Design Foundations',
  'Learn to shape maintainable product systems with clear service and UI boundaries.',
  'A practical skill for architects and senior engineers who need to define maintainable foundations, explicit contracts, and scalable domain boundaries.',
  'published',
  'public',
  '93af53e2-0c31-4a38-a5fa-bc1c3ff6fb7f'
)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO skill_modules (id, skill_id, slug, title, summary, content, position, status)
VALUES
  (
    'e02e8215-54bb-4069-a4bb-16b322580918',
    '5f9fd5c8-dc0a-4e38-aad4-fa4e3d0eecb8',
    'domain-boundaries',
    'Domain Boundaries',
    'Identify entities, workflows, and ownership seams before implementation.',
    'Map domain concepts, ownership, and lifecycle states before expanding the implementation surface.',
    0,
    'published'
  ),
  (
    '44f889b3-7997-48f2-8e72-1f77a996fd26',
    '5f9fd5c8-dc0a-4e38-aad4-fa4e3d0eecb8',
    'delivery-ops',
    'Delivery and Operations',
    'Translate platform design into maintainable operating practices and tooling.',
    'Connect architecture with migrations, environments, tests, and operational visibility so the system can scale safely.',
    1,
    'published'
  )
ON CONFLICT (skill_id, slug) DO NOTHING;

INSERT INTO user_skill_enrollments (user_id, skill_id, status, progress_percent)
VALUES ('ce43b778-b766-46ae-b619-1bd2a2ddf7eb', '5f9fd5c8-dc0a-4e38-aad4-fa4e3d0eecb8', 'in_progress', 35)
ON CONFLICT (user_id, skill_id) DO NOTHING;
