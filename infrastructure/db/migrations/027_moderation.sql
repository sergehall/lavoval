-- moderation_reports, moderation_actions

create table moderation_reports (
  id uuid primary key default gen_random_uuid(),
  entity_type text not null, -- skill/review/profile/run
  entity_id uuid not null,
  reporter_user_id uuid references users(id) on delete set null,
  reason text not null,
  details text,
  status text not null default 'open', -- open/reviewing/resolved/rejected
  created_at timestamptz not null default now(),
  resolved_at timestamptz
);

create index idx_moderation_reports_entity on moderation_reports(entity_type, entity_id);
create index idx_moderation_reports_status on moderation_reports(status);

create table moderation_actions (
  id uuid primary key default gen_random_uuid(),
  entity_type text not null,
  entity_id uuid not null,
  admin_user_id uuid not null references users(id) on delete restrict,
  action_type text not null, -- hide/archive/warn/delete/restore
  reason text,
  created_at timestamptz not null default now()
);

create index idx_moderation_actions_entity on moderation_actions(entity_type, entity_id);
