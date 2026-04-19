-- extend skill_runs with agent info, token tracking, cost; add run_feedback

alter table skill_runs
  add column agent_id uuid references agents(id) on delete set null,
  add column skill_version_id uuid references skill_versions(id) on delete set null,
  add column normalized_input_text text,
  add column output_text text,
  add column tokens_input int,
  add column tokens_output int,
  add column estimated_cost numeric(12,6);

create index idx_skill_runs_agent_id on skill_runs(agent_id);

create table run_feedback (
  run_id uuid primary key references skill_runs(id) on delete cascade,
  user_id uuid not null references users(id) on delete cascade,
  rating int not null check (rating between 1 and 5),
  usefulness_score int check (usefulness_score between 1 and 5),
  would_use_again boolean,
  comment text,
  created_at timestamptz not null default now()
);
