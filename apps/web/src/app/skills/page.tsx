import { fetchSkills } from '@/shared/api/server-client';
import { SkillsCatalog } from '@/features/skills/skills-catalog';

export default async function SkillsPage() {
  const { data: skills } = await fetchSkills();

  return <SkillsCatalog skills={skills} />;
}
