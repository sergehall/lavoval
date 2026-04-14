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
            Manage your marketplace identity, the skills you offer, and the expertise you want to
            discover from other people.
          </p>
        </div>
        <Badge tone="success">Role: {session.user.role}</Badge>
      </section>
      <section className="grid">
        <Card>
          <div className="stack stack--sm">
            <h2>Marketplace identity</h2>
            <p>
              {profile.bio ||
                'Add a short bio so people can understand your background, perspective, and why your skills are worth exploring.'}
            </p>
            <Link href="/account/profile" className="muted">
              Refine identity
            </Link>
          </div>
        </Card>
        <Card>
          <div className="stack stack--sm">
            <h2>My skill offers</h2>
            <p>
              {mySkills.length} skill offers currently belong to your account, including private
              drafts you can shape before publishing them to the marketplace.
            </p>
            <Link href="/account/my-skills" className="muted">
              Manage my offers
            </Link>
          </div>
        </Card>
        <Card>
          <div className="stack stack--sm">
            <h2>Marketplace pulse</h2>
            <p>
              {skills.length} published skill offers are currently visible in the public
              marketplace.
            </p>
            <Link href="/skills" className="muted">
              Explore live offers
            </Link>
          </div>
        </Card>
      </section>
    </div>
  );
}
