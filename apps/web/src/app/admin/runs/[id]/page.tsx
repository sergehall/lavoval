import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import {
  fetchAdminRunById,
  requireAdminSession,
  withValidSession,
} from '@/shared/api/server-client';
import { formatDate } from '@/shared/lib/utils';

function prettyJSON(value: unknown) {
  return JSON.stringify(value ?? {}, null, 2);
}

export default async function AdminRunDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  await requireAdminSession();
  const { data: run } = await withValidSession((session) =>
    fetchAdminRunById(session.accessToken, id),
  );

  return (
    <div className="stack stack--lg">
      <div className="section-heading">
        <div className="stack stack--sm">
          <h1>Runtime run detail</h1>
          <p className="muted">
            Inspect a marketplace-wide run from the governance surface, including payloads and
            execution metadata.
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
            <span>User ID: {run.userId}</span>
          </div>
        </div>
      </Card>

      <Card>
        <div className="inline-actions muted">
          <span>Run ID: {run.id}</span>
          <span>Created {formatDate(run.createdAt)}</span>
          {run.meta.durationMs ? <span>Duration: {run.meta.durationMs} ms</span> : null}
          <span>Input fields: {run.meta.inputKeysCount}</span>
          <span>Output fields: {run.meta.outputKeysCount}</span>
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
    </div>
  );
}
