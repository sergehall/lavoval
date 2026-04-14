import { getRuntimePreset } from '@lavoval/engine';
import { Button } from '@/shared/ui/button';
import { Textarea } from '@/shared/ui/textarea';

export function RunSkillForm({
  action,
  entrypoint,
}: {
  action: (formData: FormData) => void | Promise<void>;
  entrypoint: string;
}) {
  const preset = getRuntimePreset(entrypoint);

  return (
    <form action={action} className="stack stack--md">
      <div className="stack stack--sm">
        <h2>Run this skill</h2>
        <p className="muted">
          This skill uses the <strong>{preset.entrypoint}</strong> runtime entrypoint.{' '}
          {preset.description}
        </p>
        <BadgeLike>{preset.outputHint}</BadgeLike>
      </div>
      <label>
        <span>{preset.inputLabel}</span>
        <Textarea
          name="text"
          rows={6}
          placeholder={preset.inputPlaceholder}
        />
      </label>
      <Button type="submit">Run skill</Button>
    </form>
  );
}

function BadgeLike({ children }: { children: string }) {
  return <span className="badge badge--neutral">{children}</span>;
}
