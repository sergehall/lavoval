import Link from 'next/link';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import { SkillEditorForm } from '@/features/admin/skill-editor-form';
import { createOwnSkillAction } from '@/features/skills/actions';
import { fetchMySkills, withValidSession } from '@/shared/api/server-client';
import { formatDate } from '@/shared/lib/utils';

export default async function MySkillsPage() {
  const { data: skills } = await withValidSession((session) => fetchMySkills(session.accessToken));

  return (
    <div className="stack stack--lg">
      <Card>
        <div className="stack stack--md">
          <div className="section-heading">
            <h2>Create a new skill offer</h2>
            <Badge tone="warning">Private until you publish</Badge>
          </div>
          <SkillEditorForm action={createOwnSkillAction} submitLabel="Create offer draft" />
        </div>
      </Card>
      <Card>
        <div className="data-list">
          {skills.map((skill) => (
            <article key={skill.id} className="data-list__item">
              <div className="section-heading">
                <div className="stack stack--sm">
                  <h2>{skill.title}</h2>
                  <p>{skill.summary}</p>
                </div>
                <Badge tone={skill.status === 'published' ? 'success' : 'warning'}>
                  {skill.status}
                </Badge>
              </div>
              <div className="inline-actions muted">
                <span>Visibility: {skill.visibility}</span>
                <span>{skill.modulesCount} modules</span>
                <span>Updated {formatDate(skill.updatedAt)}</span>
              </div>
              <Link href={`/account/my-skills/${skill.id}`} className="muted">
                Edit offer
              </Link>
            </article>
          ))}
        </div>
      </Card>
    </div>
  );
}
