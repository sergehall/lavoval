-- Owner: serge.hall.dev@gmail.com
-- id:    2403431b-d38d-41fc-b1ab-729266934f6e

-- ─────────────────────────────────────────────────────────────────────────────
-- Skills
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO lavoval_skills (
  id, slug, title, summary, description,
  status, visibility, created_by,
  skill_type, difficulty, is_featured,
  category_id, published_at
) VALUES

-- 1. Prompt Engineering
(
  'b1000001-0000-0000-0000-000000000001',
  'prompt-engineering-mastery',
  'Prompt Engineering Mastery',
  'Learn to write precise, effective prompts for any LLM — from GPT to Claude.',
  'A hands-on skill covering the full spectrum of prompt engineering: anatomy of a good prompt, chain-of-thought reasoning, system instructions, few-shot patterns, and iterative refinement techniques.',
  'published', 'public', '2403431b-d38d-41fc-b1ab-729266934f6e',
  'workflow', 'beginner', true,
  'fc8a0f36-0dbe-4e15-ad6b-f141c6cedb86', NOW()
),

-- 2. AI Code Review
(
  'b1000002-0000-0000-0000-000000000002',
  'ai-powered-code-review',
  'AI-Powered Code Review',
  'Set up automated, AI-assisted code review workflows that catch real issues.',
  'Learn to integrate AI review tools into your pull request pipeline, define custom rules, reduce noise, and focus AI attention on security, performance, and correctness — not style.',
  'published', 'public', '2403431b-d38d-41fc-b1ab-729266934f6e',
  'workflow', 'middle', false,
  'fc8a0f36-0dbe-4e15-ad6b-f141c6cedb86', NOW()
),

-- 3. Python Data Pipeline
(
  'b1000003-0000-0000-0000-000000000003',
  'python-data-pipeline-ai',
  'Data Pipelines with Python and AI',
  'Build reliable ETL pipelines and automate data cleaning with AI assistance.',
  'A practical skill for data engineers and analysts: design extraction and transformation pipelines in Python, use AI to detect anomalies and fix dirty data, and ship automated reports to stakeholders.',
  'published', 'public', '2403431b-d38d-41fc-b1ab-729266934f6e',
  'workflow', 'middle', false,
  'b6f1d113-e516-4b00-b517-90a73f3fa94a', NOW()
),

-- 4. Workflow Automation
(
  'b1000004-0000-0000-0000-000000000004',
  'workflow-automation-ai',
  'Workflow Automation with AI',
  'Connect your tools and eliminate repetitive tasks using AI-powered automation.',
  'Map your manual workflows, connect services via API, and embed AI decision points using platforms like Make, Zapier, or n8n. Includes error handling, monitoring, and rollback strategies.',
  'published', 'public', '2403431b-d38d-41fc-b1ab-729266934f6e',
  'workflow', 'beginner', true,
  '4b16fc9a-e0ca-4fb1-86aa-05f63326e8bd', NOW()
),

-- 5. Technical Writing
(
  'b1000005-0000-0000-0000-000000000005',
  'technical-writing-for-developers',
  'Technical Writing for Developers',
  'Write documentation that developers actually read — clear, structured, and maintainable.',
  'Covers documentation-as-code philosophy, writing READMEs that convert, structuring API references, generating docs from code with AI, and maintaining docs alongside a fast-moving codebase.',
  'published', 'public', '2403431b-d38d-41fc-b1ab-729266934f6e',
  'workflow', 'beginner', false,
  '61e866b2-e21d-44d4-b614-dee99113f530', NOW()
),

-- 6. Product Discovery with AI
(
  'b1000006-0000-0000-0000-000000000006',
  'product-discovery-ai',
  'Product Discovery with AI',
  'Use AI to accelerate user research, synthesize feedback, and prioritize what to build.',
  'A product skill for PMs and founders: automate interview synthesis, cluster user feedback, run AI-assisted prioritization frameworks, and generate insight reports — all without losing the human signal.',
  'published', 'public', '2403431b-d38d-41fc-b1ab-729266934f6e',
  'workflow', 'middle', false,
  '7f7a384e-83a9-41e0-8478-b0a113098d4c', NOW()
)

