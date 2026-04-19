import type { Metadata } from 'next';
import { fetchSkills, fetchCategories, fetchTags } from '@/shared/api/server-client';
import { SkillsCatalog } from '@/features/skills/skills-catalog';
import { Card } from '@/shared/ui/card';

export const dynamic = 'force-dynamic';

export const metadata: Metadata = {
  title: 'Browse Skills — Lavoval',
  description:
    'Explore public skills on Lavoval. Discover reusable expertise packaged for people and AI agents.',
  alternates: { canonical: '/skills' },
  openGraph: {
    title: 'Lavoval Skills | Browse Expertise',
    description: 'Explore public skills on Lavoval and discover practical expertise.',
    url: '/skills',
  },
};

function getParam(params: Record<string, string | string[] | undefined>, key: string): string {
  const v = params[key];
  return Array.isArray(v) ? (v[0] ?? '') : (v ?? '');
}

export default async function SkillsPage({
  searchParams,
}: {
  searchParams?: Promise<Record<string, string | string[] | undefined>>;
}) {
  const params = searchParams ? await searchParams : {};

  const difficulty = getParam(params, 'difficulty');
  const sort = getParam(params, 'sort');

  const filter = {
    q: getParam(params, 'q') || undefined,
    category: getParam(params, 'category') || undefined,
    subcategory: getParam(params, 'subcategory') || undefined,
    tags: getParam(params, 'tags') || undefined,
    difficulty: (difficulty || undefined) as 'junior' | 'middle' | 'senior' | undefined,
    skillType: (getParam(params, 'skillType') || undefined) as
      | 'workflow'
      | 'guide'
      | 'prompt-pack'
      | 'agent-ready'
      | 'service'
      | undefined,
    sort: (sort || undefined) as 'popular' | 'new' | 'rating' | 'runs' | undefined,
    agentReady: params.agentReady === 'true' ? true : undefined,
  };

  const [skillsResult, categories, tags] = await Promise.all([
    fetchSkills(filter),
    fetchCategories(),
    fetchTags(),
  ]);
  const skills = skillsResult?.data ?? [];

  return (
    <div className="stack stack--skills-page">
      <Card className="skills-page__hero">
        <div className="stack stack--sm">
          <h1>Explore skills</h1>
          <p className="muted">
            Browse published skills created by the community. Discover reusable expertise ready to
            read, apply, or run through an AI agent.
          </p>
        </div>
      </Card>
      <SkillsCatalog
        skills={skills}
        categories={categories ?? []}
        tags={tags ?? []}
        activeFilter={filter}
      />
    </div>
  );
}
