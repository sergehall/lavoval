import { Card } from '@/shared/ui/card';
import { SkillEditorForm } from '@/features/admin/skill-editor-form';
import { deleteSkillAction, updateSkillAction } from '@/features/admin/actions';
import { fetchAdminSkillById, withValidSession } from '@/shared/api/server-client';

export default async function AdminSkillEditPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const { data: skill } = await withValidSession((session) =>
    fetchAdminSkillById(session.accessToken, id),
  );

  return (
    <div className="stack stack--lg">
      <div className="stack stack--sm">
        <h1>Review skill offer</h1>
        <p className="muted">
          Update publishing state, content framing, and visibility while preserving a consistent
          governance workflow for the marketplace.
        </p>
      </div>
      <Card>
        <SkillEditorForm
          skill={skill}
          action={updateSkillAction.bind(null, skill.id)}
          archiveAction={deleteSkillAction.bind(null, skill.id)}
        />
      </Card>
    </div>
  );
}
