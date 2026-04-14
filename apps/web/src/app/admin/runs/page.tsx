import { RunsHistory } from '@/features/runtime/runs-history';
import {
  fetchAdminRuns,
  requireAdminSession,
  withValidSession,
} from '@/shared/api/server-client';

export default async function AdminRunsPage() {
  await requireAdminSession();
  const { data: runs } = await withValidSession((session) => fetchAdminRuns(session.accessToken));

  return (
    <div className="stack stack--lg">
      <div className="stack stack--sm">
        <h1>Runtime observability</h1>
        <p className="muted">
          Review every skill execution across the marketplace, inspect failures, and watch how
          entrypoints are being used in the wild.
        </p>
      </div>
      <RunsHistory runs={runs} detailBasePath="/admin/runs" />
    </div>
  );
}
