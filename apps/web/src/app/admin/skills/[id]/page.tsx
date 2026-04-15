import { Card } from '@/shared/ui/card';
import { Badge } from '@/shared/ui/badge';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Textarea } from '@/shared/ui/textarea';
import { SkillEditorForm } from '@/features/admin/skill-editor-form';
import { deleteSkillAction, updateSkillAction } from '@/features/admin/actions';
import {
  createModuleAction,
  deleteModuleAction,
  updateModuleAction,
} from '@/features/admin/modules/actions';
import {
  fetchAdminModules,
  fetchAdminSkillById,
  withValidSession,
} from '@/shared/api/server-client';

const STATUS_TONE = {
  draft: 'warning',
  published: 'success',
  archived: 'neutral',
} as const;

export default async function AdminSkillEditPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  const [{ data: skill }, { data: modules }] = await Promise.all([
    withValidSession((session) => fetchAdminSkillById(session.accessToken, id)),
    withValidSession((session) => fetchAdminModules(session.accessToken, id)),
  ]);

  const boundCreate = createModuleAction.bind(null, id);

  return (
    <div className="stack stack--lg">
      {/* ── Skill editor ─────────────────────────────── */}
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

      {/* ── Modules ──────────────────────────────────── */}
      <div className="stack stack--sm">
        <h2>
          Modules{' '}
          <span className="muted" style={{ fontSize: 14, fontWeight: 400 }}>
            ({modules.length})
          </span>
        </h2>
        <p className="muted">Learning units that make up this skill. Ordered by position.</p>
      </div>

      {/* Existing modules */}
      {modules.map((mod) => {
        const boundUpdate = updateModuleAction.bind(null, id, mod.id);
        const boundDelete = deleteModuleAction.bind(null, id, mod.id);
        return (
          <Card key={mod.id}>
            <details className="module-editor">
              <summary className="module-editor__summary">
                <span className="module-editor__pos">#{mod.position}</span>
                <span className="module-editor__title">{mod.title}</span>
                <span className="module-editor__slug muted">{mod.slug}</span>
                <Badge tone={STATUS_TONE[mod.status as keyof typeof STATUS_TONE] ?? 'neutral'}>
                  {mod.status}
                </Badge>
              </summary>
              <form action={boundUpdate} className="stack stack--md module-editor__form">
                <div className="form-grid">
                  <label>
                    <span>Slug</span>
                    <Input
                      name="slug"
                      defaultValue={mod.slug}
                      required
                      minLength={2}
                      maxLength={100}
                      pattern="^[a-z0-9-]+"
                      title="Lowercase letters, digits, and hyphens only"
                    />
                  </label>
                  <label>
                    <span>Position</span>
                    <Input name="position" type="number" defaultValue={mod.position} min={0} />
                  </label>
                  <label className="form-grid__full">
                    <span>Title</span>
                    <Input
                      name="title"
                      defaultValue={mod.title}
                      required
                      minLength={2}
                      maxLength={200}
                    />
                  </label>
                  <label className="form-grid__full">
                    <span>Summary</span>
                    <Input
                      name="summary"
                      defaultValue={mod.summary}
                      required
                      minLength={2}
                      maxLength={500}
                    />
                  </label>
                  <label className="form-grid__full">
                    <span>Content</span>
                    <Textarea
                      name="content"
                      defaultValue={mod.content}
                      rows={8}
                      required
                      minLength={10}
                    />
                  </label>
                  <label>
                    <span>Status</span>
                    <select name="status" defaultValue={mod.status}>
                      <option value="draft">draft</option>
                      <option value="published">published</option>
                      <option value="archived">archived</option>
                    </select>
                  </label>
                </div>
                <div style={{ display: 'flex', gap: 8 }}>
                  <Button type="submit">Save module</Button>
                  <form action={boundDelete}>
                    <button type="submit" className="module-delete-btn">
                      Delete
                    </button>
                  </form>
                </div>
              </form>
            </details>
          </Card>
        );
      })}

      {/* Add new module */}
      <Card>
        <h2 className="card__title">Add module</h2>
        <form action={boundCreate} className="stack stack--md">
          <div className="form-grid">
            <label>
              <span>Slug</span>
              <Input
                name="slug"
                required
                minLength={2}
                maxLength={100}
                pattern="^[a-z0-9-]+"
                title="Lowercase letters, digits, and hyphens only"
                placeholder="intro-to-typescript"
              />
            </label>
            <label>
              <span>
                Position <small>(optional, appends to end if blank)</small>
              </span>
              <Input name="position" type="number" min={0} placeholder="auto" />
            </label>
            <label className="form-grid__full">
              <span>Title</span>
              <Input
                name="title"
                required
                minLength={2}
                maxLength={200}
                placeholder="Introduction to TypeScript"
              />
            </label>
            <label className="form-grid__full">
              <span>Summary</span>
              <Input
                name="summary"
                required
                minLength={2}
                maxLength={500}
                placeholder="A concise overview of what this module covers."
              />
            </label>
            <label className="form-grid__full">
              <span>Content</span>
              <Textarea
                name="content"
                rows={8}
                required
                minLength={10}
                placeholder="Full module content in plain text or markdown."
              />
            </label>
            <label>
              <span>Status</span>
              <select name="status" defaultValue="draft">
                <option value="draft">draft</option>
                <option value="published">published</option>
              </select>
            </label>
          </div>
          <div>
            <Button type="submit">Add module</Button>
          </div>
        </form>
      </Card>
    </div>
  );
}
