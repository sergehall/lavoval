import { RunsHistory } from '@/features/runtime/runs-history';
import { fetchAdminRuns, requireAdminSession, withValidSession } from '@/shared/api/server-client';

export default async function AdminRunsPage() {
  await requireAdminSession();
  const { data: runs } = await withValidSession((session) => fetchAdminRuns(session.accessToken));

  return (
    <div className="stack stack--lg">
      <RunsHistory runs={runs} detailBasePath="/admin/runs" />
    </div>
  );
}
