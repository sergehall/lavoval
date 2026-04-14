import Link from 'next/link';
import type { SkillRun } from '@lavoval/contracts';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import { formatDate } from '@/shared/lib/utils';

function toneForStatus(status: SkillRun['status']) {
  switch (status) {
    case 'completed':
      return 'success';
    case 'failed':
      return 'warning';
    case 'running':
      return 'warning';
    default:
      return 'neutral';
  }
}

export function RunsHistory({ runs }: { runs: SkillRun[] }) {
  return (
    <Card>
      <div className="data-list">
        {runs.length > 0 ? (
          runs.map((run) => (
            <article key={run.id} className="data-list__item">
              <div className="section-heading">
                <div className="stack stack--sm">
                  <h2>Run {run.id.slice(0, 8)}</h2>
                  <p className="muted">Skill ID: {run.skillId}</p>
                </div>
                <Badge tone={toneForStatus(run.status)}>{run.status}</Badge>
              </div>
              <div className="inline-actions muted">
                <span>Created {formatDate(run.createdAt)}</span>
                <span>{Object.keys(run.input ?? {}).length} input fields</span>
                <span>{Object.keys(run.output ?? {}).length} output fields</span>
              </div>
              <Link href={`/account/runs/${run.id}`} className="muted">
                Open run details
              </Link>
            </article>
          ))
        ) : (
          <div className="empty-state stack stack--md">
            <Badge tone="warning">No runs yet</Badge>
            <h2>You have not executed any skills yet.</h2>
            <p className="muted">
              Open a public skill, run it from the detail page, and your history will start showing
              up here.
            </p>
          </div>
        )}
      </div>
    </Card>
  );
}
