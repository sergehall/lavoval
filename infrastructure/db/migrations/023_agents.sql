-- agents registry and skill-agent compatibility

create table agents (
  id uuid primary key default gen_random_uuid(),
  slug text not null unique,
  name text not null,
  provider text not null,       -- openai/anthropic/local/custom
  model_name text not null,
  description text,

  supports_text boolean not null default true,
  supports_code boolean not null default false,
  supports_tools boolean not null default false,
  supports_web boolean not null default false,
  supports_files boolean not null default false,
  supports_multimodal boolean not null default false,
  supports_json_output boolean not null default false,
  max_context_tokens int,

  pricing_json jsonb,
  limits_json jsonb,

  status text not null default 'active', -- active/deprecated/hidden
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table skill_agent_compatibility (
  skill_id uuid not null references skills(id) on delete cascade,
  agent_id uuid not null references agents(id) on delete cascade,

  compatibility_score numeric(5,2) not null default 0,
  success_rate numeric(5,2) not null default 0,
  avg_rating numeric(3,2) not null default 0,
  runs_count int not null default 0,

  tested_by_system boolean not null default false,
  tested_by_users boolean not null default false,
  notes text,

  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),

  primary key (skill_id, agent_id)
);

create index idx_skill_agent_compat_agent_id on skill_agent_compatibility(agent_id);

-- add FK from skills to agents (recommended_agent_id added in 022)
alter table skills
  add constraint fk_skills_recommended_agent
  foreign key (recommended_agent_id) references agents(id) on delete set null;

-- seed well-known agents
insert into agents (slug, name, provider, model_name, description,
  supports_code, supports_tools, supports_json_output, max_context_tokens) values
  ('claude-opus-4',    'Claude Opus 4',       'anthropic', 'claude-opus-4-5',            'Most capable Anthropic model for complex reasoning',   true, true, true, 200000),
  ('claude-sonnet-4',  'Claude Sonnet 4.6',   'anthropic', 'claude-sonnet-4-6',          'Balanced Anthropic model — fast and capable',          true, true, true, 200000),
  ('claude-haiku-4',   'Claude Haiku 4.5',    'anthropic', 'claude-haiku-4-5-20251001',  'Fastest and most affordable Anthropic model',          true, true, true, 200000),
  ('gpt-4o',           'GPT-4o',              'openai',    'gpt-4o',                     'OpenAI flagship multimodal model',                     true, true, true, 128000),
  ('gpt-4o-mini',      'GPT-4o Mini',         'openai',    'gpt-4o-mini',                'Fast and affordable OpenAI model',                     true, true, true, 128000),
  ('gemini-2-flash',   'Gemini 2.0 Flash',    'google',    'gemini-2.0-flash',           'Fast Google model with tool use',                      true, true, true, 1048576);
