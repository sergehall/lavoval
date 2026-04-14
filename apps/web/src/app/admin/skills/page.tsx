import Link from 'next/link';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import { createSkillAction } from '@/features/admin/actions';
import { SkillEditorForm } from '@/features/admin/skill-editor-form';
import { fetchAdminSkills, withValidSession } from '@/shared/api/server-client';

export default async function AdminSkillsPage() {
  const { data: skills } = await withValidSession((session) =>
    fetchAdminSkills(session.accessToken),
  );

  return (
    <div className="stack stack--lg">
      <div className="section-heading">
        <div className="stack stack--sm">
          <h1>Skill offer governance</h1>
          <p className="muted">
            Create and steward public offers while preserving room for moderation, ranking,
            pagination, and marketplace analytics.
          </p>
        </div>
      </div>
      <Card>
        <div className="stack stack--md">
          <h2>Create a managed offer</h2>
          <SkillEditorForm action={createSkillAction} submitLabel="Create managed offer" />
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
              <div className="inline-actions">
                <span className="muted">Visibility: {skill.visibility}</span>
                <Link href={`/admin/skills/${skill.id}`} className="muted">
                  Review offer
                </Link>
              </div>
            </article>
          ))}
        </div>
      </Card>
    </div>
  );
}
