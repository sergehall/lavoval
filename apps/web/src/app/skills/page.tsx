import type { Metadata } from 'next';
import { fetchSkills } from '@/shared/api/server-client';
import { SkillsCatalog } from '@/features/skills/skills-catalog';
import { Card } from '@/shared/ui/card';

export const metadata: Metadata = {
  title: 'Browse Skill Offers And Human Expertise',
  description:
    'Explore public skill offers on Lavoval, compare human expertise, and discover practical know-how packaged into reusable modules for the AI era.',
  alternates: {
    canonical: '/skills',
  },
  openGraph: {
    title: 'Lavoval Skills | Browse Human Expertise',
    description:
      'Explore public skill offers on Lavoval and discover practical human expertise packaged for an AI-shaped world.',
    url: '/skills',
  },
};

export default async function SkillsPage({
  searchParams,
}: {
  searchParams?: Promise<Record<string, string | string[] | undefined>>;
}) {
  const { data: skills } = await fetchSkills();
  const params = searchParams ? await searchParams : {};
  const initialQuery = Array.isArray(params.q) ? (params.q[0] ?? '') : (params.q ?? '');

  return (
    <div className="stack stack--lg">
      <Card>
        <div className="stack stack--sm">
          <h1>Explore human expertise</h1>
          <p className="muted">
            Browse public skill offers created by other people, discover practical know-how, and see
            how human expertise is being packaged for an AI-shaped world.
          </p>
        </div>
      </Card>
      <SkillsCatalog skills={skills} initialQuery={initialQuery} />
    </div>
  );
}
