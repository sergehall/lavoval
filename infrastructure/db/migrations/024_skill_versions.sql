-- skill versioning

create table skill_versions (
  id uuid primary key default gen_random_uuid(),
  skill_id uuid not null references skills(id) on delete cascade,
  version_no int not null,
  is_current boolean not null default false,

  changelog text,
  content_md text not null default '',
  prompt_template text,
  system_instructions text,

  input_schema_json jsonb,
  output_schema_json jsonb,
  config_schema_json jsonb,
  example_input_json jsonb,
  example_output_json jsonb,

  created_by uuid not null references users(id) on delete restrict,
  created_at timestamptz not null default now(),

  unique(skill_id, version_no)
);

create index idx_skill_versions_skill_id on skill_versions(skill_id);
create index idx_skill_versions_is_current on skill_versions(skill_id, is_current) where is_current = true;
