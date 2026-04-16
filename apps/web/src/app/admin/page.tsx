import type { Route } from 'next';
import Link from 'next/link';
import type { AdminStats } from '@lavoval/contracts';
import type { ApiEnvelope, MailOperationalSnapshot } from '@/shared/api/types';
import { Card } from '@/shared/ui/card';
import {
  fetchAdminMailOperations,
  fetchAdminRuns,
  fetchAdminStats,
  withValidSession,
} from '@/shared/api/server-client';

function unwrapData<T>(payload: T | ApiEnvelope<T>): T {
  if (typeof payload === 'object' && payload !== null && 'data' in payload) {
    return payload.data;
  }

  return payload;
}

export default async function AdminDashboardPage() {
  const { stats, runs, mailOps } = await withValidSession(async (activeSession) => {
    const [statsPayload, { data: runs }, mailOpsPayload] = await Promise.all([
      fetchAdminStats(activeSession.accessToken),
      fetchAdminRuns(activeSession.accessToken),
      fetchAdminMailOperations(activeSession.accessToken),
    ]);

    return {
      stats: unwrapData<AdminStats>(statsPayload),
      runs,
      mailOps: unwrapData<MailOperationalSnapshot>(mailOpsPayload),
    };
  });

  return (
    <div className="stack stack--lg">
      {/* ── People ─────────────────────────────────── */}
      <section className="grid">
        <Card>
          <h2>{stats?.users.total ?? '—'}</h2>
          <p>Total registered users</p>
        </Card>
        <Card>
          <h2>{stats?.users.active ?? '—'}</h2>
          <p>Active accounts</p>
        </Card>
        <Card>
          <h2>{stats?.users.suspended ?? '—'}</h2>
          <p>Suspended accounts</p>
        </Card>
        <Card>
          <h2>{stats?.users.blocked ?? '—'}</h2>
          <p>Blocked accounts</p>
        </Card>
        <Card>
          <h2>{stats?.users.new7d ?? '—'}</h2>
          <p>New users in the last 7 days</p>
        </Card>
        <Card>
          <h2>{stats?.users.new30d ?? '—'}</h2>
          <p>New users in the last 30 days</p>
        </Card>
      </section>

      {/* ── Skill offers ───────────────────────────── */}
      <section className="grid">
        <Card>
          <h2>{stats?.skills.published ?? '—'}</h2>
          <p>Published skills live on the marketplace</p>
        </Card>
        <Card>
          <h2>{stats?.skills.pendingReview ?? '—'}</h2>
          <p>Skills awaiting moderation review</p>
        </Card>
        <Card>
          <h2>{stats?.skills.paid ?? '—'}</h2>
          <p>Paid skills with a price set</p>
        </Card>
        <Card>
          <h2>{mailOps?.countsByStatus.dead_letter ?? '—'}</h2>
          <p>Dead-letter mail jobs waiting for operator review</p>
        </Card>
        <Card>
          <h2>{runs.length}</h2>
          <p>Runtime executions recorded across the marketplace</p>
        </Card>
        <Card>
          <h2>{stats?.skills.new7d ?? '—'}</h2>
          <p>New skills created in the last 7 days</p>
        </Card>
      </section>

      <div className="inline-actions">
        <Link href="/admin/users" className="muted">
          Review people
        </Link>
        <Link href="/admin/skills" className="muted">
          Govern offers
        </Link>
        <Link href="/admin/runs" className="muted">
          Observe runtime
        </Link>
        <Link href={'/admin/mail' as Route} className="muted">
          Operate mail
        </Link>
      </div>
    </div>
  );
}
