import { Button } from '@/shared/ui/button';
import { Textarea } from '@/shared/ui/textarea';

export function RunSkillForm({
  action,
  entrypoint,
}: {
  action: (formData: FormData) => void | Promise<void>;
  entrypoint: string;
}) {
  return (
    <form action={action} className="stack stack--md">
      <div className="stack stack--sm">
        <h2>Run this skill</h2>
        <p className="muted">
          This skill uses the <strong>{entrypoint}</strong> runtime entrypoint. For the current v2
          milestone, mock executors mostly work with a single text input.
        </p>
      </div>
      <label>
        <span>Input text</span>
        <Textarea
          name="text"
          rows={6}
          placeholder="Paste a prompt, paragraph, notes, or source text to run through this skill."
        />
      </label>
      <Button type="submit">Run skill</Button>
    </form>
  );
}
