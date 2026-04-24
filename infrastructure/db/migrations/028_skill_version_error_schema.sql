alter table lavoval_skill_versions
  add column if not exists error_schema_json jsonb;
