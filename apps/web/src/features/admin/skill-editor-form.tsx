import type { SkillDetail } from '@lavoval/contracts';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Textarea } from '@/shared/ui/textarea';

type SkillEditorFormProps = {
  skill?: SkillDetail;
  action: (formData: FormData) => void | Promise<void>;
  archiveAction?: (formData: FormData) => void | Promise<void>;
  submitLabel?: string;
  archiveLabel?: string;
};

function FieldLabel({ label, hint }: { label: string; hint: string }) {
  return (
    <span className="field-label">
      <span>{label}</span>
      <span className="field-hint" tabIndex={0} aria-label={`${label} help`}>
        ?<span className="field-hint__tooltip">{hint}</span>
      </span>
    </span>
  );
}

export function SkillEditorForm({
  skill,
  action,
  archiveAction,
  submitLabel,
  archiveLabel = 'Archive skill',
}: SkillEditorFormProps) {
  return (
    <form action={action} className="stack stack--md form-grid">
      <label>
        <FieldLabel
          label="Slug"
          hint="Defines the URL key for this skill. Expected format: lowercase English words separated with hyphens, for example `prompt-engineering-basics`. Avoid spaces, uppercase characters, and special symbols."
        />
        <Input
          name="slug"
          defaultValue={skill?.slug}
          placeholder="prompt-engineering-basics"
          required
        />
      </label>
      <label>
        <FieldLabel
          label="Title"
          hint="Defines the display name shown in the catalog, detail page, and account views. Keep it short, clear, and readable as a product-facing title."
        />
        <Input
          name="title"
          defaultValue={skill?.title}
          placeholder="Prompt Engineering Basics"
          required
        />
      </label>
      <label className="form-grid__full">
        <FieldLabel
          label="Summary"
          hint="Defines the short preview used in cards and lists. Recommended format: 1-2 sentences covering what the skill teaches, who it is for, and the value it gives."
        />
        <Textarea
          name="summary"
          defaultValue={skill?.summary}
          placeholder="A short 1-2 sentence description that explains what this skill teaches and who it is for."
          required
          rows={3}
        />
      </label>
      <label className="form-grid__full">
        <FieldLabel
          label="Description"
          hint="Defines the main skill overview in markdown. Use it for goals, audience, outcomes, structure, references, and any content that should appear on the full skill page."
        />
        <Textarea
          name="description"
          defaultValue={skill?.description}
          placeholder={`# What you will learn

This skill helps users understand...

## Who it is for

- Beginners who need...
- Teams who want...

## Expected outcome

After completing this skill, the user will be able to...`}
          required
          rows={10}
        />
        <span className="muted">
          Use markdown here. This is the full skill overview page: goals, audience, outcomes,
          structure, and any key notes.
        </span>
      </label>
      <label>
        <FieldLabel
          label="Status"
          hint="Controls lifecycle state. `Draft` keeps the record in progress, `Published` makes it usable in the product, and `Archived` removes it from active circulation."
        />
        <select name="status" defaultValue={skill?.status ?? 'draft'} className="input">
          <option value="draft">Draft</option>
          <option value="published">Published</option>
          <option value="archived">Archived</option>
        </select>
      </label>
      <label>
        <FieldLabel
          label="Visibility"
          hint="Controls audience scope. `Private` keeps the skill inside account and admin workflows. `Public` allows it to appear in the public catalog once ready."
        />
        <select name="visibility" defaultValue={skill?.visibility ?? 'private'} className="input">
          <option value="private">Private</option>
          <option value="public">Public</option>
        </select>
      </label>
      <div className="form-grid__full toolbar">
        <Button type="submit">{submitLabel ?? (skill ? 'Save skill' : 'Create skill')}</Button>
        {skill && archiveAction ? (
          <button className="button button--danger" formAction={archiveAction}>
            {archiveLabel}
          </button>
        ) : null}
      </div>
    </form>
  );
}
