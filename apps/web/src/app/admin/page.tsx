import Link from 'next/link';
import { Card } from '@/shared/ui/card';
import {
  fetchAdminSkills,
  fetchAdminRuns,
  fetchAdminUsers,
  requireAdminSession,
  withValidSession,
} from '@/shared/api/server-client';

export default async function AdminDashboardPage() {
  const session = await requireAdminSession();
  const { users, skills, runs } = await withValidSession(async (activeSession) => {
    const [{ data: users }, { data: skills }, { data: runs }] = await Promise.all([
      fetchAdminUsers(activeSession.accessToken),
      fetchAdminSkills(activeSession.accessToken),
      fetchAdminRuns(activeSession.accessToken),
    ]);

    return { users, skills, runs };
  });

  return (
    <div className="stack stack--lg">
      <div className="stack stack--sm">
        <h1>Marketplace governance</h1>
        <p className="muted">
          Steward trust across the exchange: monitor people, oversee published offers, and keep the
          marketplace healthy as it grows.
        </p>
      </div>
      <section className="grid">
        <Card>
          <h2>{users.length}</h2>
          <p>Registered people participating in the exchange and ready for richer trust controls.</p>
        </Card>
        <Card>
          <h2>{skills.length}</h2>
          <p>
            Skill offers under governance with lifecycle states, visibility controls, and archive
            support.
          </p>
        </Card>
        <Card>
          <h2>{runs.length}</h2>
          <p>
            Runtime executions recorded across the marketplace, ready for observability, failure
            triage, and future audit trails.
          </p>
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
      </div>
    </div>
  );
}
