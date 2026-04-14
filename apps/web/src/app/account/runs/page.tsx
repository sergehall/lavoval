import { fetchSkillRuns, withValidSession } from '@/shared/api/server-client';
import { RunsHistory } from '@/features/runtime/runs-history';

export default async function RunsPage() {
  const { data: runs } = await withValidSession((session) => fetchSkillRuns(session.accessToken));

  return (
    <div className="stack stack--lg">
      <div className="stack stack--sm">
        <h1>Your run history</h1>
        <p className="muted">
          Every skill execution is stored here so you can inspect what you ran, when it happened,
          and what came back from the runtime.
        </p>
      </div>
      <RunsHistory runs={runs} />
    </div>
  );
}
