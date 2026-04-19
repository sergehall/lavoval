-- skill_reviews, saved_skills, collections, collection_items, skill_forks

create table skill_reviews (
  id uuid primary key default gen_random_uuid(),
  skill_id uuid not null references skills(id) on delete cascade,
  user_id uuid not null references users(id) on delete cascade,
  run_id uuid references skill_runs(id) on delete set null,
  rating int not null check (rating between 1 and 5),
  review_text text,
  created_at timestamptz not null default now(),
  unique (skill_id, user_id)
);

create index idx_skill_reviews_skill_id on skill_reviews(skill_id);
create index idx_skill_reviews_user_id on skill_reviews(user_id);

create table saved_skills (
  user_id uuid not null references users(id) on delete cascade,
  skill_id uuid not null references skills(id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (user_id, skill_id)
);

create index idx_saved_skills_user_id on saved_skills(user_id);

create table collections (
  id uuid primary key default gen_random_uuid(),
  owner_id uuid not null references users(id) on delete cascade,
  title text not null,
  description text,
  visibility text not null default 'public', -- public/private
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index idx_collections_owner_id on collections(owner_id);

create table collection_items (
  collection_id uuid not null references collections(id) on delete cascade,
  skill_id uuid not null references skills(id) on delete cascade,
  sort_order int not null default 0,
  added_at timestamptz not null default now(),
  primary key (collection_id, skill_id)
);

create index idx_collection_items_collection_id on collection_items(collection_id);

create table skill_forks (
  id uuid primary key default gen_random_uuid(),
  source_skill_id uuid not null references skills(id) on delete cascade,
  forked_skill_id uuid not null references skills(id) on delete cascade,
  forked_by uuid not null references users(id) on delete cascade,
  created_at timestamptz not null default now()
);
