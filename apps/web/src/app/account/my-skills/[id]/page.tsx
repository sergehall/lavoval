import { Card } from '@/shared/ui/card';
import { SkillEditorForm } from '@/features/admin/skill-editor-form';
import { deleteOwnSkillAction, updateOwnSkillAction } from '@/features/skills/actions';
import {
  fetchCategories,
  fetchMySkillById,
  fetchTags,
  withValidSession,
} from '@/shared/api/server-client';

export default async function MySkillEditPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  const [{ data: skill }, categories, tags] = await Promise.all([
    withValidSession((session) => fetchMySkillById(session.accessToken, id)),
    fetchCategories(),
    fetchTags(),
  ]);

  return (
    <div className="stack stack--lg">
      <div className="stack stack--sm">
        <h1>Edit my skill offer</h1>
        <p className="muted">
          Shape how your expertise appears in the marketplace, keep it private while refining it,
          and publish when it is ready for discovery.
        </p>
      </div>
      <Card>
        <SkillEditorForm
          skill={skill}
          categories={categories ?? []}
          tags={tags ?? []}
          action={updateOwnSkillAction.bind(null, skill.id)}
          archiveAction={deleteOwnSkillAction.bind(null, skill.id)}
          submitLabel="Save offer"
          archiveLabel="Archive offer"
        />
      </Card>
    </div>
  );
}
