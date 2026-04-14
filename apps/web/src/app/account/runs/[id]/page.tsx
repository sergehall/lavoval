import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import { fetchSkillRunById, withValidSession } from '@/shared/api/server-client';
import { formatDate } from '@/shared/lib/utils';

function prettyJSON(value: unknown) {
  return JSON.stringify(value ?? {}, null, 2);
}

export default async function RunDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const { data: run } = await withValidSession((session) =>
    fetchSkillRunById(session.accessToken, id),
  );

  return (
    <div className="stack stack--lg">
      <div className="section-heading">
        <div className="stack stack--sm">
          <h1>Run details</h1>
          <p className="muted">
            Inspect the stored input, output, and lifecycle timestamps for this execution.
          </p>
        </div>
        <Badge>{run.status}</Badge>
      </div>

      <Card>
        <div className="stack stack--sm">
          <h2>{run.skill.title}</h2>
          <div className="inline-actions muted">
            <span>Slug: {run.skill.slug}</span>
            <span>Entrypoint: {run.skill.entrypoint}</span>
          </div>
        </div>
      </Card>

      <Card>
        <div className="inline-actions muted">
          <span>Run ID: {run.id}</span>
          <span>Skill ID: {run.skillId}</span>
          <span>Created {formatDate(run.createdAt)}</span>
          {run.meta.durationMs ? <span>Duration: {run.meta.durationMs} ms</span> : null}
        </div>
      </Card>

      <Card>
        <div className="stack stack--sm">
          <h2>Input</h2>
          <pre>{prettyJSON(run.input)}</pre>
        </div>
      </Card>

      <Card>
        <div className="stack stack--sm">
          <h2>Output</h2>
          <pre>{prettyJSON(run.output)}</pre>
        </div>
      </Card>

      {run.errorMessage ? (
        <Card>
          <div className="stack stack--sm">
            <h2>Error</h2>
            <p>{run.errorMessage}</p>
          </div>
        </Card>
      ) : null}

      <Card>
        <div className="inline-actions muted">
          <span>Started: {run.startedAt ? formatDate(run.startedAt) : 'not recorded'}</span>
          <span>Finished: {run.finishedAt ? formatDate(run.finishedAt) : 'not recorded'}</span>
        </div>
      </Card>
    </div>
  );
}
