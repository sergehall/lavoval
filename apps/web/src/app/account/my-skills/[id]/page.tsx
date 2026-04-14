import { Card } from '@/shared/ui/card';
import { SkillEditorForm } from '@/features/admin/skill-editor-form';
import { deleteOwnSkillAction, updateOwnSkillAction } from '@/features/skills/actions';
import { fetchMySkillById, withValidSession } from '@/shared/api/server-client';

export default async function MySkillEditPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const { data: skill } = await withValidSession((session) =>
    fetchMySkillById(session.accessToken, id),
  );

  return (
    <div className="stack stack--lg">
      <div className="stack stack--sm">
        <h1>Edit My Skill</h1>
        <p className="muted">
          Keep drafts private, switch to published when ready, and archive when the record is no
          longer active.
        </p>
      </div>
      <Card>
        <SkillEditorForm
          skill={skill}
          action={updateOwnSkillAction.bind(null, skill.id)}
          archiveAction={deleteOwnSkillAction.bind(null, skill.id)}
          submitLabel="Save changes"
          archiveLabel="Archive draft"
        />
      </Card>
    </div>
  );
}
