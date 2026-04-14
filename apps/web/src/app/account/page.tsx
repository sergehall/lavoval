import Link from 'next/link';
import { Card } from '@/shared/ui/card';
import { Badge } from '@/shared/ui/badge';
import {
  fetchMySkills,
  fetchProfile,
  fetchSkills,
  withValidSession,
} from '@/shared/api/server-client';

export default async function AccountDashboardPage() {
  const { session, profile, skills, mySkills } = await withValidSession(async (session) => {
    const [{ data: profile }, { data: skills }, { data: mySkills }] = await Promise.all([
      fetchProfile(session.accessToken),
      fetchSkills(session.accessToken),
      fetchMySkills(session.accessToken),
    ]);

    return { session, profile, skills, mySkills };
  });

  return (
    <div className="stack stack--lg">
      <section className="section-heading">
        <div className="stack stack--sm">
          <h1>Welcome back, {profile.firstName}</h1>
          <p className="muted">
            Track profile health, available skills, and future learning activity from one place.
          </p>
        </div>
        <Badge tone="success">Role: {session.user.role}</Badge>
      </section>
      <section className="grid">
        <Card>
          <div className="stack stack--sm">
            <h2>Profile readiness</h2>
            <p>
              {profile.bio ||
                'Add a short bio to support ownership, permissions, and future collaboration context.'}
            </p>
            <Link href="/account/profile" className="muted">
              Edit profile
            </Link>
          </div>
        </Card>
        <Card>
          <div className="stack stack--sm">
            <h2>My draft workflow</h2>
            <p>
              {mySkills.length} skill records currently belong to your account, including drafts you
              can publish later.
            </p>
            <Link href="/account/my-skills" className="muted">
              Manage my skills
            </Link>
          </div>
        </Card>
        <Card>
          <div className="stack stack--sm">
            <h2>Public catalog</h2>
            <p>
              {skills.length} published skill records are currently visible in the public catalog.
            </p>
            <Link href="/skills" className="muted">
              Browse published skills
            </Link>
          </div>
        </Card>
      </section>
    </div>
  );
}
