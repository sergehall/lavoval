import Link from 'next/link';
import { Card } from '@/shared/ui/card';
import {
  fetchAdminSkills,
  fetchAdminUsers,
  requireAdminSession,
  withValidSession,
} from '@/shared/api/server-client';

export default async function AdminDashboardPage() {
  const session = await requireAdminSession();
  const { users, skills } = await withValidSession(async (activeSession) => {
    const [{ data: users }, { data: skills }] = await Promise.all([
      fetchAdminUsers(activeSession.accessToken),
      fetchAdminSkills(activeSession.accessToken),
    ]);

    return { users, skills };
  });

  return (
    <div className="stack stack--lg">
      <div className="stack stack--sm">
        <h1>Admin dashboard</h1>
        <p className="muted">
          Manage users, monitor catalog growth, and prepare the system for richer governance
          controls.
        </p>
      </div>
      <section className="grid">
        <Card>
          <h2>{users.length}</h2>
          <p>Registered users ready for future permissions matrix and moderation workflows.</p>
        </Card>
        <Card>
          <h2>{skills.length}</h2>
          <p>
            Skill records under operational management with soft delete and status lifecycle
            support.
          </p>
        </Card>
        <Card>
          <h2>Roadmap ready</h2>
          <p>
            Audit logs, analytics, approvals, and notifications can be layered on the same admin
            surface.
          </p>
        </Card>
      </section>
      <div className="inline-actions">
        <Link href="/admin/users" className="muted">
          Review users
        </Link>
        <Link href="/admin/skills" className="muted">
          Manage skills
        </Link>
      </div>
    </div>
  );
}
