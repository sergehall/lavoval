-- Restored from old Render DB (dpg-d7flmkhkh4rs739rmad0-a)
-- Owner reassigned to: serge.hall.dev@gmail.com (2403431b-d38d-41fc-b1ab-729266934f6e)
-- Category: Programming → fc8a0f36-0dbe-4e15-ad6b-f141c6cedb86

INSERT INTO lavoval_skills (
  id, slug, title, summary, description,
  status, visibility, created_by,
  skill_type, difficulty, is_featured, is_verified,
  access_type, price_cents, category_id, published_at
) VALUES

(
  '11111111-1111-4111-8111-111111111111',
  'typescript-contracts-for-ai-products',
  'Advanced TypeScript Contracts For AI Platforms',
  'Engineer contract-safe TypeScript systems for AI-native products with stronger runtime guarantees, cleaner integration seams, and senior-grade delivery discipline.',
  'Engineer contract-safe TypeScript systems for AI-native products with stronger runtime guarantees, cleaner integration seams, and senior-grade delivery discipline.',
  'published', 'public', '2403431b-d38d-41fc-b1ab-729266934f6e',
  'agent-ready', 'senior', true, true,
  'free', 0, 'fc8a0f36-0dbe-4e15-ad6b-f141c6cedb86', NOW()
),

(
  '22222222-2222-4222-8222-222222222222',
  'go-service-reliability-playbook',
  'Production Go Reliability And Service Architecture',
  'Build Go services that stay reliable under load with explicit lifecycle control, stronger concurrency discipline, better observability, and safer failure handling.',
  'Build Go services that stay reliable under load with explicit lifecycle control, stronger concurrency discipline, better observability, and safer failure handling.',
  'published', 'public', '2403431b-d38d-41fc-b1ab-729266934f6e',
  'workflow', 'senior', true, true,
  'free', 0, 'fc8a0f36-0dbe-4e15-ad6b-f141c6cedb86', NOW()
),

(
  '33333333-3333-4333-8333-333333333333',
  'nextjs-nestjs-platform-architecture',
  'Next.js And NestJS Platform Architecture For Scale',
  'Design a more scalable Next.js + NestJS platform with stricter contracts, cleaner frontend-backend seams, and architecture that can survive real product growth.',
  'Design a more scalable Next.js + NestJS platform with stricter contracts, cleaner frontend-backend seams, and architecture that can survive real product growth.',
  'published', 'public', '2403431b-d38d-41fc-b1ab-729266934f6e',
  'service', 'senior', true, true,
  'free', 0, 'fc8a0f36-0dbe-4e15-ad6b-f141c6cedb86', NOW()
),

(
  '77777777-7777-4777-8777-777777777777',
  'react-performance-delivery-systems',
  'React Performance Systems For Product Teams',
  'Push React applications toward senior-level product performance by fixing rendering costs, state boundaries, and delivery bottlenecks that compound over time.',
  'Push React applications toward senior-level product performance by fixing rendering costs, state boundaries, and delivery bottlenecks that compound over time.',
  'published', 'public', '2403431b-d38d-41fc-b1ab-729266934f6e',
  'guide', 'senior', true, true,
  'free', 0, 'fc8a0f36-0dbe-4e15-ad6b-f141c6cedb86', NOW()
),

(
  '88888888-8888-4888-8888-888888888888',
  'nestjs-messaging-and-workflows',
  'NestJS Messaging, Queues, And Workflow Design',
  'Design message-driven NestJS systems with clearer queue ownership, safer retries, better contracts, and operational visibility teams can trust in production.',
  'Design message-driven NestJS systems with clearer queue ownership, safer retries, better contracts, and operational visibility teams can trust in production.',
  'published', 'public', '2403431b-d38d-41fc-b1ab-729266934f6e',
  'workflow', 'senior', true, true,
  'free', 0, 'fc8a0f36-0dbe-4e15-ad6b-f141c6cedb86', NOW()
),

(
  '99999999-9999-4999-8999-999999999999',
  'ai-product-discovery-sprint',
  'AI Product Discovery And Execution Strategy',
  'Turn ambitious AI ideas into sharper product bets, tighter technical scopes, and execution plans grounded in workflows, constraints, and delivery reality.',
  'Turn ambitious AI ideas into sharper product bets, tighter technical scopes, and execution plans grounded in workflows, constraints, and delivery reality.',
  'published', 'public', '2403431b-d38d-41fc-b1ab-729266934f6e',
  'service', 'senior', true, true,
  'free', 0, 'fc8a0f36-0dbe-4e15-ad6b-f141c6cedb86', NOW()
)

ON CONFLICT (slug) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- Modules (restored verbatim)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO lavoval_skill_modules (id, skill_id, slug, title, summary, content, position, status) VALUES

-- TypeScript Contracts
('11111111-aaaa-4111-8111-111111111111', '11111111-1111-4111-8111-111111111111',
 'schema-first-design', 'Schema-first design',
 'Model safe request and response contracts before wiring UI or services.',
 'Start with DTOs, schemas, and explicit payload ownership so integration drift becomes visible early.',
 0, 'published'),
