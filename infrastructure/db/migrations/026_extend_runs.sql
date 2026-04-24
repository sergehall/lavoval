-- extend lavoval_skill_runs with agent info, token tracking, cost; add lavoval_run_feedback

alter table lavoval_skill_runs
  add column agent_id uuid references lavoval_agents(id) on delete set null,
  add column skill_version_id uuid references lavoval_skill_versions(id) on delete set null,
  add column normalized_input_text text,
  add column output_text text,
  add column tokens_input int,
  add column tokens_output int,
  add column estimated_cost numeric(12,6);

create index lavoval_idx_skill_runs_agent_id on lavoval_skill_runs(agent_id);

create table lavoval_run_feedback (
  run_id uuid primary key references lavoval_skill_runs(id) on delete cascade,
  user_id uuid not null references lavoval_users(id) on delete cascade,
  rating int not null check (rating between 1 and 5),
  usefulness_score int check (usefulness_score between 1 and 5),
  would_use_again boolean,
  comment text,
  created_at timestamptz not null default now()
);
