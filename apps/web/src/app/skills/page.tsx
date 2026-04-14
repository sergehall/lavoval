import { fetchSkills } from '@/shared/api/server-client';
import { SkillsCatalog } from '@/features/skills/skills-catalog';
import { Card } from '@/shared/ui/card';

export default async function SkillsPage() {
  const { data: skills } = await fetchSkills();

  return (
    <div className="stack stack--lg">
      <Card>
        <div className="stack stack--sm">
          <h1>Explore human expertise</h1>
          <p className="muted">
            Browse public skill offers created by other people, discover practical know-how, and
            see how human expertise is being packaged for an AI-shaped world.
          </p>
        </div>
      </Card>
      <SkillsCatalog skills={skills} />
    </div>
  );
}