('11111111-bbbb-4111-8111-111111111111', '11111111-1111-4111-8111-111111111111',
 'runtime-validation', 'Runtime validation',
 'Use runtime parsing to protect boundaries where TypeScript alone cannot help.',
 'Validate form input, query params, webhooks, and API responses so production behavior stays trustworthy.',
 1, 'published'),

-- Go Reliability
('22222222-aaaa-4222-8222-222222222222', '22222222-2222-4222-8222-222222222222',
 'lifecycle-and-context', 'Lifecycle and context',
 'Design goroutines and request flow around ownership, cancellation, and shutdown.',
 'Every goroutine should have an owner, every timeout should be intentional, and shutdown should be predictable.',
 0, 'published'),
('22222222-bbbb-4222-8222-222222222222', '22222222-2222-4222-8222-222222222222',
 'bounded-concurrency', 'Bounded concurrency',
 'Use errgroup, semaphores, and worker patterns without losing clarity.',
 'Prefer simple coordination, preserve errors, and avoid invisible background work that becomes impossible to operate.',
 1, 'published'),

-- Next.js + NestJS
('33333333-aaaa-4333-8333-333333333333', '33333333-3333-4333-8333-333333333333',
 'frontend-backend-contracts', 'Frontend-backend contracts',
 'Align App Router data boundaries with validated NestJS DTOs and auth flows.',
 'Keep server-only logic on the server, validate payloads at edges, and prevent accidental coupling across layers.',
 0, 'published'),
('33333333-bbbb-4333-8333-333333333333', '33333333-3333-4333-8333-333333333333',
 'platform-evolution', 'Platform evolution',
 'Grow from a modular monolith toward clearer service boundaries only when they are justified.',
 'Use domain modules, event boundaries, and delivery gates to reduce chaos while the product surface expands.',
 1, 'published'),

-- React Performance
('77777777-aaaa-4777-8777-777777777777', '77777777-7777-4777-8777-777777777777',
 'render-budget', 'Render budget',
 'Find and remove avoidable render churn in real product surfaces.',
 'Start with interaction hotspots, not theory, and use the smallest effective fix.',
 0, 'published'),
('77777777-bbbb-4777-8777-777777777777', '77777777-7777-4777-8777-777777777777',
 'state-placement', 'State placement',
 'Move state closer to where it belongs so the whole tree stops paying for local decisions.',
 'Reduce coupling and re-renders by being more intentional about boundaries.',
 1, 'published'),

-- NestJS Messaging
('88888888-aaaa-4888-8888-888888888888', '88888888-8888-4888-8888-888888888888',
 'events-and-queues', 'Events and queues',
 'Model async workflows so retries and side effects stay visible.',
 'Design queue-driven flows with better ownership and fewer magical service hops.',
 0, 'published'),
('88888888-bbbb-4888-8888-888888888888', '88888888-8888-4888-8888-888888888888',
 'operational-feedback', 'Operational feedback',
 'Add the visibility needed to trust asynchronous delivery in production.',
 'Use logs, alerts, and replay-safe contracts to keep workflows understandable.',
 1, 'published'),

-- AI Product Discovery
('99999999-aaaa-4999-8999-999999999999', '99999999-9999-4999-8999-999999999999',
 'problem-shaping', 'Problem shaping',
 'Convert vague AI excitement into a better user problem statement.',
 'Decide where AI changes workflow value instead of adding novelty only.',
 0, 'published'),
('99999999-bbbb-4999-8999-999999999999', '99999999-9999-4999-8999-999999999999',
 'mvp-scope', 'MVP scope',
 'Reduce delivery risk by turning broad ambition into a narrower, testable first move.',
 'Clarify assumptions, interfaces, and constraints before implementation starts.',
 1, 'published')

ON CONFLICT (skill_id, slug) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- Enrollments
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO lavoval_user_skill_enrollments (user_id, skill_id, status, progress_percent)
VALUES
  ('2403431b-d38d-41fc-b1ab-729266934f6e', '11111111-1111-4111-8111-111111111111', 'in_progress', 0),
  ('2403431b-d38d-41fc-b1ab-729266934f6e', '22222222-2222-4222-8222-222222222222', 'in_progress', 0),
  ('2403431b-d38d-41fc-b1ab-729266934f6e', '33333333-3333-4333-8333-333333333333', 'in_progress', 0),
  ('2403431b-d38d-41fc-b1ab-729266934f6e', '77777777-7777-4777-8777-777777777777', 'in_progress', 0),
  ('2403431b-d38d-41fc-b1ab-729266934f6e', '88888888-8888-4888-8888-888888888888', 'in_progress', 0),
  ('2403431b-d38d-41fc-b1ab-729266934f6e', '99999999-9999-4999-8999-999999999999', 'in_progress', 0)
ON CONFLICT (user_id, skill_id) DO NOTHING;