ON CONFLICT (slug) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- Modules
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO lavoval_skill_modules (id, skill_id, slug, title, summary, content, position, status) VALUES

-- Prompt Engineering modules
('c1000001-0001-0000-0000-000000000001', 'b1000001-0000-0000-0000-000000000001',
 'anatomy-of-a-prompt', 'Anatomy of a Good Prompt',
 'Understand the building blocks: role, context, instruction, format, and constraints.',
 'Every effective prompt has five components. Learn to identify and tune each one independently to get predictable, high-quality outputs from any model.',
 0, 'published'),
('c1000001-0002-0000-0000-000000000001', 'b1000001-0000-0000-0000-000000000001',
 'chain-of-thought', 'Chain-of-Thought and Reasoning Techniques',
 'Make models show their work — and get dramatically better answers.',
 'CoT prompting, zero-shot vs few-shot, self-consistency, and tree-of-thought: when to use each and how to combine them for complex reasoning tasks.',
 1, 'published'),
('c1000001-0003-0000-0000-000000000001', 'b1000001-0000-0000-0000-000000000001',
 'system-prompt-design', 'System Prompt Design',
 'Write system instructions that shape model behavior reliably across sessions.',
 'Persona, tone, guardrails, and output format — define them in the system prompt so every conversation starts from the right baseline without repeating yourself.',
 2, 'published'),

-- AI Code Review modules
('c1000002-0001-0000-0000-000000000002', 'b1000002-0000-0000-0000-000000000002',
 'ai-review-setup', 'Setting Up AI Review Workflows',
 'Integrate AI reviewers into your GitHub or GitLab pull request pipeline.',
 'Step-by-step: connect an AI review tool, configure repository rules, and decide which checks run automatically vs. on demand.',
 0, 'published'),
('c1000002-0002-0000-0000-000000000002', 'b1000002-0000-0000-0000-000000000002',
 'writing-review-rules', 'Writing Effective Review Rules',
 'Define custom rules that catch security issues, performance bugs, and anti-patterns.',
 'Move beyond generic linting. Write focused AI instructions that reflect your team conventions and the classes of bugs that actually ship to production.',
 1, 'published'),
('c1000002-0003-0000-0000-000000000002', 'b1000002-0000-0000-0000-000000000002',
 'ci-integration', 'Integrating with CI/CD',
 'Gate deployments on AI review results without slowing down the pipeline.',
 'Configure pass/fail thresholds, handle flaky AI responses, and set up async review so the pipeline does not block on slow model calls.',
 2, 'published'),

-- Python Data Pipeline modules
('c1000003-0001-0000-0000-000000000003', 'b1000003-0000-0000-0000-000000000003',
 'etl-fundamentals', 'ETL Fundamentals',
 'Design extraction, transformation, and loading pipelines that are easy to maintain.',
 'Idempotency, schema evolution, partitioning, and retry logic — the principles every reliable pipeline must follow before adding any AI layer.',
 0, 'published'),
('c1000003-0002-0000-0000-000000000003', 'b1000003-0000-0000-0000-000000000003',
 'ai-data-cleaning', 'AI-Assisted Data Cleaning',
 'Use LLMs to detect anomalies, normalize messy fields, and classify ambiguous rows.',
 'Practical patterns for calling AI APIs inside a pandas or Polars pipeline, batching requests efficiently, and logging every AI decision for auditability.',
 1, 'published'),
('c1000003-0003-0000-0000-000000000003', 'b1000003-0000-0000-0000-000000000003',
 'automated-reporting', 'Automating Reports',
 'Generate and deliver stakeholder reports from pipeline outputs — on schedule, zero manual work.',
 'Combine data summaries with an AI narrative layer, render to PDF or Markdown, and push to Slack or email automatically after each pipeline run.',
 2, 'published'),

-- Workflow Automation modules
('c1000004-0001-0000-0000-000000000004', 'b1000004-0000-0000-0000-000000000004',
 'workflow-mapping', 'Workflow Mapping',
 'Identify and document the manual steps worth automating before writing a single integration.',
 'Use a simple trigger-action-condition model to map existing workflows, estimate ROI, and decide which tools to connect first.',
 0, 'published'),
