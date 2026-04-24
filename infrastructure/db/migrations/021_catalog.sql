-- lavoval_categories, lavoval_subcategories, lavoval_tags, lavoval_skill_tag_links

create table lavoval_categories (
  id uuid primary key default gen_random_uuid(),
  slug text not null unique,
  name text not null,
  description text,
  icon text,
  sort_order int not null default 0,
  created_at timestamptz not null default now()
);

create table lavoval_subcategories (
  id uuid primary key default gen_random_uuid(),
  category_id uuid not null references lavoval_categories(id) on delete cascade,
  slug text not null unique,
  name text not null,
  sort_order int not null default 0,
  created_at timestamptz not null default now()
);

create table lavoval_tags (
  id uuid primary key default gen_random_uuid(),
  slug text not null unique,
  name text not null,
  kind text not null default 'topic', -- topic/tool/language/industry/agent_capability
  created_at timestamptz not null default now()
);

create table lavoval_skill_tag_links (
  skill_id uuid not null references lavoval_skills(id) on delete cascade,
  tag_id uuid not null references lavoval_tags(id) on delete cascade,
  primary key (skill_id, tag_id)
);

create index lavoval_idx_skill_tag_links_skill_id on lavoval_skill_tag_links(skill_id);
create index lavoval_idx_skill_tag_links_tag_id on lavoval_skill_tag_links(tag_id);
create index lavoval_idx_subcategories_category_id on lavoval_subcategories(category_id);

-- seed default lavoval_categories
insert into lavoval_categories (slug, name, sort_order) values
  ('programming',   'Programming',  1),
  ('design',        'Design',       2),
  ('marketing',     'Marketing',    3),
  ('writing',       'Writing',      4),
  ('research',      'Research',     5),
  ('product',       'Product',      6),
  ('automation',    'Automation',   7),
  ('data',          'Data',         8);

-- seed lavoval_subcategories for programming
insert into lavoval_subcategories (category_id, slug, name, sort_order)
select id, s.slug, s.name, s.sort_order
from lavoval_categories, (values
  ('frontend',       'Frontend',       1),
  ('backend',        'Backend',        2),
  ('devops',         'DevOps',         3),
  ('testing',        'Testing',        4),
  ('mobile',         'Mobile',         5),
  ('ai-engineering', 'AI Engineering', 6),
  ('system-design',  'System Design',  7)
) as s(slug, name, sort_order)
where lavoval_categories.slug = 'programming';
