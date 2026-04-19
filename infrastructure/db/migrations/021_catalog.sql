-- categories, subcategories, tags, skill_tag_links

create table categories (
  id uuid primary key default gen_random_uuid(),
  slug text not null unique,
  name text not null,
  description text,
  icon text,
  sort_order int not null default 0,
  created_at timestamptz not null default now()
);

create table subcategories (
  id uuid primary key default gen_random_uuid(),
  category_id uuid not null references categories(id) on delete cascade,
  slug text not null unique,
  name text not null,
  sort_order int not null default 0,
  created_at timestamptz not null default now()
);

create table tags (
  id uuid primary key default gen_random_uuid(),
  slug text not null unique,
  name text not null,
  kind text not null default 'topic', -- topic/tool/language/industry/agent_capability
  created_at timestamptz not null default now()
);

create table skill_tag_links (
  skill_id uuid not null references skills(id) on delete cascade,
  tag_id uuid not null references tags(id) on delete cascade,
  primary key (skill_id, tag_id)
);

create index idx_skill_tag_links_skill_id on skill_tag_links(skill_id);
create index idx_skill_tag_links_tag_id on skill_tag_links(tag_id);
create index idx_subcategories_category_id on subcategories(category_id);

-- seed default categories
insert into categories (slug, name, sort_order) values
  ('programming',   'Programming',  1),
  ('design',        'Design',       2),
  ('marketing',     'Marketing',    3),
  ('writing',       'Writing',      4),
  ('research',      'Research',     5),
  ('product',       'Product',      6),
  ('automation',    'Automation',   7),
  ('data',          'Data',         8);

-- seed subcategories for programming
insert into subcategories (category_id, slug, name, sort_order)
select id, s.slug, s.name, s.sort_order
from categories, (values
  ('frontend',       'Frontend',       1),
  ('backend',        'Backend',        2),
  ('devops',         'DevOps',         3),
  ('testing',        'Testing',        4),
  ('mobile',         'Mobile',         5),
  ('ai-engineering', 'AI Engineering', 6),
  ('system-design',  'System Design',  7)
) as s(slug, name, sort_order)
where categories.slug = 'programming';