('c1000004-0002-0000-0000-000000000004', 'b1000004-0000-0000-0000-000000000004',
 'connecting-ai-apis', 'Connecting AI APIs',
 'Add AI decision points to your automation: classify, summarize, route, or generate content.',
 'HTTP request nodes, authentication patterns, prompt templates inside automations, and handling variable-length AI responses safely.',
 1, 'published'),
('c1000004-0003-0000-0000-000000000004', 'b1000004-0000-0000-0000-000000000004',
 'error-handling-monitoring', 'Error Handling and Monitoring',
 'Build automations that survive API failures, timeouts, and unexpected data shapes.',
 'Retry logic, dead-letter queues, alerting on failure streaks, and a simple dashboard to keep your automations healthy in production.',
 2, 'published'),

-- Technical Writing modules
('c1000005-0001-0000-0000-000000000005', 'b1000005-0000-0000-0000-000000000005',
 'docs-as-code', 'Documentation as Code',
 'Store, review, and deploy docs the same way you deploy software.',
 'Markdown + Git + CI: set up a docs pipeline where pull requests include documentation changes and broken links fail the build automatically.',
 0, 'published'),
('c1000005-0002-0000-0000-000000000005', 'b1000005-0000-0000-0000-000000000005',
 'api-reference-writing', 'Writing API References',
 'Structure endpoint documentation that answers the questions developers actually ask.',
 'Cover authentication, request/response shapes, error codes, and rate limits. Use AI to generate first drafts from code comments and OpenAPI specs.',
 1, 'published'),
('c1000005-0003-0000-0000-000000000005', 'b1000005-0000-0000-0000-000000000005',
 'readme-that-converts', 'READMEs That Convert',
 'Write a README that turns visitors into users in under 60 seconds.',
 'Hook, install, run, trust — the four sections every open-source README needs. Includes templates and before/after rewrites with AI assistance.',
 2, 'published'),

-- Product Discovery modules
('c1000006-0001-0000-0000-000000000006', 'b1000006-0000-0000-0000-000000000006',
 'user-research-automation', 'Automating User Research',
 'Set up AI pipelines that process interviews, surveys, and support tickets at scale.',
 'Transcribe interviews, extract quotes, tag themes, and surface patterns — automatically, so you spend time on insight rather than cleanup.',
 0, 'published'),
('c1000006-0002-0000-0000-000000000006', 'b1000006-0000-0000-0000-000000000006',
 'synthesizing-feedback', 'Synthesizing Feedback',
 'Cluster and rank user signals from multiple sources into a single prioritized view.',
 'Combine App Store reviews, Intercom tickets, NPS comments, and sales call notes into one AI-generated digest your whole team can act on.',
 1, 'published'),
('c1000006-0003-0000-0000-000000000006', 'b1000006-0000-0000-0000-000000000006',
 'prioritization-frameworks', 'AI-Assisted Prioritization',
 'Score and rank features using RICE, ICE, or custom frameworks — augmented by AI.',
 'Feed your backlog into an AI model with your scoring rubric and get a ranked list with reasoning. Learn to validate and override AI scores with domain knowledge.',
 2, 'published')

ON CONFLICT (skill_id, slug) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- Enrollments (owner enrolled in all skills)
-- ─────────────────────────────────────────────────────────────────────────────

INSERT INTO lavoval_user_skill_enrollments (user_id, skill_id, status, progress_percent)
VALUES
  ('2403431b-d38d-41fc-b1ab-729266934f6e', 'b1000001-0000-0000-0000-000000000001', 'in_progress', 60),
  ('2403431b-d38d-41fc-b1ab-729266934f6e', 'b1000002-0000-0000-0000-000000000002', 'in_progress', 30),
  ('2403431b-d38d-41fc-b1ab-729266934f6e', 'b1000003-0000-0000-0000-000000000003', 'in_progress', 10),
  ('2403431b-d38d-41fc-b1ab-729266934f6e', 'b1000004-0000-0000-0000-000000000004', 'completed',   100),
  ('2403431b-d38d-41fc-b1ab-729266934f6e', 'b1000005-0000-0000-0000-000000000005', 'in_progress', 45),
  ('2403431b-d38d-41fc-b1ab-729266934f6e', 'b1000006-0000-0000-0000-000000000006', 'in_progress', 20)
ON CONFLICT (user_id, skill_id) DO NOTHING;
