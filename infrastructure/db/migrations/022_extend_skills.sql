-- extend lavoval_skills table with marketplace fields

alter table lavoval_skills
  add column category_id uuid references lavoval_categories(id) on delete set null,
  add column subcategory_id uuid references lavoval_subcategories(id) on delete set null,
  add column skill_type text not null default 'workflow',   -- guide/workflow/prompt-pack/agent-ready/service
  add column difficulty text not null default 'middle',     -- junior/middle/senior
  add column cover_url text,
  add column icon_url text,
  add column is_agent_ready boolean not null default false,
  add column recommended_agent_id uuid,                     -- filled after lavoval_agents table added
  add column estimated_time_minutes int,
  add column language_code text not null default 'en',
  add column success_rate numeric(5,2) not null default 0,
  add column avg_rating numeric(3,2) not null default 0,
  add column runs_count int not null default 0,
  add column saves_count int not null default 0,
  add column forks_count int not null default 0,
  add column published_at timestamptz;

create index lavoval_idx_skills_category_id on lavoval_skills(category_id);
create index lavoval_idx_skills_subcategory_id on lavoval_skills(subcategory_id);
create index lavoval_idx_skills_skill_type on lavoval_skills(skill_type);
create index lavoval_idx_skills_difficulty on lavoval_skills(difficulty);
create index lavoval_idx_skills_is_agent_ready on lavoval_skills(is_agent_ready);
create index lavoval_idx_skills_avg_rating on lavoval_skills(avg_rating desc);
create index lavoval_idx_skills_runs_count on lavoval_skills(runs_count desc);
